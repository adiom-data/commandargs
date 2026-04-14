package naming

import "testing"

func TestPascalToKebab(t *testing.T) {
	cases := []struct{ in, want string }{
		{"ServerStart", "server-start"},
		{"Root", "root"},
		{"MyApp", "my-app"},
		{"HTTPServer", "h-t-t-p-server"},
		{"simple", "simple"},
	}
	for _, c := range cases {
		got := PascalToKebab(c.in)
		if got != c.want {
			t.Errorf("PascalToKebab(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSnakeToKebab(t *testing.T) {
	cases := []struct{ in, want string }{
		{"log_level", "log-level"},
		{"config_file_path", "config-file-path"},
		{"simple", "simple"},
	}
	for _, c := range cases {
		got := SnakeToKebab(c.in)
		if got != c.want {
			t.Errorf("SnakeToKebab(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestExportName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"server", "Server"},
		{"my-app", "MyApp"},
		{"log-level", "LogLevel"},
	}
	for _, c := range cases {
		got := ExportName(c.in)
		if got != c.want {
			t.Errorf("ExportName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
