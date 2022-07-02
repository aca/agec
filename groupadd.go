package main

import (
	"github.com/spf13/cobra"
)

type groupaddOpt struct {
	Group string
	Users []string

	ctx *Context
}

func runGroupadd(opts *groupaddOpt) error {
	err := opts.ctx.Config.AddGroup(opts.Group, opts.Users)
	if err != nil {
		return err
	}
	return opts.ctx.WriteConfigFile()
}

func newGroupAddCmd(ctx *Context) *cobra.Command {
	opts := &groupaddOpt{
		ctx: ctx,
	}
	cmd := &cobra.Command{
		Use:   "groupadd",
		Short: "Creates agec group",
		Example: `  # add group devops with user john, merry
  agec groupadd devops -u john,merry`,
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
			return runGroupadd(opts)
		},
		Args: cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	f := cmd.Flags()
	f.StringSliceVarP(&opts.Users, "user", "u", nil, "List of users")
	cmd.RegisterFlagCompletionFunc("user", UserCompletion(ctx))

	return cmd
}
