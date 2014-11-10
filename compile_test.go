package main

import (
	"testing"

	cv "github.com/smartystreets/goconvey/convey"
)

func TestCapnpWillCompileOurOutput(t *testing.T) {

	cv.Convey("Given a parsable golang source file with a simple struct", t, func() {
		cv.Convey("then when we generate capnp code, it should compile", func() {

			in1 := `
type in1 struct {
	Str string
	N   int
    D   float64
}`

			s1 := ExtractFromString(in1)
			//expect1 := `struct In1 { str @0: Text; n @1: Int64; d @2: Float64; } `
			//cv.So(string(s1), cv.ShouldEqual, expect1)

			// no news on compile is good news
			cv.So(string(CapnpCompileFragment(s1)), cv.ShouldEqual, ``)

		})
	})
}
