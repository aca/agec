package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"filippo.io/age"
	"github.com/spf13/cobra"
)

type initOpt struct {
	out io.Writer

	ctx *Context
}

func runInit(opts *initOpt) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	rootConfig := filepath.Join(wd, defaultConfigFilename)

	if opts.ctx != nil {
		if opts.ctx.RootConfig == rootConfig {
			return fmt.Errorf("%q already exists", opts.ctx.RootConfig)
		}
	}

	k, err := age.GenerateX25519Identity()
	if err != nil {
		return err
	}

	root := User{
		Name: "root",
		Recipients: []string{
			k.Recipient().String(),
		},
	}

	ctx := &Context{}
	ctx.RootDir = wd
	ctx.RootConfig = rootConfig
	ctx.Config = &Config{
		Version: "v1",
		Users: []User{
			root,
		},
		Secrets: []Secret{},
		Groups: []Group{
			{
				Name: "root",
				Members: []string{
					"root",
				},
			},
		},
	}

	fmt.Fprintf(os.Stderr, "Initalized agec in %q\n", rootConfig)
	fmt.Fprintf(os.Stderr, "\nGenerated user:root, group:root\n")
	fmt.Fprintf(os.Stdout, "# created: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(os.Stdout, "# public key: %s\n", k.Recipient())
	fmt.Fprintf(os.Stdout, "%s\n", k)
	return ctx.WriteConfigFile()
}

func newInitCmd(ctx *Context) *cobra.Command {
	opts := &initOpt{
		ctx: ctx,
	}
	cmd := &cobra.Command{
		Use:           "init",
		Short:         "Initalizes agec under the current directory",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(opts)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
	return cmd
}
