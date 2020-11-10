package mysql_defaults_file

import (
	"testing"
)

type userPasswordHost struct {
	user     string
	password string
	host     string
}

// check the different quoted passwords work
func TestPassword(t *testing.T) {
	testIniFiles := []string{
		"testdata/my1.ini",
		"testdata/my2.ini",
		"testdata/my3.ini",
	}
	testInfo := []userPasswordHost{
		{"root", "testpassword1", "127.0.0.1"},
		{"root", "testpassword2", "127.0.0.1"},
		{"root", "testpassword3", "127.0.0.1"},
	}

	for i, _ := range testIniFiles {
		components := defaultsFileComponents(testIniFiles[i])
		if components["host"] != testInfo[i].host || components["user"] != testInfo[i].user || components["password"] != testInfo[i].password {
			t.Errorf("mismatched data for %v (host: %v, user: %v, password: %q, expected: host: %v, user: %v, password: %q)",
				testIniFiles[i],
				components["host"],
				components["user"],
				components["password"],
				testInfo[i].host,
				testInfo[i].user,
				testInfo[i].password,
			)
		}
	}

}
