package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"filippo.io/age"
	"filippo.io/age/agessh"
	"github.com/spf13/cobra"
)

func cmdMain() error {
	log.SetFlags(0)
	ctx, err := initContext()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) && !errors.Is(err, ErrConfigurationNotFound) {
			return err
		}
	}

	rootCmd, err := newRootCmd(ctx, os.Stdout, os.Args[1:])
	if err != nil {
		return err
	}

	err = rootCmd.Execute()
	if err != nil {
		log.Printf("agec: %v", err)
	}
	return err
}

func newRootCmd(ctx *Context, out io.Writer, args []string) (*cobra.Command, error) {
	versionFlag := false
	cmd := &cobra.Command{
		Use:          "agec",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			if versionFlag {
				fmt.Println(version)
			} else {
				cmd.Help()
			}
		},
	}

	f := cmd.PersistentFlags()
	f.BoolP("verbose", "v", false, "verbose output for debugging purposes")
	f.BoolVar(&versionFlag, "version", false, "print version")
	f.Parse(args)

	cmd.AddCommand(
		newInitCmd(ctx),
		newLsCmd(ctx),
		newRmCmd(ctx),

		newEncryptCmd(ctx),
		newDecryptCmd(ctx),

		newUserAddCmd(ctx),
		newUserDelCmd(ctx),
		newUserModCmd(ctx),

		newGroupDelCmd(ctx),
		newGroupAddCmd(ctx),
		newGroupModCmd(ctx),

		newChownCmd(ctx),
		newValidateCmd(ctx),

		newGroupsCmd(ctx),
	)

	return cmd, nil
}

func parseRecipient(arg string) (age.Recipient, error) {
	switch {
	case strings.HasPrefix(arg, "age1"):
		return age.ParseX25519Recipient(arg)
	case strings.HasPrefix(arg, "ssh-"):
		return agessh.ParseRecipient(arg)
	}
	return nil, fmt.Errorf("unknown recipient type: %q", arg)
}

func ParseRecipients(r io.Reader) ([]age.Recipient, error) {
	scanner := bufio.NewScanner(r)

	var recs []age.Recipient

	var n int
	for scanner.Scan() {
		n++
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		r, err := parseRecipient(line)
		if err != nil {
			return nil, fmt.Errorf("malformed recipient at line %d", n)
		}
		recs = append(recs, r)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read recipients file: %v", err)
	}
	if len(recs) == 0 {
		return nil, fmt.Errorf("no recipients found")
	}
	return recs, nil
}
