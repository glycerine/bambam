package main

import "go/ast"

// recursively extract the go type as a string
func GetTypeAsString(ty ast.Expr, sofar string) (string, string) {
	switch ty.(type) {

	case (*ast.StarExpr):
		return GetTypeAsString(ty.(*ast.StarExpr).X, sofar+"*")

	case (*ast.Ident):
		return sofar, ty.(*ast.Ident).Name

	case (*ast.ArrayType):
		// slice or array
		return GetTypeAsString(ty.(*ast.ArrayType).Elt, sofar+"[]")
	}

	return sofar, ""
}
