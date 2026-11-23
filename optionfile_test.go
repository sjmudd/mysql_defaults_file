package mysql_defaults_file

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseOptionFile(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  optionFile
	}{
		{"empty", "", optionFile{}},
		{
			"comments, blank lines and leading space",
			"# comment\n; comment\n\n   \n  [client]\n  user=app\n",
			optionFile{"client": {"user": "app"}},
		},
		{
			"directives are skipped",
			"!include /etc/other.cnf\n!includedir /etc/my.cnf.d\n[client]\nuser=app\n",
			optionFile{"client": {"user": "app"}},
		},
		{
			"group names without regard to case, trailing space removed",
			"[Client ]\nuser=app\n[CLIENT]\nhost=db1\n",
			optionFile{"client": {"user": "app", "host": "db1"}},
		},
		{
			"the last value wins",
			"[client]\nuser=one\nuser=two\n",
			optionFile{"client": {"user": "two"}},
		},
		{
			"groups kept apart",
			"[client]\nport=3306\n[mysqld]\nport=3307\n",
			optionFile{"client": {"port": "3306"}, "mysqld": {"port": "3307"}},
		},
		{
			"bare options have an empty value",
			"[client]\nskip-ssl\nno-auto-rehash  \n",
			optionFile{"client": {"skip-ssl": "", "no-auto-rehash": ""}},
		},
		{
			"underscore is dash",
			"[client]\nssl_mode=REQUIRED\n",
			optionFile{"client": {"ssl-mode": "REQUIRED"}},
		},
		{
			"spaces around name and value",
			"[client]\n  user  =   app   \n",
			optionFile{"client": {"user": "app"}},
		},
		{
			"empty value",
			"[client]\npassword=\n",
			optionFile{"client": {"password": ""}},
		},
		{
			"quotes removed",
			"[client]\na='one two'\nb=\"three\"\nc=  'four'  \n",
			optionFile{"client": {"a": "one two", "b": "three", "c": "four"}},
		},
		{
			"quotes kept when not a matching pair",
			"[client]\na='one\"\nb=\"\nc='x\n",
			optionFile{"client": {"a": "'one\"", "b": "\"", "c": "'x"}},
		},
		{
			"comment at end of line",
			"[client]\nuser=app # the app user\n",
			optionFile{"client": {"user": "app"}},
		},
		{
			"hash inside quotes is not a comment",
			"[client]\npassword='a#b' # comment\n",
			optionFile{"client": {"password": "a#b"}},
		},
		{
			"escaped quote inside quotes",
			"[client]\npassword='a\\'#b' # comment\n",
			optionFile{"client": {"password": "a'#b"}},
		},
		{
			"escapes",
			"[client]\na=x\\ny\\tz\nb=x\\sy\nc=x\\\\y\nd=x\\\"y\ne=x\\qy\nf=x\\\n",
			optionFile{"client": {
				"a": "x\ny\tz",
				"b": "x y",
				"c": `x\y`,
				"d": `x"y`,
				"e": `x\qy`,
				"f": `x\`,
			}},
		},
		{
			"value with an equals sign",
			"[client]\npassword=a=b\n",
			optionFile{"client": {"password": "a=b"}},
		},
		{
			"CRLF line endings",
			"[client]\r\nuser=app\r\n",
			optionFile{"client": {"user": "app"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOptionFile(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseOptionFileErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"group without ]", "[client\nuser=app\n", "line 1: group name without \"]\""},
		{"option before any group", "# comment\nuser=app\n", "line 2: option before any group"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseOptionFile(strings.NewReader(tt.input))
			if err == nil || err.Error() != tt.want {
				t.Errorf("got error %v, want %q", err, tt.want)
			}
		})
	}
}
