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

	cv.Convey("Given a parsable golang source file", t, func() {
		cv.Convey("then we can extract an empty struct", func() {

			ex0 := `
type Empty1 struct {
}`
			cv.So(ExtractString(ex0), cv.ShouldEqual, `type Empty1 struct { } `)

		})
		cv.Convey("then we can extract simple structs, without recursion or embedding", func() {

			ex1 := `
type ExampleStruct1 struct {
	Str string
	N   int
}`
			cv.So(ExtractString(ex1), cv.ShouldEqual, `type ExampleStruct1 struct { Str string; N int; } `)

		})
		cv.Convey("then we can extract structs that have other named structs inside", func() {

			exNest := `
type OneStruct struct {
	One int
}
type NestingStruct struct {
	One    int
	Nester OneStruct
}`

			cv.So(ExtractString(exNest), cv.ShouldEqual, `type OneStruct struct { One int; } type NestingStruct struct { One int; Nester OneStruct; } `)

		})

		cv.Convey("then we can extract structs with anonymous (embedded) structs inside", func() {

			exEmbed := `
type OneStruct struct {
	One int
}
type EmbedsOne struct {
	OneStruct
}`

			cv.So(ExtractString(exEmbed), cv.ShouldEqual, `type OneStruct struct { One int; } type EmbedsOne struct { OneStruct; } `)

		})

	})
}
