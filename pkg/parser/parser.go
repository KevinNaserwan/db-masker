package parser

import (
	"strings"
)

// ParseSQLValuesTuples parses standard SQL VALUE tuples (e.g. (1, 'John', NULL), (2, 'Doe'))
func ParseSQLValuesTuples(valuesStr string) [][]string {
	var tuples [][]string
	var currentTuple []string
	var currentVal strings.Builder
	inString := false
	escapeNext := false
	depth := 0

	runes := []rune(valuesStr)
	length := len(runes)

	for i := 0; i < length; i++ {
		char := runes[i]

		if escapeNext {
			currentVal.WriteRune(char)
			escapeNext = false
			continue
		}

		if char == '\\' && inString {
			currentVal.WriteRune(char)
			escapeNext = true
			continue
		}

		if char == '\'' {
			if inString && i+1 < length && runes[i+1] == '\'' {
				currentVal.WriteString("''")
				i++
				continue
			}
			inString = !inString
			currentVal.WriteRune(char)
			continue
		}

		if !inString {
			if char == '(' {
				depth++
				if depth == 1 {
					currentTuple = nil
					currentVal.Reset()
					continue
				}
			} else if char == ')' {
				depth--
				if depth == 0 {
					currentTuple = append(currentTuple, strings.TrimSpace(currentVal.String()))
					currentVal.Reset()
					tuples = append(tuples, currentTuple)
					continue
				}
			} else if char == ',' && depth == 1 {
				currentTuple = append(currentTuple, strings.TrimSpace(currentVal.String()))
				currentVal.Reset()
				continue
			}
		}

		if depth > 0 {
			currentVal.WriteRune(char)
		}
	}

	return tuples
}

// SerializeSQLTuples re-encodes parsed tuples into standard SQL syntax
func SerializeSQLTuples(tuples [][]string) string {
	var parts []string
	for _, tuple := range tuples {
		parts = append(parts, "("+strings.Join(tuple, ", ")+")")
	}
	return strings.Join(parts, ", ")
}
