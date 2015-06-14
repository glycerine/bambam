package main

import (
	"testing"

	cv "github.com/glycerine/goconvey/convey"
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

			s1, err := ExtractFromString(in1)
			if err != nil {
				panic(err)
			}

			//expect1 := `struct In1 { str @0: Text; n @1: Int64; d @2: Float64; } `
			//cv.So(string(s1), cv.ShouldEqual, expect1)

			// no news on compile is good news
			_, err, x := CapnpCompileFragment(s1)
			cv.So(err, cv.ShouldEqual, nil)

			x.Cleanup()
		})
	})
}
