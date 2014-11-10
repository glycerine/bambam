package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

func ParseCmdLine() string {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "bambam needs exactly one golang source file to process.\n")
		os.Exit(1)
	}
	fn := os.Args[1]
	if !strings.HasSuffix(fn, ".go") {
		fmt.Fprintf(os.Stderr, "error: bambam input file '%s' did not end in '.go'.\n")
		os.Exit(1)
	}
	return fn
}

func main() {
	readMe := ParseCmdLine()

	x := NewExtractor()
	x.fieldPrefix = "   "
	x.fieldSuffix = "\n"

	by := x.AssembleCapnpFile(ExtractStructs(readMe, nil, x))
	os.Stdout.Write(by.Bytes())
	fmt.Printf("\n")
	fmt.Printf("##compile with:\n\n##   capnp compile -ogo yourfile.capnp\n\n")
}

type Extractor struct {
	fieldCount  int
	out         bytes.Buffer
	pkgName     string
	importDecl  string
	fieldPrefix string
	fieldSuffix string

	// for testing purposes
	myCounts *Extractor
	myNames  []string
}

func NewExtractor() *Extractor {
	return &Extractor{
		pkgName:    "testpkg",
		importDecl: "testpkg",
	}
}

func ExtractFromString(src string) []byte {
	return ExtractStructs("", "package main; "+src, nil)
}

func ExtractString2String(src string) string {
	return string(ExtractStructs("", "package main; "+src, nil))
}

// ExtractStructs pulls out the struct definitions from a golang source file.
//
// src has to be string, []byte, or io.Reader, as in parser.ParseFile(). src
// can be nil if fname is provided. See http://golang.org/pkg/go/parser/#ParseFile
//
func ExtractStructs(fname string, src interface{}, x *Extractor) []byte {

	if x == nil {
		x = NewExtractor()
	}

	fset := token.NewFileSet() // positions are relative to fset

	f, err := parser.ParseFile(fset, fname, src, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	//	fmt.Printf("parsed output f.Decls is:\n")
	//	goon.Dump(f.Decls)

	//fmt.Printf("len(f.Decls) = %d\n", len(f.Decls))

	for _, v := range f.Decls {
		switch v.(type) {
		case *ast.GenDecl:
			d := v.(*ast.GenDecl)
			//fmt.Printf("dump of d, an %#v = \n", ty)
			//goon.Dump(d)

			//fmt.Printf("\n\n\n  detail dump of d.Specs\n")
			for _, spe := range d.Specs {
				switch spe.(type) {
				case (*ast.TypeSpec):

					// go back and print the comments
					if d.Doc != nil && d.Doc.List != nil && len(d.Doc.List) > 0 {
						for _, com := range d.Doc.List {
							x.GenerateComment(com.Text)
						}
					}

					typeSpec := spe.(*ast.TypeSpec)
					//fmt.Printf("\n\n *ast.TypeSpec spe = \n")

					if typeSpec.Name.Obj.Kind == ast.Typ {

						switch typeSpec.Name.Obj.Decl.(type) {
						case (*ast.TypeSpec):
							//curStructName := typeSpec.Name.String()
							curStructName := typeSpec.Name.Name
							ts2 := typeSpec.Name.Obj.Decl.(*ast.TypeSpec)

							//fmt.Printf("\n\n  in ts2 = %#v\n", ts2)
							//goon.Dump(ts2)

							switch ts2.Type.(type) {
							default:
								// *ast.InterfaceType and *ast.Ident end up here.
								//fmt.Printf("\n\n unrecog type ty = %#v\n", ty)
							case (*ast.StructType):
								stru := ts2.Type.(*ast.StructType)

								x.StartStruct(curStructName)
								//fmt.Printf("\n\n stru = %#v\n", stru)
								//goon.Dump(stru)

								if stru.Fields != nil {
									for _, fld := range stru.Fields.List {
										if fld != nil {
											//fmt.Printf("\n\n    fld.Names = %#v\n", fld.Names) // looking for
											//goon.Dump(fld.Names)

											if len(fld.Names) == 0 {
												// field without name: embedded/anonymous struct
												x.GenerateEmbedded(fld.Type.(*ast.Ident).Name)

											} else {
												// field with name
												for _, ident := range fld.Names {

													switch ident.Obj.Decl.(type) {
													case (*ast.Field):
														// named field
														fld2 := ident.Obj.Decl.(*ast.Field)

														//fmt.Printf("\n\n    fld2 = %#v\n", fld2)
														//goon.Dump(fld2)

														switch fld2.Type.(type) {

														case (*ast.StarExpr):
															star2 := fld2.Type.(*ast.StarExpr)
															x.GenerateStructField(ident.Name, star2.X.(*ast.Ident).Name, fld2, NotList)

														case (*ast.Ident):
															ident2 := fld2.Type.(*ast.Ident)
															x.GenerateStructField(ident.Name, ident2.Name, fld2, NotList)
														case (*ast.ArrayType):
															// slice or array
															array2 := fld2.Type.(*ast.ArrayType)
															switch array2.Elt.(type) {
															case (*ast.Ident):
																x.GenerateStructField(ident.Name, array2.Elt.(*ast.Ident).Name, fld2, YesIsList)
															case (*ast.StarExpr):
																x.GenerateStructField(ident.Name, array2.Elt.(*ast.StarExpr).X.(*ast.Ident).Name, fld2, YesIsList)
															}
														}
													}
												}
											}

										}
									}
								}

								//fmt.Printf("} // end of %s \n\n", typeSpec.Name) // prod
								x.EndStruct()

								//goon.Dump(stru)
								//fmt.Printf("\n =========== end stru =======\n\n\n")
							}
						}

						//fmt.Printf("spe.Name.Obj.Kind = %s\n", typeSpec.Name.Obj.Kind)

						//fmt.Printf("spe.Name.Obj = %#v\n", typeSpec.Name.Obj)
						//goon.Dump(typeSpec.Name.Obj)

						//fmt.Printf("spe.Name.Obj.Decl = %#v\n", typeSpec.Name.Obj.Decl)
						//goon.Dump(typeSpec.Name.Obj.Decl)
					}

				}
			}
		}
	}

	return x.out.Bytes()
}

func (x *Extractor) StartStruct(name string) {
	x.fieldCount = 0
	cname := UppercaseCapnpTypeName(name)
	fmt.Fprintf(&x.out, "struct %s { %s", cname, x.fieldSuffix)
}
func (x *Extractor) EndStruct() {
	fmt.Fprintf(&x.out, "} %s", x.fieldSuffix)
}

func (x *Extractor) GenerateComment(c string) {
	skipCommentsForNow := false
	if !skipCommentsForNow {
		fmt.Fprintf(&x.out, "%s\n", c) // prod
	}
}

func UppercaseCapnpTypeName(name string) string {
	if len(name) == 0 {
		return name
	}

	// gotta upercase the first letter of type (struct) names
	runes := []rune(name)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)

}

