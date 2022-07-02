package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"filippo.io/age"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

type encryptOpt struct {
	Users  []string
	Groups []string
	Files  []string
	Force  bool

	ctx *Context
}

func runEncrypt(opts *encryptOpt) error {
	ctx := opts.ctx
	wd := ctx.WorkingDir

	for _, file := range opts.Files {
		encryptedFile := file + ".age"

		if fileExists(encryptedFile) && !opts.Force {
			log.Printf("skipping %q. Use --force to re-encrypt file", file)
			continue
		}

		var recs []age.Recipient

		relpath, err := filepath.Rel(ctx.RootDir, filepath.Join(wd, file))
		if err != nil {
			return err
		}

		secret, err := ctx.Config.GetSecret(relpath)
		if err != nil {
			// new secret
			secret = Secret{
				Users:  opts.Users,
				Groups: opts.Groups,
				Path:   relpath,
			}

			recs, err = ctx.Config.GetRecipients(opts.Users, opts.Groups)
			if err != nil {
				return err
			}
		} else {
			// existing secret
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

		plainFile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer plainFile.Close()

		_, err = io.Copy(agew, plainFile)
		if err != nil {
			return err
		}

		if agew.Close() != nil {
			return err
		}

		err = os.WriteFile(encryptedFile, out.Bytes(), 0o644)
		if err != nil {
			return err
		}

		fmt.Printf("encrypted %q\n", encryptedFile)

		if plainFile.Close() != nil {
			return err
		}

		err = os.Remove(file)
		if err != nil {
			return err
		}

		fmt.Printf("rm %q\n", file)

		ctx.Config.SaveSecret(secret)

	}

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
		Args:          cobra.MinimumNArgs(1),
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
			opts.Files, err = sanitizeEncryptArgs(args)
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

func sanitizeEncryptArgs(paths []string) ([]string, error) {
	files := []string{}

	for _, p := range paths {
		fileInfo, err := os.Lstat(p)
		if err != nil {
			return nil, err
		}

		if !fileInfo.Mode().IsRegular() {
			return nil, fmt.Errorf("%q is not a regular file", p)
		}

		files = append(files, p)
	}

	files = lo.Uniq(files)
	return files, nil
}
