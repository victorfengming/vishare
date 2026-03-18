package config_test

import (
	"os"
	"testing"

	"github.com/victorfengming/vishare/internal/config"
	"github.com/victorfengming/vishare/internal/defaults"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "vishare-cfg-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestServerValid(t *testing.T) {
	p := writeTemp(t, `
role        = "server"
listen_addr = "0.0.0.0:24800"

[[screens]]
name        = "ubuntu-main"
edge_right  = "windows-pc"

[[screens]]
name      = "windows-pc"
edge_left = "ubuntu-main"
`)
	cfg, err := config.Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Role != config.RoleServer {
		t.Fatalf("role mismatch")
	}
	if len(cfg.Screens) != 2 {
		t.Fatalf("expected 2 screens, got %d", len(cfg.Screens))
	}
}

func TestClientValid(t *testing.T) {
	p := writeTemp(t, `
role        = "client"
server_addr = "192.168.100.98:24800"
screen_name = "windows-pc"
`)
	cfg, err := config.Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Role != config.RoleClient {
		t.Fatalf("role mismatch")
	}
	if cfg.ServerAddr != "192.168.100.98:24800" {
		t.Fatalf("server_addr mismatch")
	}
	if cfg.MouseSpeed != defaults.MouseSpeed {
		t.Fatalf("mouse_speed default mismatch: got %v", cfg.MouseSpeed)
	}
}

func TestClientMouseSpeedConfigured(t *testing.T) {
	p := writeTemp(t, `
role        = "client"
server_addr = "192.168.100.98:24800"
screen_name = "windows-pc"
mouse_speed = 0.25
`)
	cfg, err := config.Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MouseSpeed != 0.25 {
		t.Fatalf("mouse_speed mismatch: got %v", cfg.MouseSpeed)
	}
}

func TestMissingRole(t *testing.T) {
	p := writeTemp(t, `listen_addr = "0.0.0.0:24800"`)
	_, err := config.Load(p)
	if err == nil {
		t.Fatal("expected error for missing role")
	}
}

func TestServerMissingListenAddr(t *testing.T) {
	p := writeTemp(t, `
role = "server"
[[screens]]
name = "s1"
`)
	_, err := config.Load(p)
	if err == nil {
		t.Fatal("expected error for missing listen_addr")
	}
}

func TestServerDuplicateScreen(t *testing.T) {
	p := writeTemp(t, `
role        = "server"
listen_addr = "0.0.0.0:24800"

[[screens]]
name = "dup"

[[screens]]
name = "dup"
`)
	_, err := config.Load(p)
	if err == nil {
		t.Fatal("expected error for duplicate screen name")
	}
}

func TestServerUnknownEdge(t *testing.T) {
	p := writeTemp(t, `
role        = "server"
listen_addr = "0.0.0.0:24800"

[[screens]]
name       = "ubuntu-main"
edge_right = "nonexistent"
`)
	_, err := config.Load(p)
	if err == nil {
		t.Fatal("expected error for unknown edge reference")
	}
}

func TestClientMissingServerAddr(t *testing.T) {
	p := writeTemp(t, `
role        = "client"
screen_name = "windows-pc"
`)
	_, err := config.Load(p)
	if err == nil {
		t.Fatal("expected error for missing server_addr")
	}
}

func TestClientInvalidMouseSpeed(t *testing.T) {
	p := writeTemp(t, `
role        = "client"
server_addr = "192.168.100.98:24800"
screen_name = "windows-pc"
mouse_speed = 1.5
`)
	_, err := config.Load(p)
	if err == nil {
		t.Fatal("expected error for invalid mouse_speed")
	}
}
