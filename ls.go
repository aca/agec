package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type lsOpt struct {
	ctx *Context
}

func runLs(opts *lsOpt) error {
	for _, secret := range opts.ctx.Config.Secrets {
		abs := filepath.Join(opts.ctx.RootDir, secret.Path)
		if strings.HasPrefix(abs, opts.ctx.WorkingDir) {
			relpath, err := filepath.Rel(opts.ctx.WorkingDir, abs)
			if err != nil {
				return err
			}
			fmt.Println(relpath)
		}
	}
	return nil
}

func newLsCmd(ctx *Context) *cobra.Command {
	opts := &lsOpt{
		ctx: ctx,
	}

	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List encrypted files under the current directory",
		Example: `  # re-encrypt all secrets under current directory
  agec ls | xargs agec encrypt --force

  # decrypt all secrets under current directory
  agec ls | xargs -I{} agec decrypt {}.age`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if ctx == nil {
				return ErrConfigurationNotFound
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLs(opts)
		},
		ValidArgsFunction: SecretCompletion(ctx),
	}

	return cmd
}
