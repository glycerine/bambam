package main

import (
	"testing"

	cv "github.com/smartystreets/goconvey/convey"
)

func TestCapnpStructNaming(t *testing.T) {

	cv.Convey("Given go structs that we will marshal into capnp structs", t, func() {
		cv.Convey("then the names of the two types in go code should be distinct: Cpn suffix attached to the capnp structs", func() {

			ex0 := `
type Extra struct {
  A int
}`
			cv.So(ExtractString2String(ex0), ShouldMatchModuloSpaces, `struct ExtraCapn { a @0: Int64; } `)
		})
	})
}

func TestMarshal(t *testing.T) {

	cv.Convey("Given go structs", t, func() {
		cv.Convey("then bambam should generate automatic marshal-ing code from the go into the capnproto objects", func() {

		})
	})
}

func TestUnMarshal(t *testing.T) {

	cv.Convey("Given go structs", t, func() {
		cv.Convey("then bambam should generate automatic unmarshal-ing code from the capnp into the go structs", func() {

		})
	})
}
