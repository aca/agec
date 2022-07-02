package main

import (
	"github.com/spf13/cobra"
)

type groupDelOpt struct {
	Group string

	ctx *Context
}

func runGroupDel(opts *groupDelOpt) error {
	cfg := opts.ctx.Config
	group := opts.Group

	err := cfg.RemoveGroup(group)
	if err != nil {
		return err
	}
	return opts.ctx.WriteConfigFile()
}

func newGroupDelCmd(ctx *Context) *cobra.Command {
	opts := &groupDelOpt{
		ctx: ctx,
	}
	cmd := &cobra.Command{
		Use:           "groupdel",
		Short:         "Deletes agec group",
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if ctx == nil {
				return ErrConfigurationNotFound
			}

			opts.Group = args[0]

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGroupDel(opts)
		},
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: GroupCompletion(ctx),
	}

	return cmd
}
