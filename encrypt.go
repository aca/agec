package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
)

type SecretFile struct {
	Input         string
	RelFromRoot   string
	RelFromWD     string
	ABS           string
	RealPath      string
	DecryptedPath string
}

type encryptOpt struct {
	Users   []string
	Groups  []string
	FileArg SecretFile
	Force   bool

	ctx *Context
}

func runEncrypt(opts *encryptOpt) (rerr error) {
	ctx := opts.ctx

	f := opts.FileArg
	encryptedFile := f.Input + ".age"

	if fileExists(encryptedFile) && !opts.Force {
		log.Printf("skipping %q. Use --force to re-encrypt file", f)
		return nil
	}

	var recs []age.Recipient

	secret, err := ctx.Config.GetSecret(f.RelFromRoot)
	if err != nil {
		// new secret
		secret = Secret{
			Users:  opts.Users,
			Groups: opts.Groups,
			Path:   f.RelFromRoot,
		}

		recs, err = ctx.Config.GetRecipients(opts.Users, opts.Groups)
		if err != nil {
			return err
		}
	} else {
		// existing secret
		// NOTES: Add message to run chown first to chown for secret
		// Fail if user,group has changed

		if len(opts.Users) != 0 || len(opts.Groups) != 0 {
			if !cmp.Equal(opts.Users, secret.Users) {
				return fmt.Errorf("owner of secret has changed, `agec chown -u %v %v` first", strings.Join(opts.Users, ","), f.Input)
			}
			if !cmp.Equal(opts.Groups, secret.Groups) {
				return fmt.Errorf("owner of secret has changed, `agec chown -g %v %v` first", strings.Join(opts.Groups, ","), f.Input)
			}
		}

		recs, err = ctx.Config.GetRecipients(secret.Users, secret.Groups)
		if err != nil {
			return err
		}
	}

	out := &bytes.Buffer{}
	agew, err := age.Encrypt(out, recs...)
	if err != nil {
		return err
	}
	defer agew.Close()

	plainFile, err := os.Open(f.ABS)
	if err != nil {
		return err
	}

	_, err = io.Copy(agew, plainFile)
	if err != nil {
		return err
	}

	if agew.Close() != nil {
		return err
	}

	if plainFile.Close() != nil {
		return err
	}

	err = WriteFile(encryptedFile, out.Bytes())
	if err != nil {
		return err
	}

	fmt.Printf("encrypted %q\n", encryptedFile)

	err = os.MkdirAll(filepath.Dir(f.DecryptedPath), 0o777)
	if err != nil {
		return err
	}

	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Rename(f.RealPath, f.DecryptedPath)
	if err != nil {
		return err
	}

	relativelinkpath, err := filepath.Rel(filepath.Dir(f.ABS), f.DecryptedPath)
	if err != nil {
		return err
	}

	err = os.Remove(f.Input)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	err = os.Symlink(relativelinkpath, f.Input)
	if err != nil {
		return err
	}

	ctx.Config.SaveSecret(secret)
	return ctx.WriteConfigFile()
}

func newEncryptCmd(ctx *Context) *cobra.Command {
	opts := &encryptOpt{
		ctx: ctx,
	}

	cmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt and track secrets",
		Example: `  # john, merry can decrypt secret
  agec encrypt -u john,merry secret.txt

  # members of "admin" group or john can decrypt secret
  agec encrypt -g admin -u john secret.txt

  # set default user/group arguments by setting environment variable
  # AGEC_USER, AGRC_GROUP.
  AGEC_GROUP=admin,devops agec encrypt secret.txt

  # re-encrypt all secrets under current directory
  agec ls | xargs agec encrypt --force`,

		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if ctx == nil {
				return ErrConfigurationNotFound
			}

			if opts.Users == nil {
				opts.Users = ctx.DefaultUsers
			}

			if opts.Groups == nil {
				opts.Groups = ctx.DefaultGroups
			}

			var err error
			opts.FileArg, err = sanitizeEncryptArg(ctx, args[0], ctx.WorkingDir)
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEncrypt(opts)
		},
	}

	f := cmd.Flags()
	f.StringSliceVarP(&opts.Users, "user", "u", nil, "List of users")
	f.StringSliceVarP(&opts.Groups, "group", "g", nil, "List of groups")
	f.BoolVarP(&opts.Force, "force", "f", false, "Overwrite an existing decrypted file")

	cmd.RegisterFlagCompletionFunc("user", UserCompletion(ctx))
	cmd.RegisterFlagCompletionFunc("group", GroupCompletion(ctx))

	return cmd
}

func sanitizeEncryptArg(ctx *Context, arg string, workingDir string) (SecretFile, error) {
	tf := SecretFile{
		Input: arg,
	}

	fileInfo, err := os.Lstat(tf.Input)
	if err != nil {
		return tf, err
	}

	if fileInfo.Mode().IsDir() {
		return tf, fmt.Errorf("%q is not a regular file", arg)
	}

	if fileInfo.Mode()&os.ModeSymlink != 0 {
		tf.RealPath, err = os.Readlink(tf.Input)
		if err != nil {
			return tf, err
		}
	} else {
		tf.RealPath = tf.Input
	}

	tf.ABS, err = filepath.Abs(tf.Input)
	if err != nil {
		return tf, err
	}

	tf.RelFromWD, err = filepath.Rel(ctx.WorkingDir, tf.ABS)
	if err != nil {
		return tf, err
	}

	tf.RelFromRoot, err = filepath.Rel(ctx.RootDir, tf.ABS)
	if err != nil {
		return tf, err
	}

	tf.DecryptedPath = filepath.Join(ctx.SecretStorePath, tf.RelFromRoot)
	return tf, nil
}
