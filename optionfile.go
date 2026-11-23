package mysql_defaults_file

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// optionFile holds an option file's options: group name (lower case)
// to option name to value. Option names use "-" for "_", as the MySQL
// client treats them the same. An option given without "=" has an
// empty value. A later value for an option replaces an earlier one.
type optionFile map[string]map[string]string

// parseOptionFile reads an option file with the rules of the MySQL
// client (mysys/my_default.cc), except that "!include" and
// "!includedir" directives are not followed:
//
//   - blank lines, and lines starting with "#" or ";", are skipped
//   - a line starting with "!" is a directive, and is skipped
//   - "[group]" starts a group; its name is matched without regard to case
//   - an option before any group is an error
//   - "#" outside quotes starts a comment to the end of the line
//   - "name" or "name = value", spaces around each trimmed
//   - a value in matching single or double quotes has them removed
//   - the escapes \n \t \r \b \s \" \' \\ are decoded; others keep the "\"
func parseOptionFile(r io.Reader) (optionFile, error) {
	of := optionFile{}
	var group map[string]string
	scanner := bufio.NewScanner(r)
	for n := 1; scanner.Scan(); n++ {
		line := strings.TrimLeft(scanner.Text(), " \t\r\f\v")
		if line == "" || line[0] == '#' || line[0] == ';' || line[0] == '!' {
			continue
		}
		if line[0] == '[' {
			end := strings.IndexByte(line, ']')
			if end < 0 {
				return nil, fmt.Errorf("line %d: group name without \"]\"", n)
			}
			name := strings.ToLower(strings.TrimRight(line[1:end], " \t"))
			if of[name] == nil {
				of[name] = map[string]string{}
			}
			group = of[name]
			continue
		}
		if group == nil {
			return nil, fmt.Errorf("line %d: option before any group", n)
		}
		line = removeEndComment(line)
		name, value, hasValue := strings.Cut(line, "=")
		name = strings.ReplaceAll(strings.TrimSpace(name), "_", "-")
		if hasValue {
			value = unescape(unquote(strings.TrimSpace(value)))
		}
		group[name] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return of, nil
}

// removeEndComment cuts line at the first "#" outside quotes.
func removeEndComment(line string) string {
	var quote byte
	escape := false
	for i := 0; i < len(line); i++ {
		c := line[i]
		if (c == '\'' || c == '"') && !escape {
			switch quote {
			case 0:
				quote = c
			case c:
				quote = 0
			}
		}
		if quote == 0 && c == '#' {
			return line[:i]
		}
		escape = quote != 0 && c == '\\' && !escape
	}
	return line
}

// unquote removes one pair of matching quotes around s.
func unquote(s string) string {
	if len(s) >= 2 && (s[0] == '\'' || s[0] == '"') && s[0] == s[len(s)-1] {
		return s[1 : len(s)-1]
	}
	return s
}

// unescape decodes the escapes the MySQL client decodes in values.
func unescape(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i == len(s)-1 {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case 'r':
			b.WriteByte('\r')
		case 'b':
			b.WriteByte('\b')
		case 's':
			b.WriteByte(' ')
		case '"', '\'', '\\':
			b.WriteByte(s[i])
		default:
			b.WriteByte('\\')
			b.WriteByte(s[i])
		}
	}
	return b.String()
}