func LowercaseCapnpFieldName(name string) string {
	if len(name) == 0 {
		return name
	}

	// gotta lowercase the first letter of field names
	runes := []rune(name)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

const YesIsList = true
const NotList = false

func (x *Extractor) GenerateStructField(name string, typeName string, fld *ast.Field, isList bool) {

	loweredName := LowercaseCapnpFieldName(name)
	typeDisplayed := UppercaseCapnpTypeName(typeName)

	switch typeName {
	case "string":
		typeDisplayed = "Text"
	case "int":
		typeDisplayed = "Int64"
	case "bool":
		typeDisplayed = "Bool"
	case "int8":
		typeDisplayed = "Int8"
	case "int16":
		typeDisplayed = "Int16"
	case "int32":
		typeDisplayed = "Int32"
	case "int64":
		typeDisplayed = "Int64"
	case "uint8":
		typeDisplayed = "UInt8"
	case "uint16":
		typeDisplayed = "UInt16"
	case "uint32":
		typeDisplayed = "UInt32"
	case "uint64":
		typeDisplayed = "UInt64"
	case "float32":
		typeDisplayed = "Float32"
	case "float64":
		typeDisplayed = "Float64"
	case "byte":
		if isList {
			typeDisplayed = "Data"
			isList = false
		} else {
			typeDisplayed = "Uint8"
		}
	}

	if isList {
		fmt.Fprintf(&x.out, "%s%s @%d: List(%s); %s", x.fieldPrefix, loweredName, x.fieldCount, typeDisplayed, x.fieldSuffix)
	} else {
		fmt.Fprintf(&x.out, "%s%s @%d: %s; %s", x.fieldPrefix, loweredName, x.fieldCount, typeDisplayed, x.fieldSuffix)
	}
	x.fieldCount++
}

func (x *Extractor) GenerateEmbedded(typeName string) {
	fmt.Fprintf(&x.out, "%s; ", typeName) // prod
}

func getNewCapnpId() string {
	id, err := exec.Command("capnp", "id").CombinedOutput()
	if err != nil {
		panic(err)
	}
	n := len(id)
	if n > 0 {
		id = id[:n-1]
	}

	return string(id)
}

func (x *Extractor) AssembleCapnpFile(in []byte) *bytes.Buffer {
	var by bytes.Buffer

	id := getNewCapnpId()

	fmt.Fprintf(&by, `%s;
using Go = import "go.capnp";
$Go.package("%s");
$Go.import("%s");
%s`, id, x.pkgName, x.importDecl, x.fieldSuffix)
	by.Write(in)
	fmt.Fprintf(&by, "\n")

	return &by
}

func CapnpCompileFragment(in []byte) []byte {
	x := NewExtractor()
	return x.CapnpCompileFragment(in)
}

func (x *Extractor) CapnpCompileFragment(in []byte) []byte {

	f, err := ioutil.TempFile(".", "capnp.test.")
	if err != nil {
		panic(err)
	}
	defer os.Remove(f.Name())

	by := x.AssembleCapnpFile(in)
	debug := string(by.Bytes())

	f.Write(by.Bytes())
	f.Close()

	compiled, err := CapnpCompilePath(f.Name())
	if err != nil {
		return []byte(fmt.Sprintf("error compiling the generated capnp code: '%s'; error: '%s'\n", debug, err) + string(compiled))
	}

	return compiled
}

func CapnpCompilePath(fname string) ([]byte, error) {
	defer os.Remove(fname + ".go")

	by, err := exec.Command("capnp", "compile", "-ogo", fname).CombinedOutput()
	if err != nil {
		return by, err
	}
	return by, nil
}
