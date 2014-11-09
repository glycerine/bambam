package main

import (
	"testing"

	cv "github.com/smartystreets/goconvey/convey"
)

type EmptyStruct struct {
}
type OneStruct struct {
	One int
}
type HalfPrivStruct struct {
	One int
	two int
}

type NestingStruct struct {
	One    int
	Nester OneStruct
}

type EmbeddingStruct struct {
	OneStruct
}

// An example struct with a comment
// second line of comment.
/*
 C comments attached too
*/
type ExampleStruct1 struct {
	Str string
	N   int
}

func TestSimpleStructExtraction(t *testing.T) {

	cv.Convey("Given a compilable (syntax correct) golang (Go) source file", t, func() {
		cv.Convey("then we should be able properly to extract simple structs, without recursion or embedding", func() {
			//cv.So(res, cv.ShouldEqual, false)
		})
	})
}
