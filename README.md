## mysql_defaults_file

Access in Go to MySQL via a defaults_file.

If using the MySQL command line utilities such as `mysql` or
`mysqladmin` you can provide a defaults-file option which stores
the credentials of the MySQL server you want to connect to. If no
specific defaults file is mentioned these utilities look in `~/.my.cnf`
for this file.

The Go sql interface does not support this functionality yet it can
be quite convenient as it avoids the need to explicitly provide credentials.

This small module fills in that gap by providing a function to allow you
to connect to MySQL using a specified defaults-file, or using the
`~/.my.cnf` if you do not specify a defaults-file path.

There is also a function BuildDSN which allows you to build up a Go
dsn for MySQL using various entries in a mysql .ini file.

This logic could be simplified by using
github.com/go-sql-driver/mysql.Config together with Config.FormatDSN(),
but there are a few minor differences in behaviour such as the
default timezone handling using mysql.Config being UTC compared to
mysql's command line client using the system timezone.

The functions provided are used by [ps-top](http://github.com/sjmudd/ps-top)
to simplify the connectivity and have been split off from it as they
may be useful for other programs that connect to MySQL.

The code has been extended to handle quoted usernames and passwords,
removing whitespace and quotes if found. Quoting with single or
double quotes is permitted.

### Usage

Usage:

```
import (
	...
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sjmudd/mysql_defaults_file"
	...
)

// Get the DSN information from the defaults file, provide the
// database to connect to and use sql.Open as normal
db, err = sql.Open("mysql", mysql_defaults_file.DefaultDSN("", "mydb"))

// Get the DSN from a non-standard defaults file, provide the
// database to connect to and use sql.Open as normal
db, err = sql.Open("mysql", mysql_defaults_file.DefaultDSN("/path/to/.my.cnf", "mydb"))

// open the connection to the database using the default defaults-file (original way).
db, err = mysql_defaults_file.OpenUsingDefaultsFile("mysql", "", "performance_schema")

// open the connection to the database using the default defaults-file (shorter form).
db, err = mysql_defaults_file.Open("", "")

// open the connection to the database using a specific defaults-file and to mydb.
db, err = mysql_defaults_file.Open("/path/to/my.ini", "mydb")
```

The errors you get back will be the same as calling `sql.Open( "mysql",..... )`.

### Licensing

BSD 2-Clause License

### Feedback

Feedback and patches welcome.

Simon J Mudd
<sjmudd@pobox.com>

### Code Documenton
[godoc.org/github.com/sjmudd/mysql_defaults_file](http://godoc.org/github.com/sjmudd/mysql_defaults_file)
