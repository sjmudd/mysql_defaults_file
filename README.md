## mysql_defaults_file

Connect to MySQL from Go using the credentials in a MySQL option file,
such as `~/.my.cnf`, as the `mysql` command line client does.

The library reads one option file: the path given, or `~/.my.cnf` if
none is given. From its `[client]` group it takes `user`, `password`,
`host`, `port`, `socket` and `database`, and builds a
[go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)
configuration or DSN from them.

It is used by [ps-top](https://github.com/sjmudd/ps-top).

### Usage

```go
package main

import (
	"database/sql"
	"log"

	"github.com/go-sql-driver/mysql"
	"github.com/sjmudd/mysql_defaults_file"
)

func main() {
	// ~/.my.cnf, if it exists. Use Load("/path/to/my.cnf") for a file
	// that must exist.
	cfg, err := mysql_defaults_file.LoadOptional("")
	if err != nil {
		log.Fatal(err)
	}
	mc, err := cfg.MySQLConfig("mydb")
	if err != nil {
		log.Fatal(err)
	}
	connector, err := mysql.NewConnector(mc)
	if err != nil {
		log.Fatal(err)
	}
	db := sql.OpenDB(connector)
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
}
```

`Config.DSN("mydb")` returns the same configuration as a DSN string for
`sql.Open("mysql", dsn)`.

- `Load` returns an error if the file cannot be read. `LoadOptional`
  treats a file that does not exist as empty, as the MySQL client does
  for `~/.my.cnf`.
- A port that is not a number from 0 to 65535 is an error.
- User defaults to `$USER`. With a socket set, the socket is used.
  Otherwise TCP to the host (default `127.0.0.1`) and port, where port
  0 means 3306, as in the MySQL client.

The file is read with the MySQL client's rules: comments, options
without a value, quoted values, escapes, group names without regard to
case, and "-" and "_" in option names as the same.

Unlike the MySQL client, the library does not:

- read the other default files (`/etc/my.cnf`, `/etc/mysql/my.cnf`,
  `$MYSQL_HOME/my.cnf`, `~/.mylogin.cnf`)
- read the `[mysql]` group
- follow `!include` or `!includedir`

### Deprecated functions

`NewConfig`, `BuildDSN`, `DefaultDSN`, `OpenUsingDefaultsFile`, `Open`
and `OpenUsingEnvironment` still work as before, but are deprecated:
they exit the program instead of returning an error, including on
valid files with options without a value (such as `skip-ssl`) or
`!include` lines, and they ignore a port that is not a number and wrap
one out of range (70000 becomes 4464). Use the functions above instead.

### Development

`make check` runs `gofmt`, `go vet`, `golangci-lint`, the tests and
`go mod tidy -diff`. `make lint-install` installs the pinned
`golangci-lint`.

### Licensing

BSD 2-Clause License

### Feedback

Feedback and patches welcome.

Simon J Mudd
<sjmudd@pobox.com>

### Code Documentation

[pkg.go.dev/github.com/sjmudd/mysql_defaults_file](https://pkg.go.dev/github.com/sjmudd/mysql_defaults_file)
