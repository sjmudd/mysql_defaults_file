package mysql_defaults_file

import (
	"testing"
)

// check the different quoted passwords work
func TestUserAndPassword(t *testing.T) {
	testIniFiles := []string{
		"testdata/my1.ini",
		"testdata/my2.ini",
		"testdata/my3.ini",
		"testdata/my4.ini",
		"testdata/my5.ini",
	}
	wanted := []Config{
		{Host: "127.0.0.1", Port: 3306, User: "root1", Password: "testpassword1"},
		{Host: "127.0.0.1", Port: 3306, User: "root1", Password: "testpassword2"},
		{Host: "127.0.0.1", Port: 3306, User: "root1", Password: "testpassword3"},
		{Host: "127.0.0.1", Port: 3306, User: "root4", Password: "testpassword1"},
		{Host: "127.0.0.1", Port: 3307, User: "root4", Password: "testpassword1"},
	}

	for i := range testIniFiles {
		config := NewConfig(testIniFiles[i])
		expected := wanted[i]
		if config.Host != expected.Host ||
			config.User != expected.User ||
			config.Password != expected.Password {
			t.Errorf("mismatched data for %v (host: %v:%v, user: %v, password: %q, expected: host: %v:%v, user: %v, password: %q)",
				testIniFiles[i],
				config.Host,
				config.Port,
				config.User,
				config.Password,
				expected.Host,
				expected.Port,
				expected.User,
				expected.Password,
			)
		}
	}
}
