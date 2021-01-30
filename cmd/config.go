package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type config struct {
	org, team, token string
}

func getConfig() (*config, error) {
	v := viper.New()
	v.SetConfigName("githubsync")
	v.AddConfigPath("/etc/githubsync")
	v.AddConfigPath(".")
	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("failed to read config file: %v", err))
	}

	c := new(config)

	if c.org = v.GetString("github.org-slug"); len(c.org) == 0 {
		return nil, errors.New("missing required config: org-slug")
	}

	if c.team = v.GetString("github.team-slug"); len(c.team) == 0 {
		return nil, errors.New("missing required config: team-slug")
	}

	if c.token = v.GetString("github.token"); len(c.token) == 0 {
		return nil, errors.New("missing required config: token")
	}

	return c, nil
}
