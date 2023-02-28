package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func fileExists(path string) bool {
	_, err := os.Lstat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func WriteFile(filename string, data []byte) (err error) {
	f, err := os.CreateTemp(filepath.Dir(filename), filepath.Base(filename)+".tmp")
	if err != nil {
		return err
	}
	tmpName := f.Name()
	defer func() {
		if err != nil {
			f.Close()
			os.Remove(tmpName)
		}
	}()
	if _, err := f.Write(data); err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		if err := f.Chmod(0o644); err != nil {
			return err
		}
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, filename)
}

func newGroupsCmd(ctx *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "groups",
		Hidden:        true,
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if ctx == nil {
				return ErrConfigurationNotFound
			}
			return nil
		},
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				for _, g := range ctx.Config.Groups {
					fmt.Println(g.Name)
				}
			} else {
				user := args[0]
				for _, g := range ctx.Config.Groups {
					if lo.Contains(g.Members, user) {
						fmt.Println(g.Name)
					}
				}
			}
			return nil
		},
		ValidArgsFunction: UserCompletion(ctx),
	}

	return cmd
}
