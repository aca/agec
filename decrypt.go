package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
	"filippo.io/age/agessh"
	"github.com/spf13/cobra"
)

type decryptOpt struct {
	IdentityFile string
	Identities   []age.Identity
	FileArg      string
	Force        bool
	ctx          *Context
}

func sanitizeDecryptArg(arg string, workingDir string) (string, error) {
	if filepath.Ext(arg) != ".age" {
		return "", fmt.Errorf("%q is not valid encrypted file", arg)
	}

	fileInfo, err := os.Lstat(arg)
	if err != nil {
		return "", err
	}

	if !fileInfo.Mode().IsRegular() {
		return "", fmt.Errorf("%q is not a regular file", arg)
	}

	if !filepath.IsAbs(arg) {
		arg = filepath.Join(workingDir, arg)
	}

	relpath, err := filepath.Rel(workingDir, arg)
	if err != nil {
		return "", fmt.Errorf("fail to get relpath of %q: %v", arg, err)
	}
	return relpath, nil
}

func runDecrypt(opts *decryptOpt) error {
	file := opts.FileArg

	decryptedFileLinkPath := filepath.Join(opts.ctx.WorkingDir, strings.TrimSuffix(file, ".age"))
	if !opts.Force && fileExists(decryptedFileLinkPath) {
		return fmt.Errorf("skipping %q. Use --force to re-decrypt file", decryptedFileLinkPath)
	}

	relPath, err := filepath.Rel(opts.ctx.RootDir, decryptedFileLinkPath)
	if err != nil {
		return err
	}
	decryptedFilePath := filepath.Join(opts.ctx.SecretStorePath, relPath)

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	ager, err := age.Decrypt(f, opts.Identities...)
	if err != nil {
		return fmt.Errorf("failed to decrypt %q: %v", file, err)
	}

	err = os.MkdirAll(filepath.Dir(decryptedFilePath), 0o777)
	if err != nil && os.IsNotExist(err) {
		return err
	}

	dstFile, err := os.Create(decryptedFilePath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, ager)
	if err != nil {
		return fmt.Errorf("failed to copy %q: %v", file, err)
	}

	
	// relp, err := filepath.Rel(opts.ctx.WorkingDir, decryptedFilePath)
	relp, err := filepath.Rel(filepath.Dir(filepath.Join(opts.ctx.WorkingDir, file)), decryptedFilePath)
	if err != nil {
		return err
	}

	err = os.Symlink(relp, decryptedFileLinkPath)
	if err != nil && !os.IsExist(err) {
		return err
	}

	fmt.Printf("decrypted %q\n", decryptedFileLinkPath)
	return nil
}

func newDecryptCmd(ctx *Context) *cobra.Command {
	opts := &decryptOpt{
		ctx: ctx,
	}
	cmd := &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt secrets tracked by agec",
		Example: `  # By default, agec will try with key found in '~/.ssh/*'
  agec decrypt secret.txt
  
  # Specify identity file to decrypt
  agec decrypt secret.txt -i ~/.ssh/id_rsa

  # or pass private key to stdin
  cat ~/.ssh/id_rsa | agec decrypt secret.txt -i -
  
  # or set env AGEC_IDENTITY_FILE
  AGEC_IDENTITY_FILE=~/.ssh/id_rsa agec decrypt secret.txt
  
  # decrypt all in current directory
  fd --extension age | xargs agec decrypt`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if ctx == nil {
				return ErrConfigurationNotFound
			}

			if opts.IdentityFile == "" {
				homedir, err := os.UserHomeDir()
				if err != nil {
					return err
				}

				sshdir := filepath.Join(homedir, ".ssh")

				keyfiles, err := ioutil.ReadDir(sshdir)
				if err != nil {
					return err
				}

				for _, keyfile := range keyfiles {
					if keyfile.IsDir() {
						continue
					}
					ident, err := parseIdentitiesFile(filepath.Join(sshdir, keyfile.Name()))
					if err != nil {
						// TODO: debug log
						continue
					}
					opts.Identities = append(opts.Identities, ident...)
				}

				if len(opts.Identities) == 0 {
					return errors.New("failed to load identities from ~/.ssh, specify identity file")
				}

			} else {
				ident, err := parseIdentitiesFile(opts.IdentityFile)
				if err != nil {
					return err
				}
				opts.Identities = append(opts.Identities, ident...)
			}

			var err error
			opts.FileArg, err = sanitizeDecryptArg(args[0], ctx.WorkingDir)
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDecrypt(opts)
		},
		ValidArgsFunction: EncryptedFileCompletion(ctx),
	}

	f := cmd.Flags()
	f.StringVarP(&opts.IdentityFile, "identity", "i", os.Getenv("AGEC_IDENTITY"), "Specify identity file to decrypt file")
	f.BoolVarP(&opts.Force, "force", "f", false, "Overwrite an existing encrypted file")

	return cmd
}

// https://github.com/FiloSottile/age/blob/cff70cffe2f665ef67cf243daafb064f0feb61a9/cmd/age/parse.go#L126
func parseIdentitiesFile(name string) ([]age.Identity, error) {
	var f *os.File
	if name == "-" {
		f = os.Stdin
	} else {
		var err error
		f, err = os.Open(name)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %v", err)
		}
		defer f.Close()
	}

	b := bufio.NewReader(f)
	p, _ := b.Peek(14) // length of "age-encryption" and "-----BEGIN AGE"
	peeked := string(p)

	switch {
	case strings.HasPrefix(peeked, "-----BEGIN"):
		const privateKeySizeLimit = 1 << 14 // 16 KiB
		contents, err := io.ReadAll(io.LimitReader(b, privateKeySizeLimit))
		if err != nil {
			return nil, fmt.Errorf("failed to read %q: %v", name, err)
		}
		if len(contents) == privateKeySizeLimit {
			return nil, fmt.Errorf("failed to read %q: file too long", name)
		}
		return parseSSHIdentity(name, contents)

	// An unencrypted age identity file.
	default:
		ids, err := age.ParseIdentities(b)
		if err != nil {
			return nil, fmt.Errorf("failed to read %q: %v", name, err)
		}
		return ids, nil
	}
}

func parseSSHIdentity(name string, pemBytes []byte) ([]age.Identity, error) {
	id, err := agessh.ParseIdentity(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("malformed SSH identity in %q: %v", name, err)
	}

	return []age.Identity{id}, nil
}
