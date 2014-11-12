package main

import (
	"testing"

	cv "github.com/smartystreets/goconvey/convey"
)

func TestPointerInStruct(t *testing.T) {

	cv.Convey("Given a struct that contains a pointer to a struct", t, func() {
		cv.Convey("then the pointer element should be de-pointerized", func() {

			ex0 := `
type big struct {}
type s1 struct {
  MyBig *big
}`
			cv.So(ExtractString2String(ex0), ShouldStartWithModuloWhiteSpace, `struct BigCapn { } struct S1Capn { myBig @0: BigCapn; } `)

		})
	})
}
