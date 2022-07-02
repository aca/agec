package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

type userDelOpt struct {
	User string

	ctx *Context
}

func runUserDel(opts *userDelOpt) error {
	cfg := opts.ctx.Config
	user := opts.User

	err := cfg.RemoveUser(user)
	if err != nil {
		return err
	}

	fmt.Printf("removed user %q\n", user)

	return opts.ctx.WriteConfigFile()
}

func newUserDelCmd(ctx *Context) *cobra.Command {
	opts := &userDelOpt{
		ctx: ctx,
	}
	cmd := &cobra.Command{
		Use:           "userdel",
		Short:         "Deletes agec user",
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if ctx == nil {
				return ErrConfigurationNotFound
			}

			opts.User = args[0]

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserDel(opts)
		},
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: UserCompletion(ctx),
	}

	return cmd
}
