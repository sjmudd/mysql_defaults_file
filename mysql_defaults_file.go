// mysql_defaults_file provides a way of accessing MySQL via a defaults-file.
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
func get_environ(name string) string {
	for i := range os.Environ() {
		s := os.Environ()[i]
		k_v := strings.Split(s, "=")

		if k_v[0] == name {
			return k_v[1]
		}
	}
	return ""
}

// convert ~ to $HOME
func convert_filename(filename string) string {
	for i := range filename {
		if filename[i] == '~' {
			//			fmt.Println("Filename before", filename )
			filename = filename[:i] + get_environ("HOME") + filename[i+1:]
			//			fmt.Println("Filename after", filename )
			break
		}
	}

	return filename
}

// read the defaults file and return the values
func defaults_file_components(defaults_file string) map[string]string {
	defaults_file = convert_filename(defaults_file)

	components := make(map[string]string)

	i, err := go_ini.LoadFile(defaults_file)
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

// build the dsn we're going to use to connect with based on the
// components in the defaults file
func build_dsn(components map[string]string, database string) string {
	dsn := ""

	// USER
	_, ok := components["user"]
	if ok {
		dsn = components["user"]
	} else {
		dsn = get_environ("USER")
	}
	// PASSWORD
	_, ok = components["password"]
	if ok {
		dsn += ":" + components["password"]
	}

	// SOCKET or HOST? SOCKET TAKES PRECEDENCE if both defined.
	_, ok_socket := components["socket"]
	_, ok_host := components["host"]
	if ok_socket || ok_host {
		if ok_socket {
			dsn += "@unix(" + components["socket"] + ")/"
		} else {
			host_port := components["host"]
			_, ok := components["port"]
			if ok {
				host_port += ":" + components["port"] // stored as string so no need to convert
			} else {
				host_port += ":3306" // we always need _some_ port so if we don't provide one assume MySQL's default port
			}

			dsn += "@tcp(" + host_port + ")/"
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

// open a connection only using a defaults file
func OpenUsingDefaultsFile(sql_driver string, defaults_file string, database string) (*sql.DB, error) {
	if defaults_file == "" {
		defaults_file = "~/.my.cnf"
	}

	new_dsn := build_dsn(defaults_file_components(defaults_file), database)

	return sql.Open(sql_driver, new_dsn)
}
