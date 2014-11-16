package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type Field struct {
	capname           string
	GoCapGoName       string // Uppercased-first-letter of capname, as generated in go bindings.
	capType           string
	goName            string
	goType            string
	goTypePrefix      string
	tagValue          string
	isList            bool
	capIdFromTag      int
	orderOfAppearance int
	finalOrder        int
	embedded          bool
}

type Struct struct {
	capName              string
	goName               string
	fld                  []*Field
	longestField         int
	comment              string
	capIdMap             map[int]*Field
	firstNonTextListSeen bool
	listNum              int
}

func (s *Struct) computeFinalOrder() {

	// assign Field.finalOrder to all values in s.fld, from 0..(len(s.fld)-1)
	max := len(s.fld) - 1
	// check for out of bounds requests
	for _, f := range s.fld {
		if f.capIdFromTag > max {
			err := fmt.Errorf(`problem in capid tag '%s' on field '%s' in struct '%s': number '%d' is beyond the count of fields we have, largest available is %d`, f.tagValue, f.goName, s.goName, f.capIdFromTag, max)
			panic(err)
		}
	}

	// only one field? already done
	if len(s.fld) == 1 {
		s.fld[0].finalOrder = 0
		return
	}
	// INVAR: no duplicate requests, and all are in bounds.

	// wipe slate clean
	for _, f := range s.fld {
		f.finalOrder = -1
	}

	final := make([]*Field, len(s.fld))

	// assign from map
	for _, v := range s.capIdMap {
		v.finalOrder = v.capIdFromTag
		final[v.capIdFromTag] = v
	}

	appear := make([]*Field, len(s.fld))
	copy(appear, s.fld)

	sort.Sort(ByOrderOfAppearance(appear))

	//find next available slot, and fill in, in order of appearance
	write := 0

	for read := 0; read <= max; read++ {
		if appear[read].finalOrder != -1 {
			continue
		}
		// appear[read] needs an assignment

		done := advanceWrite(&write, final)
		if !done {
			// final[write] has a slot for an assignment
			final[write] = appear[read]
			final[write].finalOrder = write
		}

	}
}

// returns true if done. if false, then final[*pw] is available for writing.
func advanceWrite(pw *int, final []*Field) bool {
	for n := len(final); *pw < n; (*pw)++ {
		if final[*pw] == nil { // .finalOrder == -1 {
			return false
		}
	}
	return true
}

func NewStruct(capName, goName string) *Struct {
	return &Struct{
		capName:  capName,
		goName:   goName,
		fld:      []*Field{},
		capIdMap: map[int]*Field{},
	}
}

type Extractor struct {
	fieldCount  int
	out         bytes.Buffer
	pkgName     string
	importDecl  string
	fieldPrefix string
	fieldSuffix string

	curStruct      *Struct
	heldComment    string
	extractPrivate bool

	// map structs' goName <-> capName
	goType2capType map[string]string
	capType2goType map[string]string

	// key is goName
	srs        map[string]*Struct
	ToGoCode   map[string][]byte
	ToCapnCode map[string][]byte
	SaveCode   map[string][]byte
	LoadCode   map[string][]byte

	compileDir *TempDir
}

func NewExtractor() *Extractor {
	return &Extractor{
		pkgName:        "testpkg",
		importDecl:     "testpkg",
		goType2capType: make(map[string]string),
		capType2goType: make(map[string]string),

		// key is goTypeName
		ToGoCode:   make(map[string][]byte),
		ToCapnCode: make(map[string][]byte),
		SaveCode:   make(map[string][]byte),
		LoadCode:   make(map[string][]byte),
		srs:        make(map[string]*Struct),
		compileDir: NewTempDir(),
	}
}

func (x *Extractor) Cleanup() {
	if x.compileDir != nil {
		x.compileDir.Cleanup()
	}
}

