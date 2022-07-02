package main

import (
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
)

type rmOpt struct {
	Files []string

	ctx *Context
}

func runRm(opts *rmOpt) error {
	for _, f := range opts.Files {
		abs := filepath.Join(opts.ctx.WorkingDir, f)
		rel, err := filepath.Rel(opts.ctx.RootDir, abs)
		if err != nil {
			return err
		}

		err = opts.ctx.Config.RemoveSecret(rel)
		if err != nil {
			return err
		} else {
			log.Printf("removed secret %q", rel)
		}
	}

	return opts.ctx.WriteConfigFile()
}

func newRmCmd(ctx *Context) *cobra.Command {
	opts := &rmOpt{
		ctx: ctx,
	}
	cmd := &cobra.Command{
		Use:           "rm",
		Short:         "Remove secret file and untrack from agec",
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if ctx == nil {
				return ErrConfigurationNotFound
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Files = args
			return runRm(opts)
		},
		ValidArgsFunction: SecretCompletion(ctx),
	}
	return cmd
}
