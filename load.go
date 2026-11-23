package mysql_defaults_file

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
)

const (
	// DefaultFile is the option file read when no path is given.
	DefaultFile = "~/.my.cnf"
	defaultHost = "127.0.0.1"
)

// Load reads the [client] group of the option file at path, or of
// DefaultFile if path is empty. A file that cannot be read or parsed,
// or a port that is not a number from 0 to 65535, is an error.
func Load(path string) (Config, error) {
	return load(path, false)
}

// LoadOptional is Load, except that a file that does not exist is not
// an error: it returns a Config with only Filename set.
func LoadOptional(path string) (Config, error) {
	return load(path, true)
}

func load(path string, optional bool) (Config, error) {
	if path == "" {
		path = DefaultFile
	}
	filename, err := expandHome(path)
	if err != nil {
		return Config{}, err
	}
	config := Config{Filename: filename}

	f, err := os.Open(filename)
	if err != nil {
		if optional && errors.Is(err, fs.ErrNotExist) {
			return config, nil
		}
		return Config{}, err
	}
	defer func() { _ = f.Close() }()

	of, err := parseOptionFile(f)
	if err != nil {
		return Config{}, fmt.Errorf("%s: %w", filename, err)
	}
	client := of["client"]
	config.User = client["user"]
	config.Password = client["password"]
	config.Host = client["host"]
	config.Socket = client["socket"]
	config.Database = client["database"]
	if value, ok := client["port"]; ok {
		port, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return Config{}, fmt.Errorf("%s: port %q: not a number from 0 to 65535", filename, value)
		}
		config.Port = uint16(port)
	}
	return config, nil
}

// expandHome replaces a leading "~" or "~/" with the home directory.
func expandHome(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, path[1:]), nil
}

// MySQLConfig returns the driver configuration for c, connecting to
// database, or to c.Database if database is empty:
//
//   - user: c.User, else $USER
//   - a socket if c.Socket is set, else TCP to c.Host (default 127.0.0.1)
//     and c.Port, where port 0 means 3306, as in the MySQL client
//
// Use it with mysql.NewConnector and sql.OpenDB.
func (c Config) MySQLConfig(database string) (*mysql.Config, error) {
	mc := mysql.NewConfig()
	mc.User = c.User
	if mc.User == "" {
		mc.User = os.Getenv("USER")
	}
	mc.Passwd = c.Password
	if c.Socket != "" {
		mc.Net = "unix"
		mc.Addr = c.Socket
	} else {
		host, port := c.Host, int(c.Port)
		if host == "" {
			host = defaultHost
		}
		if port == 0 {
			port = defaultMySQLPort
		}
		mc.Net = "tcp"
		mc.Addr = net.JoinHostPort(host, strconv.Itoa(port))
	}
	mc.DBName = database
	if mc.DBName == "" {
		mc.DBName = c.Database
	}
	return mc, nil
}

// DSN returns MySQLConfig(database) as a DSN for sql.Open("mysql", ...).
func (c Config) DSN(database string) (string, error) {
	mc, err := c.MySQLConfig(database)
	if err != nil {
		return "", err
	}
	return mc.FormatDSN(), nil
}