func (x *Extractor) GenerateTranslators() {

	for _, s := range x.srs {

		x.SaveCode[s.goName] = []byte(fmt.Sprintf(`
func (s *%s) Save(w io.Writer) {
  	seg := capn.NewBuffer(nil)
  	%sGoToCapn(seg, s)
  	seg.WriteTo(w)
}
 `, s.goName, s.goName))

		x.LoadCode[s.goName] = []byte(fmt.Sprintf(` 
func (s *%s) Load(r io.Reader) {
  	capMsg, err := capn.ReadFromStream(r, nil)
  	if err != nil {
  		panic(fmt.Errorf("capn.ReadFromStream error: %%s", err))
  	}
  	z := %sReadRoot%s(capMsg)
      %sToGo(z, s)
}
`, s.goName, x.packageDot(), s.capName, s.capName))

		x.ToGoCode[s.goName] = []byte(fmt.Sprintf(`
func %sToGo(src %s, dest *%s) *%s { 
  if dest == nil { 
    dest = &%s{} 
  }
%s
  return dest
} 
`, s.capName, s.capName, s.goName, s.goName, s.goName, x.SettersToGo(s.goName)))

		x.ToCapnCode[s.goName] = []byte(fmt.Sprintf(`
func %sGoToCapn(seg *capn.Segment, src *%s) %s { 
  dest := AutoNew%s(seg)
%s
  return dest
} 
`, s.goName, s.goName, s.capName, s.capName, x.SettersToCapn(s.goName)))

	}
}

func (x *Extractor) packageDot() string {
	if x.pkgName == "" || x.pkgName == "main" {
		return ""
	}
	return x.pkgName + "."
}

func (x *Extractor) SettersToGo(goName string) string {
	var buf bytes.Buffer
	myStruct := x.srs[goName]
	if myStruct == nil {
		panic(fmt.Sprintf("bad goName '%s'", goName))
	}
	//fmt.Printf("\n\n SettersToGo running on myStruct = %#v\n", myStruct)
	//for i, f := range myStruct.fld {
	//fmt.Printf("\n\n SettersToGo running on myStruct.fld[%d] = %#v\n", i, f)
	i := 0
	for _, f := range myStruct.fld {

		if f.isList {
			x.SettersToGoListHelper(&buf, myStruct, f)
		} else {
			switch f.goType {
			case "int":
				fmt.Fprintf(&buf, "  dest.%s = int(src.%s())\n", f.goName, f.GoCapGoName)
			case "int64":
				fmt.Fprintf(&buf, "  dest.%s = int64(src.%s())\n", f.goName, f.GoCapGoName)
			case "float64":
				fmt.Fprintf(&buf, "  dest.%s = float64(src.%s())\n", f.goName, f.GoCapGoName)
			case "string":
				fmt.Fprintf(&buf, "  dest.%s = src.%s()\n", f.goName, f.GoCapGoName)
			}
		}
		i++
	}
	return string(buf.Bytes())
}

func (x *Extractor) SettersToGoListHelper(buf io.Writer, myStruct *Struct, f *Field) {

	//fmt.Printf("debug: field f = %#v\n", f)

	// special case Text / string slices
	if f.capType == "Text" {
		fmt.Fprintf(buf, "  dest.%s = src.%s().ToArray()\n", f.goName, f.GoCapGoName)
		return
	}
	if !myStruct.firstNonTextListSeen {
		fmt.Fprintf(buf, "\n    var n int\n")
		myStruct.firstNonTextListSeen = true
	}
	// add a dereference (*) in from of the ToGo() invocation for go types that aren't pointers.
	addStar := "*"
	if isPointerType(f.goTypePrefix) {
		addStar = ""
	}

	fmt.Fprintf(buf, `
    // %s
	n = src.%s().Len()
	dest.%s = make(%s%s, n)
	for i := 0; i < n; i++ {
        dest.%s[i] = %s
    }

`, f.goName, f.GoCapGoName, f.goName, f.goTypePrefix, f.goType, f.goName, ElemStarCapToGo(addStar, f))

}

func ElemStarCapToGo(addStar string, f *Field) string {
	fmt.Printf("f = %#v   addStar = %v\n", f, addStar)
	if IsIntrinsicGoType(f.goType) {
		fmt.Printf("\n intrinsic detected.\n")
		return fmt.Sprintf("%s(src.%s().At(i))", f.goType, f.goName)
	} else {
		fmt.Printf("\n non-intrinsic detected.\n")
		return fmt.Sprintf("%s%sToGo(src.%s().At(i), nil)", addStar, f.capType, f.goName)
	}
}

