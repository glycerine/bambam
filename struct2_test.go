package main

import (
	"testing"

	cv "github.com/smartystreets/goconvey/convey"
)

func TestKeywordConflictDetection(t *testing.T) {

	cv.Convey("Given a golang source file with a struct named Data", t, func() {
		cv.Convey("then we should insist on rename. Same goes for any go structures that conflict with capnp reserved words.", func() {

			ex0 := `
type Data struct {
}`
			cv.So(func() { ExtractString2String(ex0) }, cv.ShouldPanic)

			ex3 := `
type Data struct {
}`
			cv.So(func() { ExtractString2String(ex3) }, cv.ShouldPanic)

			ex1 := `
type Void struct {
}`
			cv.So(func() { ExtractString2String(ex1) }, cv.ShouldPanic)

			ex2 := `
type void struct {
}`
			cv.So(func() { ExtractString2String(ex2) }, cv.ShouldPanic)

		})

	})
}
