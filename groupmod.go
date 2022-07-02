package main

import (
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

type groupModOpt struct {
	Group  string
	Users  []string
	Append bool

	ctx *Context
}

func runGroupMod(opts *groupModOpt) error {
	ctx := opts.ctx
	g, err := ctx.Config.GetGroup(opts.Group)
	if err != nil {
		return err
	}

	if opts.Append {
		g.Members = append(g.Members, opts.Users...)
		g.Members = lo.Uniq(g.Members)
	} else {
		g.Members = opts.Users
	}

	return opts.ctx.WriteConfigFile()
}

func newGroupModCmd(ctx *Context) *cobra.Command {
	opts := &groupModOpt{
		ctx: ctx,
	}
	cmd := &cobra.Command{
		Use:   "groupmod",
		Short: "Modify a group definition",
		Example: `  # change "devops" group member to "john","james"
  agec groupmod -u john devops
  
  # append member "john" to "devops" group
  agec groupmod -u john -a devops`,
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
			return runGroupMod(opts)
		},
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: GroupCompletion(ctx),
	}

	f := cmd.Flags()
	f.StringSliceVarP(&opts.Users, "user", "u", nil, "List of users")
	cmd.RegisterFlagCompletionFunc("user", UserCompletion(ctx))

	f.BoolVarP(&opts.Append, "append", "a", false, "Append users to the existing member list, rather than replacing it")
	return cmd
}
