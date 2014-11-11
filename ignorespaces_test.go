package main

import "testing"

func TestShouldMatchModuloSpaces(t *testing.T) {
	pass(t, so("", ShouldMatchModuloSpaces, ""))
	pass(t, so("a", ShouldMatchModuloSpaces, "a"))
	pass(t, so(" a ", ShouldMatchModuloSpaces, "a"))
	pass(t, so("   a   ", ShouldMatchModuloSpaces, "a"))
	pass(t, so("a", ShouldMatchModuloSpaces, " a "))
	pass(t, so("a", ShouldMatchModuloSpaces, "   a   "))
	pass(t, so("a b c de fgh   ij", ShouldMatchModuloSpaces, "  abcdefghi    j "))

	fail(t, so("j  ", ShouldMatchModuloSpaces, "j k"), "Expected expected string 'j k' and actual string 'j ' to match (ignoring {' '}) (but they did not!)")

	fail(t, so("asdf", ShouldMatchModuloSpaces), "This assertion requires exactly 1 comparison values (you provided 0).")
	fail(t, so("asdf", ShouldMatchModuloSpaces, 1, 2, 3), "This assertion requires exactly 1 comparison values (you provided 3).")

	fail(t, so(123, ShouldMatchModuloSpaces, 23), "Both arguments to this assertion must be strings (you provided int and int).")

}
