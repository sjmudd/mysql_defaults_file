// Package mysql_defaults_file provides a way of accessing MySQL via a defaults-file.
package mysql_defaults_file

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/vaughan0/go-ini"
)

const (
	defaultMySQLPort   = 3306
	defaultMySQLDriver = "mysql"
)

var quoteChars = []byte(`"'`)

// Config holds the configuration taken out of the defaults file.
type Config struct {
	Filename string
	User     string
	Password string
	Socket   string
	Host     string
	Port     uint16
	Database string
}

// convert ~ to $HOME
func convertFilename(filename string) string {
	for i := range filename {
		if filename[i] == '~' {
			filename = filename[:i] + os.Getenv("HOME") + filename[i+1:]
			break
		}
	}

	return filename
}

// quoteTrim will remove leading/trailing whitespace and if the value
// looks like a quoted string remove the quotes.
func quoteTrim(val string) string {
	val = strings.TrimSpace(val)
	lenval := len(val)
	if lenval >= 2 {
		for _, ch := range quoteChars {
			if val[0] == ch && val[lenval-1] == ch {
				return val[1 : lenval-1]
			}
		}
	}
	return val
}

// NewConfig creates a Config struct using the values from the provided defaults file.
// - if defaultsFile is empty use the default of ~/.my.cnf
func NewConfig(defaultsFile string) Config {
	var config Config

	if defaultsFile == "" {
		defaultsFile = "~/.my.cnf"
	}

	defaultsFile = convertFilename(defaultsFile)
	config.Filename = defaultsFile

	i, err := ini.LoadFile(defaultsFile)
	if err != nil {
		log.Fatalf("Could not load defaults-file %q: %v", defaultsFile, err)
	}
	section := i.Section("client")

	user, ok := section["user"]
	if ok {
		// user may have odd characters or be quoted so trim if necessary
		config.User = quoteTrim(user)
	}
	password, ok := section["password"]
	if ok {
		// password may have odd characters so trim if quoted
		config.Password = quoteTrim(password)
	}
	socket, ok := section["socket"]
	if ok {
		config.Socket = socket
	}
	host, ok := section["host"]
	if ok {
		config.Host = host
	}
	port, ok := section["port"]
	if ok {
		port, err := strconv.Atoi(port)
		if err == nil {
			config.Port = uint16(port)
		}
	}
	database, ok := section["database"]
	if ok {
		config.Database = database
	}

	return config
}

// BuildDSN builds the dsn we're going to use to connect with based on a
// parameter / value string map and return the dsn as a string.
//
// Note: Config should be replaced with the mysql.Config structure
// and then we can use Config.FormatDSN() to generate the dsn directly.
// However, there are some differences between mysql.Config default behaviour
// and that from the mysql command line such as timezone handling that would
// need to be taken into account.
func BuildDSN(config Config, database string) string {
	dsn := ""

	// USER
	if config.User != "" {
		dsn = config.User
	} else {
		dsn = os.Getenv("USER")
	}
	// PASSWORD
	if config.Password != "" {
		dsn += ":" + config.Password
	}

	// SOCKET or HOST? SOCKET TAKES PRECEDENCE if both defined.
	if config.Socket != "" || config.Host != "" {
		if config.Socket != "" {
			dsn += "@unix(" + config.Socket + ")/"
		} else {
			hostPort := config.Host
			if config.Port != 0 {
				hostPort += ":" + strconv.Itoa(int(config.Port))
			} else {
				hostPort += ":" + strconv.Itoa(defaultMySQLPort)
			}

			dsn += "@tcp(" + hostPort + ")/"
		}
	} else {
		dsn += "@/" // but I'm guessing here.
	}

	if database != "" {
		dsn += database
	} else {
		if config.Database != "" {
			dsn += config.Database
		}
	}

	// Hard-code allowNativePassword=true for consistent behaviour
	// with mysql command line when talking to a 8.0 server using
	// caching-sha2 and trying to authenticate a user with
	// mysql-native-password.
	dsn += "?allowNativePasswords=true"

	//	fmt.Println("final dsn from defaults file:", dsn )
	return dsn
}

// BuildDSNFromDefaultsfile provides a dsn from the combination of a defaults file and the database to connect to
// - this format is the simplest to use as if you have a ~/.my.cnf
//   file you simply want to use this function to generate the dsn
//   to connect to
func BuildDSNFromDefaultsFile(defaultsFile string, database string) string {
	return BuildDSN(NewConfig(defaultsFile), database)
}

// OpenUsingDefaultsFile opens a connection only using a defaults file
func OpenUsingDefaultsFile(sqlDriver string, defaultsFile string, database string) (*sql.DB, error) {
	return sql.Open(sqlDriver, BuildDSNFromDefaultsFile(defaultsFile, database))
}

// Open just wraps OpenUsingDefaultsFile, assuming "mysql" as the driver.
func Open(defaultsFile string, database string) (*sql.DB, error) {
	return OpenUsingDefaultsFile(defaultMySQLDriver, defaultsFile, database)
}

// OpenUsingEnvironment will assume MYSQL_DSN is set and use that value for connecting.
func OpenUsingEnvironment(sqlDriver string) (*sql.DB, error) {
	if os.Getenv("MYSQL_DSN") == "" {
		return nil, errors.New("MYSQL_DSN not set or empty")
	}

	return sql.Open(sqlDriver, os.Getenv("MYSQL_DSN"))
}
