package main

import (
	"testing"

	cv "github.com/smartystreets/goconvey/convey"
)

func TestCommentAnnotationWorks(t *testing.T) {

	cv.Convey(`Given a golang struct named Data with an attached comment: // capname:"MyData"`, t, func() {
		cv.Convey("then we should use the capname MyData for the struct.", func() {
			ex0 := `
// capname:"MyData"
type Data struct {
}`
			cv.So(ExtractString2String(ex0), cv.ShouldEqual, `struct MyData { } `)

			ex1 := `
// capname: "MyData"
type Data struct {
}`
			cv.So(ExtractString2String(ex1), cv.ShouldEqual, `struct MyData { } `)

		})

	})
}

func TestTagAnnotationWorks(t *testing.T) {

	cv.Convey("Given a golang struct with an invalid capnp field name 'union', but with a field tag: `capname:\"MyNewFieldName\"`", t, func() {
		cv.Convey("then we should use the capname MyNewFieldName for the field, and not fail the run.", func() {
			ex0 := "type S struct { union string `capname:\"MyNewFieldName\"` \n}"
			cv.So(ExtractString2String(ex0), cv.ShouldEqual, `struct S { MyNewFieldName @0: Text; } `)

			ex1 := "type S struct { union string `capname: \t \t \"MyNewFieldName\"` \n}"
			cv.So(ExtractString2String(ex1), cv.ShouldEqual, `struct S { MyNewFieldName @0: Text; } `)

		})

	})
}

func TestTagCapidWorks(t *testing.T) {

	cv.Convey("Given the desire to preserve the field numbering in the generated Capnproto schema,", t, func() {
		cv.Convey("when we add the tag: capid:\"1\", then the field should be numbered @1.", func() {
			cv.Convey("and if there are fewer than 2 (numbered 0, 1) fields then we error out.", func() {
				ex0 := "type S struct { a string `capid:\"1\"`; b string \n}"
				cv.So(ExtractString2String(ex0), cv.ShouldEqual, `struct S { b @0: Text; a @1: Text; } `)
			})
		})

	})
}
