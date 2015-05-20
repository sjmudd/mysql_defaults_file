// Package mysql_defaults_file provides a way of accessing MySQL via a defaults-file.
package mysql_defaults_file

import (
	"database/sql"
	go_ini "github.com/vaughan0/go-ini" // not sure what to do with dashes in names
	"log"
	"os"
	"strings"
)

// There must be a better way of doing this. Fix me...
// Return the environment value of a given name.
func getEnviron(name string) string {
	for i := range os.Environ() {
		s := os.Environ()[i]
		kV := strings.Split(s, "=")

		if kV[0] == name {
			return kV[1]
		}
	}
	return ""
}

// convert ~ to $HOME
func convertFilename(filename string) string {
	for i := range filename {
		if filename[i] == '~' {
			//			fmt.Println("Filename before", filename )
			filename = filename[:i] + getEnviron("HOME") + filename[i+1:]
			//			fmt.Println("Filename after", filename )
			break
		}
	}

	return filename
}

// Read the given defaults file and return the different parameter values as a map.
func defaultsFileComponents(defaultsFile string) map[string]string {
	defaultsFile = convertFilename(defaultsFile)

	components := make(map[string]string)

	i, err := go_ini.LoadFile(defaultsFile)
	if err != nil {
		log.Fatal("Could not load ini file", err)
	}
	section := i.Section("client")

	user, ok := section["user"]
	if ok {
		components["user"] = user
	}
	password, ok := section["password"]
	if ok {
		components["password"] = password
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
func BuildDSN(components map[string]string, database string) string {
	dsn := ""

	// USER
	_, ok := components["user"]
	if ok {
		dsn = components["user"]
	} else {
		dsn = getEnviron("USER")
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