func isPointerType(goTypePrefix string) bool {
	if len(goTypePrefix) == 0 {
		return false
	}
	prefix := []rune(goTypePrefix)
	if prefix[len(prefix)-1] == '*' {
		return true
	}
	return false
}

func (x *Extractor) SettersToCapn(goName string) string {
	var buf bytes.Buffer
	t := x.srs[goName]
	if t == nil {
		panic(fmt.Sprintf("bad goName '%s'", goName))
	}
	//fmt.Printf("\n\n SettersToCapn running on myStruct = %#v\n", t)
	//for i, f := range t.fld {
	for _, f := range t.fld {
		//fmt.Printf("\n\n SettersToCapn running on t.fld[%d] = %#v\n", i, f)

		if f.isList {
			t.listNum++
			switch f.goType {
			case "int":
				fallthrough
			case "int64":
				fmt.Fprintf(&buf, `

  mylist%d := seg.NewInt64List(len(src.%s))
  for i:=0; i < len(src.%s); i++ {
     mylist%d.Set(i, int64(src.%s[i]))
  }
  dest.Set%s(mylist%d)
`, t.listNum, f.goName, f.goName, t.listNum, f.goName, f.GoCapGoName, t.listNum)
			case "float64":
				fmt.Fprintf(&buf, `

  mylist%d := seg.NewFloat64List(len(src.%s))
  for i:=0; i < len(src.%s); i++ {
     mylist%d.Set(i, src.%s[i])
  }
  dest.Set%s(mylist%d)
`, t.listNum, f.goName, f.goName, t.listNum, f.goName, f.GoCapGoName, t.listNum)
			case "string":
				fmt.Fprintf(&buf, `
  // text list
  tl%d := seg.NewTextList(len(src.%s))
  for i := range src.%s {
     tl%d.Set(i, src.%s[i])
  }
  dest.Set%s(tl%d)
`, t.listNum, f.goName, f.goName, t.listNum, f.goName, f.GoCapGoName, t.listNum)

			default:
				// handle list of struct
				//fmt.Printf("\n\n  at struct list in SettersToCap(): f = %#v\n", f)
				addAmpersand := "&"
				if isPointerType(f.goTypePrefix) {
					addAmpersand = ""
				}

				fmt.Fprintf(&buf, `
  // %s -> %s (go slice to capn list)
  if len(src.%s) > 0 {
		typedList := New%sList(seg, len(src.%s))
		plist := capn.PointerList(typedList)
		i := 0
		for _, ele := range src.%s {
			plist.Set(i, capn.Object(%sGoToCapn(seg, %sele)))
			i++
		}
		dest.Set%s(typedList)
	}
`, f.goName, f.capType, f.goName, f.capType, f.goName, f.goName, f.goType, addAmpersand, f.GoCapGoName)

			} // end switch f.goType

		} else {

			switch f.goType {
			case "int":
				fmt.Fprintf(&buf, "  dest.Set%s(int64(src.%s))\n", f.GoCapGoName, f.goName)
			case "int64":
				fmt.Fprintf(&buf, "  dest.Set%s(src.%s)\n", f.GoCapGoName, f.goName)
			case "float64":
				fmt.Fprintf(&buf, "  dest.Set%s(src.%s)\n", f.GoCapGoName, f.goName)
			case "string":
				fmt.Fprintf(&buf, "  dest.Set%s(src.%s)\n", f.GoCapGoName, f.goName)
			}
		}
	}
	return string(buf.Bytes())
}

func (x *Extractor) ToGoCodeFor(goName string) []byte {
	return x.ToGoCode[goName]
}

func (x *Extractor) ToCapnCodeFor(goStructName string) []byte {
	return x.ToCapnCode[goStructName]
}

type ByFinalOrder []*Field

func (s ByFinalOrder) Len() int {
	return len(s)
}
func (s ByFinalOrder) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByFinalOrder) Less(i, j int) bool {
	return s[i].finalOrder < s[j].finalOrder
}

type ByOrderOfAppearance []*Field

func (s ByOrderOfAppearance) Len() int {
	return len(s)
}
func (s ByOrderOfAppearance) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByOrderOfAppearance) Less(i, j int) bool {
	return s[i].orderOfAppearance < s[j].orderOfAppearance
}

