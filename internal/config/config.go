package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Role string

const (
	RoleServer Role = "server"
	RoleClient Role = "client"
)

type ScreenConfig struct {
	Name        string `toml:"name"`
	EdgeLeft    string `toml:"edge_left"`
	EdgeRight   string `toml:"edge_right"`
	EdgeTop     string `toml:"edge_top"`
	EdgeBottom  string `toml:"edge_bottom"`
}

type Config struct {
	Role       Role           `toml:"role"`
	ListenAddr string         `toml:"listen_addr"`
	ServerAddr string         `toml:"server_addr"`
	ScreenName string         `toml:"screen_name"`
	Secret     string         `toml:"secret"`
	Screens    []ScreenConfig `toml:"screens"`
}

func Load(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("load config %q: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	switch c.Role {
	case RoleServer:
		if c.ListenAddr == "" {
			return fmt.Errorf("server role requires listen_addr")
		}
		if len(c.Screens) == 0 {
			return fmt.Errorf("server role requires at least one [[screens]] entry")
		}
		names := make(map[string]bool)
		for _, s := range c.Screens {
			if s.Name == "" {
				return fmt.Errorf("each screen must have a name")
			}
			if names[s.Name] {
				return fmt.Errorf("duplicate screen name %q", s.Name)
			}
			names[s.Name] = true
		}
		// validate edge references
		for _, s := range c.Screens {
			for _, edge := range []string{s.EdgeLeft, s.EdgeRight, s.EdgeTop, s.EdgeBottom} {
				if edge != "" && !names[edge] {
					return fmt.Errorf("screen %q references unknown screen %q in edge", s.Name, edge)
				}
			}
		}
	case RoleClient:
		if c.ServerAddr == "" {
			return fmt.Errorf("client role requires server_addr")
		}
		if c.ScreenName == "" {
			return fmt.Errorf("client role requires screen_name")
		}
	default:
		return fmt.Errorf("role must be %q or %q, got %q", RoleServer, RoleClient, c.Role)
	}
	return nil
}
