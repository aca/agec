package main

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func UserCompletion(ctx *Context) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		users := []string{}
		for _, user := range ctx.Config.Users {
			users = append(users, user.Name)
		}
		return users, cobra.ShellCompDirectiveNoFileComp
	}
}

func GroupCompletion(ctx *Context) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		groups := []string{}
		for _, group := range ctx.Config.Groups {
			groups = append(groups, group.Name)
		}
		return groups, cobra.ShellCompDirectiveNoFileComp
	}
}

func SecretCompletion(ctx *Context) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var completions []string

		for _, secret := range ctx.Config.Secrets {
			abs := filepath.Join(ctx.RootDir, secret.Path)
			if strings.HasPrefix(abs, ctx.WorkingDir) {
				relpath, err := filepath.Rel(ctx.WorkingDir, abs)
				if err != nil {
					continue
				}
				completions = append(completions, relpath)
			}
		}
		return completions, cobra.ShellCompDirectiveDefault
	}
}

func EncryptedFileCompletion(ctx *Context) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var candidates []string

		filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.Type().IsRegular() {
				return nil
			}

			if filepath.Ext(path) != ".age" {
				return nil
			}

			candidates = append(candidates, path)

			return nil
		})

		var completions []string
		for _, comp := range candidates {
			if strings.HasPrefix(comp, toComplete) {
				completions = append(completions, comp)
			}
		}

		return completions, cobra.ShellCompDirectiveDefault
	}
}
