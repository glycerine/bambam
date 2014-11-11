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
			cv.So(ExtractString2String(ex0), cv.ShouldEqual, `struct S1Capn { myInts  @0:   List(Int64); } `)

		})
	})
}

func TestSliceOfStructToList(t *testing.T) {

	cv.Convey("Given a parsable golang source file with struct containing a slice of struct bbb", t, func() {
		cv.Convey("then the slice should be converted to a List(Bbb) in the capnp output", func() {

			ex0 := `
type bbb struct {}
type s1 struct {
  MyBees []bbb
}`
			cv.So(ExtractString2String(ex0), cv.ShouldEqual, `struct BbbCapn { } struct S1Capn { myBees  @0:   List(BbbCapn); } `)

		})
	})
}

func TestSliceOfPointerToList(t *testing.T) {

	cv.Convey("Given a parsable golang source file with struct containing a slice of pointers to struct big", t, func() {
		cv.Convey("then the slice should be converted to a List(Big) in the capnp output", func() {

			ex0 := `
type big struct {}
type s1 struct {
  MyBigs []*big
}`
			cv.So(ExtractString2String(ex0), cv.ShouldEqual, `struct BigCapn { } struct S1Capn { myBigs  @0:   List(BigCapn); } `)

		})
	})
}

func TestSliceOfByteBecomesData(t *testing.T) {

	cv.Convey("Given golang src with []byte", t, func() {
		cv.Convey("then the slice should be converted to Data, not List(Byte), in the capnp output", func() {

			ex0 := `
type s1 struct {
  MyData []byte
}`
			cv.So(ExtractString2String(ex0), ShouldMatchModuloSpaces, `struct S1Capn { myData  @0:   Data; } `)

		})
	})
}
