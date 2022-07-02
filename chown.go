package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

type chownOpt struct {
	Users  []string
	Groups []string
	Files  []string

	ctx *Context
}

func runChown(opts *chownOpt) error {
	ctx := opts.ctx
	wd := ctx.WorkingDir
	rootDir := ctx.RootDir

	for _, u := range opts.Users {
		if _, err := ctx.Config.GetUser(u); err != nil {
			return err
		}
	}

	for _, g := range opts.Groups {
		if _, err := ctx.Config.GetGroup(g); err != nil {
			return err
		}
	}

	for _, file := range opts.Files {
		relpath, err := filepath.Rel(rootDir, filepath.Join(wd, file))
		if err != nil {
			return err
		}

		secret, err := ctx.Config.GetSecret(relpath)
		if err != nil {
			return fmt.Errorf("%q is not tracked by agec", file)
		}

		if opts.Users != nil {
			secret.Users = opts.Users
		}

		if opts.Groups != nil {
			secret.Groups = opts.Groups
		}

		ctx.Config.SaveSecret(secret)

		fmt.Printf("ownership of %q updated\n", file)
		ctx.Config.SaveSecret(secret)

		// Remove encrypted file as it's owner has changed
		err = os.Remove(file + ".age")
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return ctx.WriteConfigFile()
}

func newChownCmd(ctx *Context) *cobra.Command {
	opts := &chownOpt{
		ctx: ctx,
	}

	cmd := &cobra.Command{
		Use:   "chown",
		Short: "change owner user, group of secrets",
		Example: `  # secret.txt will be encrypted using john's public keys
  agec chown -g devops secret.txt`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveDefault
		},
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
			opts.Files, err = sanitizeChownArgs(args)
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChown(opts)
		},
	}

	f := cmd.Flags()
	f.StringSliceVarP(&opts.Users, "user", "u", nil, "List of users")
	f.StringSliceVarP(&opts.Groups, "group", "g", nil, "List of groups")

	cmd.RegisterFlagCompletionFunc("user", UserCompletion(ctx))
	cmd.RegisterFlagCompletionFunc("group", GroupCompletion(ctx))

	return cmd
}

func sanitizeChownArgs(paths []string) ([]string, error) {
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
