package myapp

import (
	"testing"
	"time"

	"github.com/urfave/cli/v2"
)

func TestParseRoot(t *testing.T) {
	var cfg *Root

	app := NewRootApp(func(c *cli.Context) error {
		var err error
		cfg, err = ParseRoot(c)
		return err
	})

	err := app.Run([]string{"myapp", "--config", "/tmp/test.yaml", "--verbose", "--log-level", "debug", "--host", "0.0.0.0", "--port", "9090", "--timeout", "5s", "--start-time", "2025-01-15T10:30:00Z", "--max-retries", "5", "--tags", "a", "--tags", "b", "--retry-codes", "429", "--retry-codes", "503", "--intervals", "1s", "--intervals", "5s"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Config != "/tmp/test.yaml" {
		t.Fatalf("expected config '/tmp/test.yaml', got %q", cfg.Config)
	}
	if !cfg.Common.Verbose {
		t.Fatal("expected verbose=true")
	}
	if cfg.Common.LogLevel != "debug" {
		t.Fatalf("expected log-level 'debug', got %q", cfg.Common.LogLevel)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Fatalf("expected host '0.0.0.0', got %q", cfg.Server.Host)
	}
	if cfg.Server.Port != 9090 {
		t.Fatalf("expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Server.Timeout.AsDuration() != 5*time.Second {
		t.Fatalf("expected timeout 5s, got %v", cfg.Server.Timeout.AsDuration())
	}
	expectedTime, _ := time.Parse(time.RFC3339, "2025-01-15T10:30:00Z")
	if !cfg.Server.StartTime.AsTime().Equal(expectedTime) {
		t.Fatalf("expected start-time %v, got %v", expectedTime, cfg.Server.StartTime.AsTime())
	}
	if cfg.Server.MaxRetries.GetValue() != 5 {
		t.Fatalf("expected max-retries 5, got %d", cfg.Server.MaxRetries.GetValue())
	}
	if len(cfg.Tags) != 2 || cfg.Tags[0] != "a" || cfg.Tags[1] != "b" {
		t.Fatalf("expected tags [a, b], got %v", cfg.Tags)
	}
	if len(cfg.RetryCodes) != 2 || cfg.RetryCodes[0] != 429 || cfg.RetryCodes[1] != 503 {
		t.Fatalf("expected retry-codes [429, 503], got %v", cfg.RetryCodes)
	}
	if len(cfg.Intervals) != 2 || cfg.Intervals[0].AsDuration() != 1*time.Second || cfg.Intervals[1].AsDuration() != 5*time.Second {
		t.Fatalf("expected intervals [1s, 5s], got %v", cfg.Intervals)
	}
}

func TestParseRootDefaults(t *testing.T) {
	var cfg *Root

	app := NewRootApp(func(c *cli.Context) error {
		var err error
		cfg, err = ParseRoot(c)
		return err
	})

	err := app.Run([]string{"myapp"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Common.LogLevel != "info" {
		t.Fatalf("expected default log-level 'info', got %q", cfg.Common.LogLevel)
	}
	if cfg.Server.Host != "localhost" {
		t.Fatalf("expected default host 'localhost', got %q", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Fatalf("expected default port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Server.Timeout.AsDuration() != 30*time.Second {
		t.Fatalf("expected default timeout 30s, got %v", cfg.Server.Timeout.AsDuration())
	}
	if cfg.Server.StartTime != nil {
		t.Fatal("expected nil start-time when not provided")
	}
	if cfg.Server.MaxRetries.GetValue() != 3 {
		t.Fatalf("expected default max-retries 3, got %d", cfg.Server.MaxRetries.GetValue())
	}
}

func TestParseRootInvalidAllowedValue(t *testing.T) {
	var parseErr error

	app := NewRootApp(func(c *cli.Context) error {
		_, parseErr = ParseRoot(c)
		return parseErr
	})

	err := app.Run([]string{"myapp", "--log-level", "invalid"})
	if err == nil {
		t.Fatal("expected error for invalid allowed value")
	}
	if parseErr == nil {
		t.Fatal("expected ParseRoot to return an error")
	}
}

func TestParseRootInvalidTimestamp(t *testing.T) {
	app := NewRootApp(func(c *cli.Context) error {
		_, err := ParseRoot(c)
		return err
	})

	err := app.Run([]string{"myapp", "--start-time", "not-a-timestamp"})
	if err == nil {
		t.Fatal("expected error for invalid timestamp")
	}
}

func TestRootCommandFlags(t *testing.T) {
	flags := RootCommandFlags()
	names := flagNames(flags)

	expected := []string{"config", "c", "verbose", "v", "log-level", "host", "port", "p", "timeout", "t", "start-time", "max-retries", "tags", "retry-codes", "intervals"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected flag %q not found", name)
		}
	}
}

func TestNewRootCommand(t *testing.T) {
	cmd := NewRootCommand(nil)
	if cmd.Name != "myapp" {
		t.Fatalf("expected command name 'myapp', got %q", cmd.Name)
	}
}

func TestServerAsSubcommand(t *testing.T) {
	var cfg *Server

	app := NewRootApp(nil,
		NewServerCommand(func(c *cli.Context) error {
			var err error
			cfg, err = ParseServer(c)
			return err
		}),
	)

	err := app.Run([]string{"myapp", "server", "--host", "10.0.0.1", "--port", "3000"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Flags.Host != "10.0.0.1" {
		t.Fatalf("expected host '10.0.0.1', got %q", cfg.Flags.Host)
	}
	if cfg.Flags.Port != 3000 {
		t.Fatalf("expected port 3000, got %d", cfg.Flags.Port)
	}
}

func TestServerCommandFlags(t *testing.T) {
	flags := ServerCommandFlags()
	names := flagNames(flags)

	expected := []string{"host", "port", "p", "timeout", "t", "start-time", "max-retries"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected flag %q not found", name)
		}
	}
}

func TestNewServerApp(t *testing.T) {
	app := NewServerApp(nil)
	if app.Name != "server" {
		t.Fatalf("expected app name 'server', got %q", app.Name)
	}
}

func flagNames(flags []cli.Flag) map[string]bool {
	m := map[string]bool{}
	for _, f := range flags {
		for _, n := range f.Names() {
			m[n] = true
		}
	}
	return m
}
