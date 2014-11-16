package main

import (
	"testing"

	cv "github.com/smartystreets/goconvey/convey"
)

func TestSave1(t *testing.T) {

	cv.Convey("Given a parsable golang source file", t, func() {
		cv.Convey("then structs with public fields get a Save() method to serialize them, and a Load() method to restore them.", func() {

			exEmbed := `
type RWTest struct {
	Hello []string
}
`
			cv.So(ExtractString2String(exEmbed), ShouldMatchModuloWhiteSpace, `
struct RWTestCapn {
  hello @0: List(Text);
}

func (s *RWTest) Save(w io.Writer) {
	seg := capn.NewBuffer(nil)
	RWTestGoToCapn(seg, s)
	seg.WriteTo(w)
}

func (s *RWTest) Load(r io.Reader) {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		panic(fmt.Errorf("capn.ReadFromStream error: %s", err))
	}
	z := testpkg.ReadRootRWTestCapn(capMsg)
    RWTestCapnToGo(z, s)
}
  
  func RWTestCapnToGo(src RWTestCapn, dest *RWTest) *RWTest { 
    if dest == nil { 
      dest = &RWTest{} 
    }
    dest.Hello = src.Hello().ToArray()
  
    return dest
  } 
    
  func RWTestGoToCapn(seg *capn.Segment, src *RWTest) RWTestCapn { 
    dest := AutoNewRWTestCapn(seg)
  
    mylist1 := seg.NewTextList(len(src.Hello))
    for i := range src.Hello {
       mylist1.Set(i, string(src.Hello[i]))
    }
    dest.SetHello(mylist1)
  
    return dest
  } 
`)

		})

	})
}

func TestSave2(t *testing.T) {

	cv.Convey("Given a parsable golang source file", t, func() {
		cv.Convey("then structs with public fields get a save() method to serialize them, and a load() method to restore them.", func() {

			exEmbed := `
type RWTest struct {
	Hello []string
    World []int
}
`
			cv.So(ExtractString2String(exEmbed), ShouldMatchModuloWhiteSpace, `
struct RWTestCapn {
  hello  @0: List(Text);
  world  @1: List(Int64);
}

func (s *RWTest) Save(w io.Writer) {
	seg := capn.NewBuffer(nil)
	RWTestGoToCapn(seg, s)
	seg.WriteTo(w)
}

func (s *RWTest) Load(r io.Reader) {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		panic(fmt.Errorf("capn.ReadFromStream error: %s", err))
	}
	z := testpkg.ReadRootRWTestCapn(capMsg)
    RWTestCapnToGo(z, s)
}

  
func RWTestCapnToGo(src RWTestCapn, dest *RWTest) *RWTest { 
    if dest == nil { 
      dest = &RWTest{} 
    }
    dest.Hello = src.Hello().ToArray()
  
      var n int
  
      // World
  	n = src.World().Len()
  	dest.World = make([]int, n)
  	for i := 0; i < n; i++ {
          dest.World[i] = int(src.World().At(i))
      }
  
  
    return dest
  } 
    
func RWTestGoToCapn(seg *capn.Segment, src *RWTest) RWTestCapn { 
    dest := AutoNewRWTestCapn(seg)
  
    mylist1 := seg.NewTextList(len(src.Hello))
    for i := range src.Hello {
       mylist1.Set(i, string(src.Hello[i]))
    }
    dest.SetHello(mylist1)
  
    mylist2 := seg.NewInt64List(len(src.World))
    for i:= range src.World {
       mylist2.Set(i, int64(src.World[i]))
    }
    dest.SetWorld(mylist2)

    return dest
}
`)

		})

	})
}
