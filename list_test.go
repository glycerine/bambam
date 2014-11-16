package main

import (
	"testing"

	cv "github.com/smartystreets/goconvey/convey"
)

func TestSliceOfStringToListOfText(t *testing.T) {

	cv.Convey("Given a struct that contains a slice of string", t, func() {
		cv.Convey("then the capnp schema should contain a list of string and list translating code should be generated", func() {

			ex0 := `
type s1 struct {
   Names []string
}`

			expected0 := `
  struct S1Capn { names  @0:   List(Text); } 

    func (s *s1) Save(w io.Writer) {
    	seg := capn.NewBuffer(nil)
    	s1GoToCapn(seg, s)
    	seg.WriteTo(w)
    }
      
    func (s *s1) Load(r io.Reader) {
    	capMsg, err := capn.ReadFromStream(r, nil)
    	if err != nil {
    		panic(fmt.Errorf("capn.ReadFromStream error: %s", err))
    	}
    	z := schema.ReadRootS1Capn(capMsg)
        S1CapnToGo(z, s)
    }
  
  func S1CapnToGo(src S1Capn, dest *s1) *s1 { 
    if dest == nil { 
      dest = &s1{} 
    }

    dest.Names = src.Names().ToArray()
  
    return dest
  } 
  
  func s1GoToCapn(seg *capn.Segment, src *s1) S1Capn { 
    dest := AutoNewS1Capn(seg)
  
    // text list
    tl := seg.NewTextList(len(src.Names))
    for i := range src.Names {
       tl.Set(i, src.Names[i])
    }
    dest.SetNames(tl)
  
    return dest
  } 
`

			cv.So(ExtractString2String(ex0), ShouldMatchModuloWhiteSpace, expected0)

		})
	})
}
