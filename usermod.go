package main

import (
	"os"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

type userModOpt struct {
	Groups         []string
	Append         bool
	User           string
	Recipients     []string
	RecipientsFile string

	ctx *Context
}

func runUserMod(opts *userModOpt) error {
	ctx := opts.ctx
	config := ctx.Config
	u, err := config.GetUser(opts.User)
	if err != nil {
		return err
	}

	if len(opts.Recipients) != 0 {
		u.Recipients = opts.Recipients
	}

	if len(opts.Groups) != 0 {
		config.Groups = lo.Map(config.Groups, func(g Group, _ int) Group {
			if lo.Contains(opts.Groups, g.Name) {
				g.Members = lo.Uniq(append(g.Members, opts.User))
			} else {
				if !opts.Append {
					g.Members = lo.Filter(g.Members, func(m string, _ int) bool {
						return m != opts.User
					})
				}
			}
			return g
		})
	}

	ctx.Config.SaveUser(u)
	return opts.ctx.WriteConfigFile()
}

func newUserModCmd(ctx *Context) *cobra.Command {
	opts := &userModOpt{
		ctx: ctx,
	}
	cmd := &cobra.Command{
		Use:   "usermod",
		Short: "Modify a user account",
		Example: `  # Set "john" member of group "devops", remove all other memberships.
  agec usermod john -g devops
  
  # Add "john" to "devops", "admin" groups
  agec usermod john --append -g devops,admin
  
  # Update recipients of user "aca"
  curl -s "https://github.com/aca.keys" | agec usermod aca -R -`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if ctx == nil {
				return ErrConfigurationNotFound
			}
			opts.User = args[0]

			switch opts.RecipientsFile {
			case "":
				break
			case "-":
				var err error
				opts.Recipients, err = readRecipients(os.Stdin)
				if err != nil {
					return err
				}
			default:
				f, err := os.Open(opts.RecipientsFile)
				if err != nil {
					return err
				}
				opts.Recipients, err = readRecipients(f)
				if err != nil {
					return err
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserMod(opts)
		},
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: UserCompletion(ctx),
	}

	f := cmd.Flags()
	f.StringSliceVarP(&opts.Groups, "group", "g", nil, "List of groups")
	cmd.RegisterFlagCompletionFunc("group", GroupCompletion(ctx))

	f.BoolVarP(&opts.Append, "append", "a", false, "Append memberships to given group lists, rather than replacing it")
	f.StringVarP(&opts.RecipientsFile, "recipients-file", "R", "", "User's recipients, if set to -, the recipients are read from standard input.")
	return cmd
}