type ByGoName []*Struct

func (s ByGoName) Len() int {
	return len(s)
}
func (s ByGoName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByGoName) Less(i, j int) bool {
	return s[i].goName < s[j].goName
}

func (x *Extractor) WriteToSchema(w io.Writer) (n int64, err error) {

	var m int
	var spaces string

	// sort structs alphabetically to get a stable (testable) ordering.
	sortedStructs := ByGoName(make([]*Struct, 0, len(x.srs)))
	for _, strct := range x.srs {
		sortedStructs = append(sortedStructs, strct)
	}
	sort.Sort(ByGoName(sortedStructs))

	for _, s := range sortedStructs {

		m, err = fmt.Fprintf(w, "%sstruct %s { %s", x.fieldSuffix, s.capName, x.fieldSuffix)
		n += int64(m)
		if err != nil {
			return
		}

		s.computeFinalOrder()

		sort.Sort(ByFinalOrder(s.fld))

		for i, fld := range s.fld {
			SetSpaces(&spaces, s.longestField, len(fld.capname))

			capType, already := x.goType2capType[fld.goType]
			if !already {
				capType = fld.capType
			}

			if fld.isList {
				m, err = fmt.Fprintf(w, "%s%s  %s@%d: %sList(%s); %s", x.fieldPrefix, fld.capname, spaces, fld.finalOrder, ExtraSpaces(i), capType, x.fieldSuffix)
				n += int64(m)
				if err != nil {
					return
				}

			} else {
				m, err = fmt.Fprintf(w, "%s%s  %s@%d: %s%s; %s", x.fieldPrefix, fld.capname, spaces, fld.finalOrder, ExtraSpaces(i), capType, x.fieldSuffix)
				n += int64(m)
				if err != nil {
					return
				}

			}
		} // end field loop

		m, err = fmt.Fprintf(w, "} %s", x.fieldSuffix)
		n += int64(m)
		if err != nil {
			return
		}

	} // end loop over structs

	return
}

func (x *Extractor) WriteToTranslators(w io.Writer) (n int64, err error) {

	var m int

	x.GenerateTranslators()

	// sort structs alphabetically to get a stable (testable) ordering.
	sortedStructs := ByGoName(make([]*Struct, 0, len(x.srs)))
	for _, strct := range x.srs {
		sortedStructs = append(sortedStructs, strct)
	}
	sort.Sort(ByGoName(sortedStructs))

	// now print the translating methods, in a second pass over structures, to accomodate
	// our test structure
	for _, s := range sortedStructs {

		m, err = fmt.Fprintf(w, "\n\n")
		n += int64(m)
		if err != nil {
			return
		}

		m, err = w.Write(x.SaveCode[s.goName])
		n += int64(m)
		if err != nil {
			return
		}

		m, err = fmt.Fprintf(w, "\n\n")
		n += int64(m)
		if err != nil {
			return
		}

		m, err = w.Write(x.LoadCode[s.goName])
		n += int64(m)
		if err != nil {
			return
		}

		m, err = fmt.Fprintf(w, "\n\n")
		n += int64(m)
		if err != nil {
			return
		}

		m, err = w.Write(x.ToGoCodeFor(s.goName))
		n += int64(m)
		if err != nil {
			return
		}

		m, err = fmt.Fprintf(w, "\n\n")
		n += int64(m)
		if err != nil {
			return
		}

		m, err = w.Write(x.ToCapnCodeFor(s.goName))
		n += int64(m)
		if err != nil {
			return
		}

	} // end second loop over structs for translating methods.

	return
}

func ExtractFromString(src string) ([]byte, error) {
	return ExtractStructs("", "package main; "+src, nil)
}

func ExtractString2String(src string) string {

	x := NewExtractor()
	defer x.Cleanup()
	_, err := ExtractStructs("", "package main; "+src, x)
	if err != nil {
		panic(err)
	}

	// final write, this time accounting for capid tag ordering
	var buf bytes.Buffer
	_, err = x.WriteToSchema(&buf)
	if err != nil {
		panic(err)
	}
	_, err = x.WriteToTranslators(&buf)
	if err != nil {
		panic(err)
	}

	return string(buf.Bytes())
}

