package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type useraddOpt struct {
	User           string
	Groups         []string
	Recipients     []string
	RecipientsFile string

	ctx *Context
}

func runUseradd(opts *useraddOpt) error {
	err := opts.ctx.Config.AddUser(opts.User, opts.Recipients)
	if err != nil {
		return err
	}

	for _, group := range opts.Groups {
		err := opts.ctx.Config.AddGroupMember(group, opts.User)
		if err != nil {
			return err
		}
	}

	return opts.ctx.WriteConfigFile()
}

func newUserAddCmd(ctx *Context) *cobra.Command {
	opts := &useraddOpt{
		ctx: ctx,
	}
	cmd := &cobra.Command{
		Use:   "useradd",
		Short: "Creates a new user with recipients read from stdin",
		Example: `  # create user "aca" with public keys from github
  curl -s "https://github.com/aca.keys" | agec useradd aca -R -`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.ctx == nil {
				return ErrConfigurationNotFound
			}

			var err error
			switch opts.RecipientsFile {
			case "":
				return errors.New("specify recipients-file for the user")
			case "-":
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

			opts.User = args[0]

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUseradd(opts)
		},
		Args: cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	f := cmd.Flags()
	f.StringSliceVarP(&opts.Groups, "group", "g", nil, "Groups which the user is a member of")
	f.StringVarP(&opts.RecipientsFile, "recipients-file", "R", "", "User's recipients, if set to -, the recipients are read from standard input.")
	cmd.RegisterFlagCompletionFunc("group", GroupCompletion(ctx))

	return cmd
}

func readRecipients(r io.Reader) (recs []string, err error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		recs = append(recs, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read recipients: %v", err)
	}

	return recs, nil
}
