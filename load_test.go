package mysql_defaults_file

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFile writes content to name in a new temporary directory and
// returns its path.
func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad(t *testing.T) {
	path := writeFile(t, "my.cnf", `
[mysqld]
port=3399
skip-name-resolve

[client]
user = app
password = "s3cr#t" # comment
host = db1.example.com
port = 3307
socket = /run/mysqld/mysqld.sock
database = shop
no-auto-rehash
`)
	want := Config{
		Filename: path,
		User:     "app",
		Password: "s3cr#t",
		Host:     "db1.example.com",
		Port:     3307,
		Socket:   "/run/mysqld/mysqld.sock",
		Database: "shop",
	}
	for name, load := range map[string]func(string) (Config, error){"Load": Load, "LoadOptional": LoadOptional} {
		got, err := load(path)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got != want {
			t.Errorf("%s: got %+v, want %+v", name, got, want)
		}
	}
}

func TestLoadPort(t *testing.T) {
	tests := []struct {
		value   string
		want    uint16
		wantErr bool
	}{
		{"0", 0, false},
		{"3306", 3306, false},
		{"65535", 65535, false},
		{"'3307'", 3307, false},
		{"-1", 0, true},
		{"65536", 0, true},
		{"70000", 0, true},
		{"99999999999999999999", 0, true},
		{"+3306", 0, true},
		{"0x10", 0, true},
		{"3_306", 0, true},
		{"3306abc", 0, true},
		{"3.5", 0, true},
		{"abc", 0, true},
		{"", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := Load(writeFile(t, "my.cnf", "[client]\nport="+tt.value+"\n"))
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), "not a number from 0 to 65535") {
					t.Errorf("got %+v, %v, want a port error", got, err)
				}
				return
			}
			if err != nil || got.Port != tt.want {
				t.Errorf("got port %d, %v, want %d", got.Port, err, tt.want)
			}
		})
	}
}

func TestLoadNoClientGroup(t *testing.T) {
	path := writeFile(t, "my.cnf", "[mysqld]\nuser=mysql\n")
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := (Config{Filename: path}); got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLoadMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.cnf")

	if _, err := Load(path); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Load: got %v, want a file not found error", err)
	}
	got, err := LoadOptional(path)
	if err != nil {
		t.Fatalf("LoadOptional: %v", err)
	}
	if want := (Config{Filename: path}); got != want {
		t.Errorf("LoadOptional: got %+v, want %+v", got, want)
	}
}

func TestLoadUnreadable(t *testing.T) {
	dir := t.TempDir() // a directory opens but cannot be read as a file
	for name, load := range map[string]func(string) (Config, error){"Load": Load, "LoadOptional": LoadOptional} {
		if _, err := load(dir); err == nil {
			t.Errorf("%s(directory): got no error", name)
		}
	}
}

func TestLoadSyntaxError(t *testing.T) {
	path := writeFile(t, "my.cnf", "user=app\n")
	_, err := Load(path)
	if want := path + ": line 1: option before any group"; err == nil || err.Error() != want {
		t.Errorf("got %v, want %q", err, want)
	}
}

func TestLoadDefaultFileAndHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, ".my.cnf"), []byte("[client]\nuser=app\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"", "~/.my.cnf"} {
		got, err := Load(path)
		if err != nil {
			t.Fatalf("Load(%q): %v", path, err)
		}
		if want := (Config{Filename: filepath.Join(home, ".my.cnf"), User: "app"}); got != want {
			t.Errorf("Load(%q): got %+v, want %+v", path, got, want)
		}
	}
}

func TestExpandHome(t *testing.T) {
	t.Setenv("HOME", "/home/app")
	tests := map[string]string{
		"~":            "/home/app",
		"~/.my.cnf":    "/home/app/.my.cnf",
		"~other/x.cnf": "~other/x.cnf",
		"/etc/~/x.cnf": "/etc/~/x.cnf",
		"x.cnf":        "x.cnf",
	}
	for in, want := range tests {
		got, err := expandHome(in)
		if err != nil || got != want {
			t.Errorf("expandHome(%q) = %q, %v, want %q", in, got, err, want)
		}
	}
}

func TestMySQLConfig(t *testing.T) {
	t.Setenv("USER", "login")
	tests := []struct {
		name     string
		config   Config
		database string
		net      string
		addr     string
		user     string
		dbName   string
	}{
		{"nothing set", Config{}, "", "tcp", "127.0.0.1:3306", "login", ""},
		{"host", Config{Host: "db1"}, "", "tcp", "db1:3306", "login", ""},
		{"host and port", Config{Host: "db1", Port: 3307}, "", "tcp", "db1:3307", "login", ""},
		{"port only", Config{Port: 3307}, "", "tcp", "127.0.0.1:3307", "login", ""},
		{"IPv6 host", Config{Host: "::1"}, "", "tcp", "[::1]:3306", "login", ""},
		{"socket", Config{Socket: "/tmp/mysql.sock"}, "", "unix", "/tmp/mysql.sock", "login", ""},
		{"socket wins over host and port", Config{Socket: "/tmp/mysql.sock", Host: "db1", Port: 3307}, "", "unix", "/tmp/mysql.sock", "login", ""},
		{"user", Config{User: "app"}, "", "tcp", "127.0.0.1:3306", "app", ""},
		{"database from file", Config{Database: "shop"}, "", "tcp", "127.0.0.1:3306", "login", "shop"},
		{"database given wins", Config{Database: "shop"}, "performance_schema", "tcp", "127.0.0.1:3306", "login", "performance_schema"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.Password = "secret"
			mc, err := tt.config.MySQLConfig(tt.database)
			if err != nil {
				t.Fatal(err)
			}
			if mc.Net != tt.net || mc.Addr != tt.addr || mc.User != tt.user || mc.Passwd != "secret" || mc.DBName != tt.dbName {
				t.Errorf("got net %q addr %q user %q passwd %q db %q, want %q %q %q %q %q",
					mc.Net, mc.Addr, mc.User, mc.Passwd, mc.DBName, tt.net, tt.addr, tt.user, "secret", tt.dbName)
			}
			if !mc.AllowNativePasswords {
				t.Error("AllowNativePasswords is false")
			}
		})
	}
}

func TestDSN(t *testing.T) {
	tests := []struct {
		config Config
		want   string
	}{
		{Config{User: "app", Password: "secret", Host: "db1", Port: 3307}, "app:secret@tcp(db1:3307)/performance_schema"},
		{Config{User: "app", Socket: "/tmp/mysql.sock"}, "app@unix(/tmp/mysql.sock)/performance_schema"},
	}
	for _, tt := range tests {
		got, err := tt.config.DSN("performance_schema")
		if err != nil || got != tt.want {
			t.Errorf("DSN(%+v) = %q, %v, want %q", tt.config, got, err, tt.want)
		}
	}
}
