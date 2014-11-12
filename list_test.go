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
  
  func S1CapnToGo(src testpkg.S1Capn, dest *s1) *s1 { 
    if dest = nil { 
      dest = &s1{} 
    }
    dest.Names = src.Names()
  
    return dest
  } 
  
  func s1GoToCapn(seg *capn.Segment, src *s1, dest testpkg.S1Capn) testpkg.S1Capn { 
    if dest = nil {
        dest := testpkg.NewS1Capn(seg)
    }
  
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

/*

			expected1 :=
`	if len(js.RunQ) > 0 {
		//fmt.Printf("len of Runq: %d, %#v\n", len(js.RunQ), js.RunQ)
		runq := schema.NewZjobList(seg, len(js.RunQ))
		plistRunq := capn.PointerList(runq)
		i = 0
		for _, j = range js.RunQ {
			zjob := JobToCapnpSegment(j, seg)
			plistRunq.Set(i, capn.Object(zjob))
			i++
		}
		zjs.SetRunq(runq)
	}

	// waitingjobs
	if len(js.WaitingJobs) > 0 {
		//fmt.Printf("len of WaitingJobs: %d, %#v\n", len(js.WaitingJobs), js.WaitingJobs)
		waitingjobs := schema.NewZjobList(seg, len(js.WaitingJobs))
		plistWaitingjobs := capn.PointerList(waitingjobs)
		i = 0
		for _, j = range js.WaitingJobs {
			zjob := JobToCapnpSegment(j, seg)
			plistWaitingjobs.Set(i, capn.Object(zjob))
			i++
		}
		zjs.SetWaitingjobs(waitingjobs)
	}
`
*/
