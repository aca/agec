package main

import (
	"errors"
	"fmt"

	"filippo.io/age"
	"github.com/samber/lo"
)

const defaultConfigFilename = ".agec.yaml"

type Config struct {
	Version string   `json:"version"`
	Users   []User   `json:"users"`
	Groups  []Group  `json:"groups"`
	Secrets []Secret `json:"secret"`
}

type Group struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

type User struct {
	Name       string   `json:"name"`
	Recipients []string `json:"recipients"`
}

type Secret struct {
	Path   string   `json:"path"`
	Groups []string `yaml:"groups"`
	Users  []string `json:"users"`
}

func (cfg *Config) AddUser(name string, recipients []string) error {
	if cfg.UserExists(name) {
		return fmt.Errorf("user %q already exists", name)
	}

	cfg.Users = append(cfg.Users, User{
		Name:       name,
		Recipients: recipients,
	})

	return nil
}

func (cfg *Config) AddGroup(name string, members []string) error {
	if cfg.GroupExists(name) {
		return fmt.Errorf("group %q already exists", name)
	}

	cfg.Groups = append(cfg.Groups, Group{
		Members: members,
		Name:    name,
	})

	return nil
}

func (cfg *Config) AddGroupMember(group string, user string) error {
	if !cfg.UserExists(user) {
		return fmt.Errorf("user %q not exists", user)
	}

	for gi, g := range cfg.Groups {
		if g.Name == group {
			if lo.Contains(g.Members, user) {
				return fmt.Errorf("%q is already member of group %q", user, group)
			}

			cfg.Groups[gi].Members = append(cfg.Groups[gi].Members, user)
			return nil
		}
	}

	return fmt.Errorf("group %q not exists", group)
}

func (cfg *Config) RemoveGroup(name string) error {
	groups := lo.Filter(cfg.Groups, func(v Group, _ int) bool {
		return v.Name != name
	})

	if len(groups) == len(cfg.Groups) {
		return fmt.Errorf("group %q not exists", name)
	}

	cfg.Groups = groups
	return nil
}

func (cfg *Config) RemoveUser(name string) error {
	users := lo.Filter(cfg.Users, func(v User, _ int) bool {
		return v.Name != name
	})

	if len(users) == len(cfg.Users) {
		return fmt.Errorf("user %q not exists", name)
	}

	cfg.Users = users

	for gi := range cfg.Groups {
		cfg.Groups[gi].Members = lo.Filter(cfg.Groups[gi].Members, func(v string, _ int) bool {
			return v != name
		})
	}

	for si := range cfg.Secrets {
		cfg.Secrets[si].Users = lo.Filter(cfg.Secrets[si].Users, func(v string, _ int) bool {
			return v != name
		})
	}

	return nil
}

func (cfg *Config) RemoveSecret(path string) error {
	secrets := lo.Filter(cfg.Secrets, func(v Secret, _ int) bool {
		return v.Path != path
	})

	if len(secrets) == len(cfg.Secrets) {
		return fmt.Errorf("secret %q not exists", path)
	}

	cfg.Secrets = secrets
	return nil
}

func (cfg *Config) GetGroup(name string) (Group, error) {
	for _, group := range cfg.Groups {
		if group.Name == name {
			return group, nil
		}
	}

	return Group{}, fmt.Errorf("group %q not exists", name)
}

func (cfg *Config) GetUser(name string) (User, error) {
	for _, user := range cfg.Users {
		if user.Name == name {
			return user, nil
		}
	}

	return User{}, fmt.Errorf("user %q not exists", name)
}

func (cfg *Config) GetSecret(path string) (Secret, error) {
	for _, secret := range cfg.Secrets {
		if secret.Path == path {
			return secret, nil
		}
	}

	return Secret{}, fmt.Errorf("secret %q not exists", path)
}

func (cfg *Config) SaveUser(u User) {
	for i := range cfg.Users {
		if cfg.Users[i].Name == u.Name {
			cfg.Users[i] = u
			return
		}
	}

	cfg.Users = append(cfg.Users, u)
}

func (cfg *Config) SaveGroup(g Group) {
	for i := range cfg.Groups {
		if cfg.Groups[i].Name == g.Name {
			cfg.Groups[i] = g
			return
		}
	}

	cfg.Groups = append(cfg.Groups, g)
}

func (cfg *Config) SaveSecret(s Secret) {
	for i := range cfg.Secrets {
		if cfg.Secrets[i].Path == s.Path {
			cfg.Secrets[i] = s
			return
		}
	}

	cfg.Secrets = append(cfg.Secrets, s)
}

func (cfg *Config) GroupExists(name string) bool {
	if _, err := cfg.GetGroup(name); err != nil {
		return false
	}
	return true
}

func (cfg *Config) UserExists(name string) bool {
	if _, err := cfg.GetUser(name); err != nil {
		return false
	}
	return true
}

func (cfg *Config) GetRecipients(users []string, groups []string) ([]age.Recipient, error) {
	keys := []string{}
	recs := []age.Recipient{}

	for _, group := range groups {
		g, err := cfg.GetGroup(group)
		if err != nil {
			return nil, err
		}
		users = append(users, g.Members...)
	}

	for _, user := range users {
		u, err := cfg.GetUser(user)
		if err != nil {
			return nil, err
		}
		keys = append(keys, u.Recipients...)
	}

	keys = lo.Uniq(keys)

	for _, key := range keys {
		rec, err := parseRecipient(key)
		if err != nil {
			return nil, err
		}
		recs = append(recs, rec)
	}

	if len(recs) == 0 {
		return nil, errors.New("no recipients found, specify valid users or groups")
	}

	return recs, nil
}
