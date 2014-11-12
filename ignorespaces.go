package main

import (
	"fmt"
	"reflect"
)

const (
	success             = ""
	needExactValues     = "This assertion requires exactly %d comparison values (you provided %d)."
	shouldMatchModulo   = "Expected expected string '%s'\n       and actual string '%s'\n to match (ignoring %s)\n (but they did not!; first diff at '%s', pos %d)"
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

	equal, vpos, _ := stringsEqualIgnoring(value, expec, ignoring)
	if equal {
		return success
	} else {
		// extract the string fragment at the differnce point to make it easier to diagnose
		diffpoint := ""
		const diffMax = 20
		vrune := []rune(value)
		n := len(vrune) - vpos + 1
		if n > diffMax {
			n = diffMax
		}
		diffpoint = string(vrune[vpos-1 : (vpos - 1 + n)])

		ignored := "{"
		switch len(ignoring) {
		case 0:
			return fmt.Sprintf(shouldMatchModulo, expec, value, "nothing", diffpoint, vpos-1)
		case 1:
			for k := range ignoring {
				ignored = ignored + fmt.Sprintf("'%c'", k)
			}
			ignored = ignored + "}"
			return fmt.Sprintf(shouldMatchModulo, expec, value, ignored, diffpoint, vpos-1)

		default:
			for k := range ignoring {
				ignored = ignored + fmt.Sprintf("'%c', ", k)
			}
			ignored = ignored + "}"
			return fmt.Sprintf(shouldMatchModulo, expec, value, ignored, diffpoint, vpos-1)
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

func ShouldStartWithModuloWhiteSpace(actual interface{}, expectedPrefix ...interface{}) string {
	if fail := need(1, expectedPrefix); fail != success {
		return fail
	}

	ignoring := map[rune]bool{' ': true, '\n': true, '\t': true}

	value, valueIsString := actual.(string)
	expecPrefix, expecIsString := expectedPrefix[0].(string)

	if !valueIsString || !expecIsString {
		return fmt.Sprintf(shouldBothBeStrings, reflect.TypeOf(actual), reflect.TypeOf(expectedPrefix[0]))
	}

	equal, vpos, _ := hasPrefixEqualIgnoring(value, expecPrefix, ignoring)
	if equal {
		return success
	} else {
		diffpoint := ""
		const diffMax = 20
		vrune := []rune(value)
		n := len(vrune) - vpos + 1
		if n > diffMax {
			n = diffMax
		}
		diffpoint = string(vrune[vpos-1 : (vpos - 1 + n)])

		ignored := "{"
		switch len(ignoring) {
		case 0:
			return fmt.Sprintf(shouldMatchModulo, expecPrefix, value, "nothing", diffpoint, vpos-1)
		case 1:
			for k := range ignoring {
				ignored = ignored + fmt.Sprintf("'%c'", k)
			}
			ignored = ignored + "}"
			return fmt.Sprintf(shouldMatchModulo, expecPrefix, value, ignored, diffpoint, vpos-1)

		default:
			for k := range ignoring {
				ignored = ignored + fmt.Sprintf("'%c', ", k)
			}
			ignored = ignored + "}"
			return fmt.Sprintf(shouldMatchModulo, expecPrefix, value, ignored, diffpoint, vpos-1)
		}
	}
}

// returns if equal, and if not then rpos and spos hold the position of first mismatch
func stringsEqualIgnoring(a, b string, ignoring map[rune]bool) (equal bool, rpos int, spos int) {
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
			return true, -1, -1 // full match
		}

		if nextr >= len(r) {
			return false, nextr, nexts
		}
		if nexts >= len(s) {
			return false, nextr, nexts
		}

		if r[nextr] != s[nexts] {
			return false, nextr, nexts
		}
		nextr++
		nexts++
	}

	return false, nextr, nexts
}

// returns if equal, and if not then rpos and spos hold the position of first mismatch
func hasPrefixEqualIgnoring(str, prefix string, ignoring map[rune]bool) (equal bool, spos int, rpos int) {
	s := []rune(str)
	r := []rune(prefix)

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
			return true, -1, -1 // full match
		}

		if nextr >= len(r) {
			return true, nexts, nextr // for prefix testing
		}
		if nexts >= len(s) {
			return false, nexts, nextr
		}

		if r[nextr] != s[nexts] {
			return false, nexts, nextr
		}
		nextr++
		nexts++
	}

	return false, nexts, nextr
}

func need(needed int, expected []interface{}) string {
	if len(expected) != needed {
		return fmt.Sprintf(needExactValues, needed, len(expected))
	}
	return success
}