func ExtractCapnToGoCode(src string, goName string) string {

	x := NewExtractor()
	defer x.Cleanup()
	_, err := ExtractStructs("", "package main; "+src, x)
	if err != nil {
		panic(err)
	}
	x.GenerateTranslators()
	return string(x.ToGoCodeFor(goName))
}

func ExtractGoToCapnCode(src string, goName string) string {

	x := NewExtractor()
	defer x.Cleanup()
	_, err := ExtractStructs("", "package main; "+src, x)
	if err != nil {
		panic(err)
	}
	x.GenerateTranslators()
	return string(x.ToCapnCodeFor(goName))
}

// ExtractStructs pulls out the struct definitions from a golang source file.
//
// src has to be string, []byte, or io.Reader, as in parser.ParseFile(). src
// can be nil if fname is provided. See http://golang.org/pkg/go/parser/#ParseFile
//
func ExtractStructs(fname string, src interface{}, x *Extractor) ([]byte, error) {
	if x == nil {
		x = NewExtractor()
		defer x.Cleanup()
	}

	return x.ExtractStructsFromOneFile(src, fname)
}

func (x *Extractor) ExtractStructsFromOneFile(src interface{}, fname string) ([]byte, error) {

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
								// *ast.InterfaceType and *ast.Ident end up here.
								//fmt.Printf("\n\n unrecog type ty = %#v\n", ty)
							case (*ast.Ident):
								goNewTypeName := ts2.Name.Obj.Name
								goTargetTypeName := ty.Name
								x.NoteTypedef(goNewTypeName, goTargetTypeName)

							case (*ast.StructType):
								stru := ts2.Type.(*ast.StructType)

								err = x.StartStruct(curStructName)
								if err != nil {
									return []byte{}, err
								}
								//fmt.Printf("\n\n stru = %#v\n", stru)
								//goon.Dump(stru)

								if stru.Fields != nil {
									for _, fld := range stru.Fields.List {
										if fld != nil {
											//fmt.Printf("\n\n    fld.Names = %#v\n", fld.Names) // looking for
											//goon.Dump(fld.Names)

											if len(fld.Names) == 0 {
												// field without name: embedded/anonymous struct
												typeName := fld.Type.(*ast.Ident).Name

												err = x.GenerateStructField(typeName, "", typeName, fld, NotList, fld.Tag, YesEmbedded)
												if err != nil {
													return []byte{}, err
												}

											} else {
												// field with name
												for _, ident := range fld.Names {

													switch ident.Obj.Decl.(type) {
													case (*ast.Field):
														// named field
														fld2 := ident.Obj.Decl.(*ast.Field)

														//fmt.Printf("\n\n    fld2 = %#v\n", fld2)
														//goon.Dump(fld2)

														typeNamePrefix, ident4 := GetTypeAsString(fld2.Type, "")
														//fmt.Printf("\n\n tnas = %#v, ident4 = %s\n", typeNamePrefix, ident4)

														err = x.GenerateStructField(ident.Name, typeNamePrefix, ident4, fld2, IsSlice(typeNamePrefix), fld2.Tag, NotEmbedded)
														if err != nil {
															return []byte{}, err
														}

														/*
															switch fld2.Type.(type) {

															case (*ast.StarExpr):
																star2 := fld2.Type.(*ast.StarExpr)
																err = x.GenerateStructField(ident.Name, star2.X.(*ast.Ident).Name, fld2, NotList, fld2.Tag, NotEmbedded)
																if err != nil {
																	return []byte{}, err
																}

															case (*ast.Ident):
																ident2 := fld2.Type.(*ast.Ident)
																err = x.GenerateStructField(ident.Name, ident2.Name, fld2, NotList, fld2.Tag, NotEmbedded)
																if err != nil {
																	return []byte{}, err
																}

															case (*ast.ArrayType):
																// slice or array
																array2 := fld2.Type.(*ast.ArrayType)
																switch array2.Elt.(type) {
																case (*ast.Ident):
																	err = x.GenerateStructField(ident.Name, array2.Elt.(*ast.Ident).Name, fld2, YesIsList, fld2.Tag, NotEmbedded)
																	if err != nil {
																		return []byte{}, err
																	}
																case (*ast.StarExpr):
																	fmt.Printf("\n\n in array type is *ast.StarExpr\n")
																	err = x.GenerateStructField(ident.Name, array2.Elt.(*ast.StarExpr).X.(*ast.Ident).Name, fld2, YesIsList, fld2.Tag, NotEmbedded)
																	if err != nil {
																		return []byte{}, err
																	}
																}
															}
														*/
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

	return x.out.Bytes(), err
}

func IsSlice(tnas string) bool {
	return strings.HasPrefix(tnas, "[]")
}

func (x *Extractor) NoteTypedef(goNewTypeName string, goTargetTypeName string) {
	// we just want to preserve the mapping, without adding Capn suffix
	//fmt.Printf("\n\n noting typedef: goNewTypeName: '%s', goTargetTypeName: '%s'\n", goNewTypeName, goTargetTypeName)
	//x.goType2capType[goNewTypeName] = goNewTypeName
	isList := false
	x.goType2capType[goNewTypeName] = x.GoTypeToCapnpType(goTargetTypeName, &isList)
}

var regexCapname = regexp.MustCompile(`capname:[ \t]*\"([^\"]+)\"`)

var regexCapid = regexp.MustCompile(`capid:[ \t]*\"([^\"]+)\"`)

func GoType2CapnType(gotypeName string) string {
	return UppercaseFirstLetter(gotypeName) + "Capn"
}

func (x *Extractor) StartStruct(goName string) error {
	x.fieldCount = 0

	capname := GoType2CapnType(goName)
	x.goType2capType[goName] = capname
	x.capType2goType[capname] = goName

	// check for rename comment, capname:"newCapName"
	if x.heldComment != "" {

		match := regexCapname.FindStringSubmatch(x.heldComment)
		if match != nil {
			if len(match) == 2 {
				capname = match[1]
			}
		}
	}

	if isCapnpKeyword(capname) {
		err := fmt.Errorf(`after uppercasing the first letter, struct '%s' becomes '%s' but this is a reserved capnp word, so please write a comment annotation just before the struct definition in go (e.g. // capname:"capName") to rename it.`, goName, capname)
		panic(err)
		return err
	}

	fmt.Fprintf(&x.out, "struct %s { %s", capname, x.fieldSuffix)

	x.curStruct = NewStruct(capname, goName)
	x.curStruct.comment = x.heldComment
	x.heldComment = ""
	x.srs[goName] = x.curStruct

	return nil
}
func (x *Extractor) EndStruct() {
	fmt.Fprintf(&x.out, "} %s", x.fieldSuffix)
}

func (x *Extractor) GenerateComment(c string) {
	x.heldComment = x.heldComment + c + "\n"
}

func UppercaseFirstLetter(name string) string {
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

const NotEmbedded = false
const YesEmbedded = true

func (x *Extractor) GenerateStructField(goFieldName string, goFieldTypePrefix string, goFieldTypeName string, fld *ast.Field, isList bool, tag *ast.BasicLit, IsEmbedded bool) error {

	if goFieldTypeName == "" {
		return nil
	}

	//fmt.Printf("\n\n\n GenerateStructField called with goFieldName = '%s', goFieldTypeName = '%s', fld = %#v, tag = %#v\n\n", goFieldName, goFieldTypeName, fld, tag)

	// if we are ignoring private (lowercase first letter) fields, then stop here.
	if !IsEmbedded {
		if len(goFieldName) > 0 && unicode.IsLower([]rune(goFieldName)[0]) && !x.extractPrivate {
			return nil
		}
	}

	curField := &Field{orderOfAppearance: x.fieldCount, embedded: IsEmbedded}

	var tagValue string
	loweredName := underToCamelCase(LowercaseCapnpFieldName(goFieldName))

	if tag != nil {
		//fmt.Printf("tag = %#v\n", tag)

		if tag.Value != "" {

			// capname tag
			match := regexCapname.FindStringSubmatch(tag.Value)
			if match != nil {
				if len(match) == 2 {
					//fmt.Printf("matched, using '%s' instead of '%s'\n", match[1], goFieldName)
					loweredName = match[1]
					tagValue = tag.Value

					if isCapnpKeyword(loweredName) {
						err := fmt.Errorf(`problem detected after applying the capname tag '%s' found on field '%s': '%s' is a reserved capnp word, so please use a *different* struct field tag (e.g. capname:"capnpName") to rename it`, tag.Value, goFieldName, loweredName)
						return err
					}

				}
			}

			// capid tag
			match2 := regexCapid.FindStringSubmatch(tag.Value)
			if match2 != nil {
				if len(match2) == 2 {
					//fmt.Printf("matched, applying capid tag '%s' for field '%s'\n", match2[1], loweredName)
					n, err := strconv.Atoi(match2[1])
					if err != nil {
						err := fmt.Errorf(`problem in capid tag '%s' on field '%s' in struct '%s': could not convert to number, error: '%s'`, match2[1], goFieldName, x.curStruct.goName, err)
						panic(err)
						return err
					}
					fld, already := x.curStruct.capIdMap[n]
					if already {
						err := fmt.Errorf(`problem in capid tag '%s' on field '%s' in struct '%s': number '%d' is already taken by field '%s'`, match2[1], goFieldName, x.curStruct.goName, n, fld.goName)
						panic(err)
						return err

					} else {
						x.curStruct.capIdMap[n] = curField
						curField.capIdFromTag = n
					}
				}
			}
		}

	}

	//fmt.Printf("\n\n\n GenerateStructField: goFieldName:'%s' -> loweredName:'%s'\n\n", goFieldName, loweredName)

	if isCapnpKeyword(loweredName) {
		err := fmt.Errorf(`after lowercasing the first letter, field '%s' becomes '%s' but this is a reserved capnp word, so please use a struct field tag (e.g. capname:"capnpName") to rename it`, goFieldName, loweredName)
		return err
	}

	capnTypeDisplayed := x.GoTypeToCapnpType(goFieldTypeName, &isList)

	//fmt.Printf("\n\n\n DEBUG:  '%s' '%s' @%d: %s; %s\n\n", x.fieldPrefix, loweredName, x.fieldCount, capnTypeDisplayed, x.fieldSuffix)

	if isList {
		fmt.Fprintf(&x.out, "%s%s @%d: List(%s); %s", x.fieldPrefix, loweredName, x.fieldCount, capnTypeDisplayed, x.fieldSuffix)
	} else {
		fmt.Fprintf(&x.out, "%s%s @%d: %s; %s", x.fieldPrefix, loweredName, x.fieldCount, capnTypeDisplayed, x.fieldSuffix)
	}

	sz := len(loweredName)
	if sz > x.curStruct.longestField {
		x.curStruct.longestField = sz
	}

	curField.capname = loweredName

	curField.GoCapGoName = UppercaseFirstLetter(loweredName)

	curField.capType = capnTypeDisplayed
	curField.goName = goFieldName
	curField.goType = goFieldTypeName
	curField.isList = isList
	curField.tagValue = tagValue
	curField.goTypePrefix = goFieldTypePrefix

	x.curStruct.fld = append(x.curStruct.fld, curField)
	x.fieldCount++

	//fmt.Printf("\n\n curField = %#v\n", curField)

	return nil
}

func (x *Extractor) GoTypeToCapnpType(goFieldTypeName string, isList *bool) (capnTypeDisplayed string) {

	switch goFieldTypeName {
	case "string":
		capnTypeDisplayed = "Text"
	case "int":
		capnTypeDisplayed = "Int64"
	case "bool":
		capnTypeDisplayed = "Bool"
	case "int8":
		capnTypeDisplayed = "Int8"
	case "int16":
		capnTypeDisplayed = "Int16"
	case "int32":
		capnTypeDisplayed = "Int32"
	case "int64":
		capnTypeDisplayed = "Int64"
	case "uint8":
		capnTypeDisplayed = "UInt8"
	case "uint16":
		capnTypeDisplayed = "UInt16"
	case "uint32":
		capnTypeDisplayed = "UInt32"
	case "uint64":
		capnTypeDisplayed = "UInt64"
	case "float32":
		capnTypeDisplayed = "Float32"
	case "float64":
		capnTypeDisplayed = "Float64"
	case "byte":
		if *isList {
			capnTypeDisplayed = "Data"
			*isList = false
		} else {
			capnTypeDisplayed = "Uint8"
		}
	default:

		alreadyKnownCapnType := x.goType2capType[goFieldTypeName]
		if alreadyKnownCapnType != "" {
			//fmt.Printf("\n\n debug: x.goType2capType[goFieldTypeName='%s'] -> '%s'\n", goFieldTypeName, alreadyKnownCapnType)
			capnTypeDisplayed = alreadyKnownCapnType
		} else {
			capnTypeDisplayed = GoType2CapnType(goFieldTypeName)
			//fmt.Printf("\n\n debug: adding to  x.goType2capType[goFieldTypeName='%s'] = '%s'\n", goFieldTypeName, capnTypeDisplayed)
			x.goType2capType[goFieldTypeName] = capnTypeDisplayed
		}

		/*		if isCapnpKeyword(capnTypeDisplayed) {
					err := fmt.Errorf(`after uppercasing the first letter, type '%s' becomes '%s' but this is a reserved capnp word, so please use a different type name`, goFieldTypeName, capnTypeDisplayed)
					panic(err)
				}
		*/
	}
	return
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

func (x *Extractor) GenCapnpHeader() *bytes.Buffer {
	var by bytes.Buffer

	id := getNewCapnpId()

	fmt.Fprintf(&by, `%s;
using Go = import "go.capnp";
$Go.package("%s");
$Go.import("%s");
%s`, id, x.pkgName, x.importDecl, x.fieldSuffix)

	return &by
}

func (x *Extractor) AssembleCapnpFile(in []byte) *bytes.Buffer {
	by := x.GenCapnpHeader()

	by.Write(in)
	fmt.Fprintf(by, "\n")

	return by
}

func CapnpCompileFragment(in []byte) ([]byte, error, *Extractor) {
	x := NewExtractor()
	out, err := x.CapnpCompileFragment(in)
	return out, err, x
}

func (x *Extractor) CapnpCompileFragment(in []byte) ([]byte, error) {

	if x.compileDir != nil {
		x.compileDir.Cleanup()
	}
	x.compileDir = NewTempDir()

	f := x.compileDir.TempFile()
	//fnCapnGoOutput := tempfile.Name() + ".go"

	by := x.AssembleCapnpFile(in)
	debug := string(by.Bytes())

	f.Write(by.Bytes())
	f.Close()

	compiled, combinedOut, err := CapnpCompilePath(f.Name())
	if err != nil {
		errmsg := fmt.Sprintf("error compiling the generated capnp code: '%s'; error: '%s'\n", debug, err) + string(combinedOut)
		return []byte(errmsg), fmt.Errorf(errmsg)
	}

	return compiled, nil
}

func CapnpCompilePath(fname string) (generatedGoFile []byte, comboOut []byte, err error) {
	goOutFn := fname + ".go"

	by, err := exec.Command("capnp", "compile", "-ogo", fname).CombinedOutput()
	if err != nil {
		return []byte{}, by, err
	}

	generatedGoFile, err = ioutil.ReadFile(goOutFn)

	return generatedGoFile, by, err
}

func SetSpaces(spaces *string, Max int, Len int) {
	if Len >= Max {
		*spaces = ""
		return
	}
	*spaces = strings.Repeat(" ", Max-Len)
}

func ExtraSpaces(fieldNum int) string {
	if fieldNum < 10 {
		return "  "
	}
	if fieldNum < 100 {
		return " "
	}
	return ""
}

func IsIntrinsicGoType(goFieldTypeName string) bool {
	fmt.Printf("\n IsIntrinsic called with '%s'\n", goFieldTypeName)

	switch goFieldTypeName {
	case "string":
		return true
	case "int":
		return true
	case "bool":
		return true
	case "int8":
		return true
	case "int16":
		return true
	case "int32":
		return true
	case "int64":
		return true
	case "uint8":
		return true
	case "uint16":
		return true
	case "uint32":
		return true
	case "uint64":
		return true
	case "float32":
		return true
	case "float64":
		return true
	case "byte":
		return true
	default:
		return false
	}
	return false
}
