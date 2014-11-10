package main

import (
	"testing"

	cv "github.com/smartystreets/goconvey/convey"
)

func TestSliceToList(t *testing.T) {

	cv.Convey("Given a parsable golang source file with struct containing a slice", t, func() {
		cv.Convey("then the slice should be converted to a List() in the capnp output", func() {

			ex0 := `
type s1 struct {
  MyInts []int
}`
			cv.So(ExtractString2String(ex0), cv.ShouldEqual, `struct S1 { myInts @0: List(Int64); } `)

		})
	})
}
