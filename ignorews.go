package main

import "fmt"

// compare strings, ignoring spaces

const debugEIS = false

func equalIgnoringSpaces(a, b string) (res bool) {

	defer func() {
		if res {
			if debugEIS {
				fmt.Printf("\n\n equalIgnoringSpaces says these are the same: '%s' and '%s'\n", a, b)
			}
		} else {
			fmt.Printf("\n\n equalIgnoringSpaces says these are NOT the same: '%s' and '%s'\n", a, b)
		}
	}()

	r := []rune(a)
	s := []rune(b)
	nextr := 0
	nexts := 0

	for {
		// skip past spaces in both r and s
		for nextr < len(r) {
			if r[nextr] == ' ' {
				nextr++
			} else {
				break
			}
		}

		for nexts < len(s) {
			if s[nexts] == ' ' {
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
