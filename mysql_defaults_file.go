// Package mysql_defaults_file provides a way of accessing MySQL via a defaults-file.
package mysql_defaults_file

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"strings"

	"github.com/vaughan0/go-ini"
)

var quoteChars = []byte(`"'`)

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

// Read the given defaults file and return the different parameter values as a map.
func defaultsFileComponents(defaultsFile string) map[string]string {
	defaultsFile = convertFilename(defaultsFile)

	components := make(map[string]string)

	i, err := ini.LoadFile(defaultsFile)
	if err != nil {
		log.Fatal("Could not load ini file", err)
	}
	section := i.Section("client")

	user, ok := section["user"]
	if ok {
		// user may have odd characters or be quoted so trim if necessary
		components["user"] = quoteTrim(user)
	}
	password, ok := section["password"]
	if ok {
		// password may have odd characters so trim if quoted
		components["password"] = quoteTrim(password)
	}
	socket, ok := section["socket"]
	if ok {
		components["socket"] = socket
	}
	host, ok := section["host"]
	if ok {
		components["host"] = host
	}
	port, ok := section["port"]
	if ok {
		components["port"] = port
	}
	database, ok := section["database"]
	if ok {
		components["database"] = database
	}

	return components
}

// BuildDSN builds the dsn we're going to use to connect with based on a
// parameter / value string map and return the dsn as a string.
//
// Note: components should be replaced with the mysql.Config structures
// and then we can use Config.FormatDSN() to generate the dsn directly.
// However, there are some differences between mysql.Config default behaviour
// and that from the mysql command line such as timezone handling that would
// need to be taken into account.
func BuildDSN(components map[string]string, database string) string {
	dsn := ""

	// USER
	_, ok := components["user"]
	if ok {
		dsn = components["user"]
	} else {
		dsn = os.Getenv("USER")
	}
	// PASSWORD
	_, ok = components["password"]
	if ok {
		dsn += ":" + components["password"]
	}

	// SOCKET or HOST? SOCKET TAKES PRECEDENCE if both defined.
	_, okSocket := components["socket"]
	_, okHost := components["host"]
	if okSocket || okHost {
		if okSocket {
			dsn += "@unix(" + components["socket"] + ")/"
		} else {
			hostPort := components["host"]
			_, ok := components["port"]
			if ok {
				hostPort += ":" + components["port"] // stored as string so no need to convert
			} else {
				hostPort += ":3306" // we always need _some_ port so if we don't provide one assume MySQL's default port
			}

			dsn += "@tcp(" + hostPort + ")/"
		}
	} else {
		dsn += "@/" // but I'm guessing here.
	}

	if len(database) > 0 {
		dsn += database
	} else {
		_, ok := components["database"]
		if ok {
			dsn += components["database"]
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

// OpenUsingDefaultsFile opens a connection only using a defaults file
func OpenUsingDefaultsFile(sqlDriver string, defaultsFile string, database string) (*sql.DB, error) {
	if defaultsFile == "" {
		defaultsFile = "~/.my.cnf"
	}

	newDSN := BuildDSN(defaultsFileComponents(defaultsFile), database)

	return sql.Open(sqlDriver, newDSN)
}

// OpenUsingEnvironment will assume MYSQL_DSN is set and use that value for connecting.
func OpenUsingEnvironment(sqlDriver string) (*sql.DB, error) {
	if os.Getenv("MYSQL_DSN") == "" {
		return nil, errors.New("MYSQL_DSN not set or empty")
	}

	return sql.Open(sqlDriver, os.Getenv("MYSQL_DSN"))
}
