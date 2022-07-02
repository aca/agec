package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Context struct {
	Debug         bool
	IdentityFile  string
	DefaultUsers  []string
	DefaultGroups []string
	RootDir       string
	RootConfig    string
	WorkingDir    string

	Config *Config
}

func initContext() (*Context, error) {
	var err error

	ctx := &Context{}

	ctx.WorkingDir, err = os.Getwd()
	if err != nil {
		return nil, err
	}

	rootConfig, err := ctx.getRootConfig()
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(rootConfig)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = yaml.Unmarshal(b, cfg)
	if err != nil {
		return nil, err
	}

	ctx.Config = cfg

	if os.Getenv("AGEC_USERS") != "" {
		ctx.DefaultUsers = strings.Split(os.Getenv("AGEC_USERS"), ",")
	}
	if os.Getenv("AGEC_GROUPS") != "" {
		ctx.DefaultGroups = strings.Split(os.Getenv("AGEC_GROUPS"), ",")
	}

	ctx.RootDir = filepath.Dir(rootConfig)
	ctx.RootConfig = rootConfig
	return ctx, nil
}

func (ctx *Context) WriteConfigFile() error {
	b, err := yaml.Marshal(ctx.Config)
	if err != nil {
		return err
	}

	return os.WriteFile(ctx.RootConfig, b, 0o771)
}

var ErrConfigurationNotFound = errors.New("failed to find root config, reached max depth")

func (ctx *Context) getRootConfig() (string, error) {
	d := ctx.WorkingDir
	for i := 0; i < 50; i++ {
		fpath := filepath.Join(d, defaultConfigFilename)
		_, err := os.Lstat(fpath)
		if err == nil {
			return fpath, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}

		parent := filepath.Dir(d)
		if d == parent {
			return "", ErrConfigurationNotFound
		}
		d = parent
	}
	return "", ErrConfigurationNotFound
}
