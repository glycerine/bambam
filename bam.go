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
	os.Stdout.Write(ExtractStructs(readMe, nil))
	fmt.Printf("\n")
}

type Extractor struct {
	fieldCount int
	out        bytes.Buffer
}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func ExtractString(src string) string {
	return string(ExtractStructs("", "package main; "+src))
}

// ExtractStructs pulls out the struct definitions from a golang source file.
//
// src has to be string, []byte, or io.Reader, as in parser.ParseFile(). src
// can be nil if fname is provided. See http://golang.org/pkg/go/parser/#ParseFile
//
func ExtractStructs(fname string, src interface{}) []byte {

	x := NewExtractor()
	x.Init()

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

							switch ty := ts2.Type.(type) {
							default:
								// never hit
								fmt.Printf("\n\n unrecog type ty = %#v\n", ty)
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
														switch fld2.Type.(type) {
														case (*ast.Ident):
															ident2 := fld2.Type.(*ast.Ident)
															x.GenerateStructField(ident.Name, ident2.Name, fld2)
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

func (x *Extractor) Init() {
	//fmt.Fprintf(&x.out, "package main; func main(){println(`extractor output ran fine.`)}\n")
}
func (x *Extractor) StartStruct(name string) {
	x.fieldCount = 0
	cname := CapnpStructName(name)
	fmt.Fprintf(&x.out, "struct %s { ", cname)
}
func (x *Extractor) EndStruct() {
	fmt.Fprintf(&x.out, "} ")
}

func (x *Extractor) GenerateComment(c string) {
	skipCommentsForNow := false
	if !skipCommentsForNow {
		fmt.Fprintf(&x.out, "%s\n", c) // prod
	}
}

func CapnpStructName(name string) string {
	if len(name) == 0 {
		return name
	}

	// gotta upercase the first letter of type (struct) names
	runes := []rune(name)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)

}

func CapnpFieldName(name string) string {
	if len(name) == 0 {
		return name
	}

	// gotta lowercase the first letter of field names
	runes := []rune(name)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func (x *Extractor) GenerateStructField(name string, typeName string, fld *ast.Field) {

	loweredName := CapnpFieldName(name)
	typeDisplayed := typeName

	switch typeName {
	case "string":
		typeDisplayed = "Text"
	case "int":
		typeDisplayed = "Int64"
	case "int64":
		typeDisplayed = "Int64"
	case "float64":
		typeDisplayed = "Float64"
	}

	fmt.Fprintf(&x.out, "%s @%d: %s; ", loweredName, x.fieldCount, typeDisplayed) // prod
	x.fieldCount++
}

func (x *Extractor) GenerateEmbedded(typeName string) {
	fmt.Fprintf(&x.out, "%s; ", typeName) // prod
}

func CapnpCompileFragment(in []byte) error {
	id, err := exec.Command("capnp", "id").CombinedOutput()
	if err != nil {
		panic(err)
	}

	f, err := ioutil.TempFile(".", "capnp.test.")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	defer os.Remove(f.Name())

	var full bytes.Buffer
	fmt.Fprintf(&full, "%s;\n", id)

	return CapnpCompile(full.Bytes())

	_, err = exec.Command("capnp", "compile", "-ogo", f.Name()).CombinedOutput()
	if err != nil {
		panic(err)
	}

	return err
}
func CapnpCompile(in []byte) error {
	//"capnp compile -ogo capiptab.capnp"
	return nil
}
