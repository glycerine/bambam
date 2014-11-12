package main

import (
	"fmt"
	"reflect"
)

const (
	success             = ""
	needExactValues     = "This assertion requires exactly %d comparison values (you provided %d)."
	shouldMatchModulo   = "Expected expected string '%s'\n       and actual string '%s'\n to match (ignoring %s)\n (but they did not!)"
	shouldBothBeStrings = "Both arguments to this assertion must be strings (you provided %v and %v)."
)

// ShouldMatchModulo receives exactly two string parameters and an ignore map. It ensures that the order
// of not-ignored characters in the two strings is identical. Runes specified in the ignore map
// are ignored for the purposes of this string comparison, and each should map to true.
// ShouldMatchModulo thus allows you to do whitespace insensitive comparison, which is very useful
// in lexer/parser work.
//
func ShouldMatchModulo(ignoring map[rune]bool, actual interface{}, expected ...interface{}) string {
	if fail := need(1, expected); fail != success {
		return fail
	}

	value, valueIsString := actual.(string)
	expec, expecIsString := expected[0].(string)

	if !valueIsString || !expecIsString {
		return fmt.Sprintf(shouldBothBeStrings, reflect.TypeOf(actual), reflect.TypeOf(expected[0]))
	}

	if stringsEqualIgnoring(value, expec, ignoring) {
		return success
	} else {
		ignored := "{"
		switch len(ignoring) {
		case 0:
			return fmt.Sprintf(shouldMatchModulo, expec, value, "nothing")
		case 1:
			for k := range ignoring {
				ignored = ignored + fmt.Sprintf("'%c'", k)
			}
			ignored = ignored + "}"
			return fmt.Sprintf(shouldMatchModulo, expec, value, ignored)

		default:
			for k := range ignoring {
				ignored = ignored + fmt.Sprintf("'%c', ", k)
			}
			ignored = ignored + "}"
			return fmt.Sprintf(shouldMatchModulo, expec, value, ignored)
		}
	}
}

// ShouldMatchModuloSpaces compares two strings but ignores ' ' spaces.
// Serves as an example of use of ShouldMatchModulo.
//
func ShouldMatchModuloSpaces(actual interface{}, expected ...interface{}) string {
	if fail := need(1, expected); fail != success {
		return fail
	}
	return ShouldMatchModulo(map[rune]bool{' ': true}, actual, expected[0])
}

func ShouldMatchModuloWhiteSpace(actual interface{}, expected ...interface{}) string {
	if fail := need(1, expected); fail != success {
		return fail
	}
	return ShouldMatchModulo(map[rune]bool{' ': true, '\n': true, '\t': true}, actual, expected[0])
}

func stringsEqualIgnoring(a, b string, ignoring map[rune]bool) bool {
	r := []rune(a)
	s := []rune(b)

	nextr := 0
	nexts := 0

	for {
		// skip past spaces in both r and s
		for nextr < len(r) {
			if ignoring[r[nextr]] {
				nextr++
			} else {
				break
			}
		}

		for nexts < len(s) {
			if ignoring[s[nexts]] {
				nexts++
			} else {
				break
			}
		}

		if nextr >= len(r) && nexts >= len(s) {
			return true
		}

		if nextr >= len(r) {
			return false
		}
		if nexts >= len(s) {
			return false
		}

		if r[nextr] != s[nexts] {
			return false
		}
		nextr++
		nexts++
	}

	return false
}

func need(needed int, expected []interface{}) string {
	if len(expected) != needed {
		return fmt.Sprintf(needExactValues, needed, len(expected))
	}
	return success
}
