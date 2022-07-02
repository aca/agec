package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
	"filippo.io/age/agessh"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

type decryptOpt struct {
	IdentityFile string
	Identities   []age.Identity
	Files        []string
	Force        bool
	ctx          *Context
}

func sanitizeDecryptArgs(paths []string) ([]string, error) {
	files := []string{}
	for _, p := range paths {
		fileInfo, err := os.Lstat(p)
		if err != nil {
			return nil, err
		}

		if fileInfo.Mode().IsRegular() {
			if filepath.Ext(p) != ".age" {
				return nil, fmt.Errorf("%q is not valid encrypted file", p)
			} else {
				files = append(files, p)
			}
		}

		if fileInfo.IsDir() {
			err := filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
				if d.Name() == ".git" {
					return filepath.SkipDir
				}

				if err != nil {
					return err
				}

				if !d.Type().IsRegular() {
					return nil
				}

				files = append(files, path)
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
	}

	files = lo.Uniq(files)
	return files, nil
}

func runDecrypt(opts *decryptOpt) error {
	for _, file := range opts.Files {
		decryptedFile := strings.TrimSuffix(file, ".age")
		_, err := os.Stat(decryptedFile)
		if !opts.Force && err == nil {
			log.Printf("skipping %q, as it already exists", decryptedFile)
			continue
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		ager, err := age.Decrypt(f, opts.Identities...)
		if err != nil {
			return fmt.Errorf("failed to decrypt %q: %v", file, err)
		}

		dstFile, err := os.Create(decryptedFile)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, ager)
		if err != nil {
			return fmt.Errorf("failed to copy %q: %v", file, err)
		}

		fmt.Printf("decrypted %q\n", decryptedFile)
	}
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
			opts.Files, err = sanitizeDecryptArgs(args)
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
