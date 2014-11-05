mysql_defaults_file
===================

Access in GO to MySQL via a defaults_file

If using the MySQL command line you can provide a defaults-file which
stores the credentials of the MySQL server you want to connect to,
thus avoiding you to have to specify this information explicitly on the
command line.

This small module allows you to do the same in go.  By default the
default file it uses is ~/.my.cnf.

Usage:

import (
	...
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sjmudd/mysql_defaults_file"
	...
)

// open the connection to the database using the default defaults-file.
dbh, err = mysql_defaults_file.OpenUsingDefaultsFile("mysql", "", "performance_schema")

The errors you get back will be the same as calling sql.Open( "mysql",..... )

Feedback and patches welcome.

Simon J Mudd
<sjmudd@pobox.com>
