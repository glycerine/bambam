package main

import (
	"testing"

	cv "github.com/smartystreets/goconvey/convey"
)

func Test011BoolFields(t *testing.T) {

	cv.Convey("Given a struct that contains a bool", t, func() {
		cv.Convey("then the capnp schema should contain a the bool field and we should generate translation code for the bool field", func() {

			ex0 := `
type s1 struct {
   IsRequest bool
}`

			expected0 := `
  struct S1Capn { isRequest  @0:   Bool; } 

    func (s *s1) Save(w io.Writer) error {
    	seg := capn.NewBuffer(nil)
    	s1GoToCapn(seg, s)
    	_, err := seg.WriteTo(w)
        return err
    }
      
    func (s *s1) Load(r io.Reader) error {
    	capMsg, err := capn.ReadFromStream(r, nil)
    	if err != nil {
    		//panic(fmt.Errorf("capn.ReadFromStream error: %s", err))
            return err
    	}
    	z := ReadRootS1Capn(capMsg)
        S1CapnToGo(z, s)
        return nil
    }
  
  func S1CapnToGo(src S1Capn, dest *s1) *s1 { 
    if dest == nil { 
      dest = &s1{} 
    }
	dest.IsRequest = src.IsRequest()
  
    return dest
  } 
  
  func s1GoToCapn(seg *capn.Segment, src *s1) S1Capn { 
    dest := AutoNewS1Capn(seg)
	dest.SetIsRequest(src.IsRequest)  
  
    return dest
  } 

`

			cv.So(ExtractString2String(ex0), ShouldMatchModuloWhiteSpace, expected0)
			//cv.So(expected0, ShouldStartWithModuloWhiteSpace, ExtractString2String(ex0))

		})
	})
}
