package main

import (
	"testing"

	cv "github.com/smartystreets/goconvey/convey"
)

func TestTagAnnotationWorks(t *testing.T) {

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
