package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

type ValidateOpt struct {
	ctx *Context
}

func runValidate(opts *ValidateOpt) error {
	ctx := opts.ctx
	for _, secret := range ctx.Config.Secrets {
		abs := filepath.Join(ctx.RootDir, secret.Path)

		if fileExists(abs) {
			rel, err := filepath.Rel(ctx.WorkingDir, abs)
			if err != nil {
				return err
			}
			return fmt.Errorf("%q not encrypted", rel)
		}
	}
	return nil
}

func newValidateCmd(ctx *Context) *cobra.Command {
	opts := &ValidateOpt{
		ctx: ctx,
	}

	cmd := &cobra.Command{
		Use:           "validate",
		Short:         "Find unencrypted secrets, validate configurations.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if ctx == nil {
				return ErrConfigurationNotFound
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(opts)
		},
	}

	return cmd
}
