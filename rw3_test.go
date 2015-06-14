package main

import (
	"os"
	"os/exec"
	"testing"

	cv "github.com/glycerine/goconvey/convey"
)

func Test017WriteRead_StructPointerWithinStruct(t *testing.T) {

	tdir := NewTempDir()
	// comment the defer out to debug any rw test failures.
	defer tdir.Cleanup()

	err := exec.Command("cp", "rw3.go.txt", tdir.DirPath+"/rw3.go").Run()
	if err != nil {
		panic(err)
	}

	MainArgs([]string{os.Args[0], "-o", tdir.DirPath, "rw3.go.txt"})

	cv.Convey("Given bambam generated go bindings: with a struct pointer within a struct", t, func() {
		cv.Convey("then we should be able to write to disk, and read back the same structure", func() {
			cv.So(err, cv.ShouldEqual, nil)

			tdir.MoveTo()

			err = exec.Command("capnpc", "-ogo", "schema.capnp").Run()
			cv.So(err, cv.ShouldEqual, nil)

			err = exec.Command("go", "build").Run()
			cv.So(err, cv.ShouldEqual, nil)

			// run it
			err = exec.Command("./" + tdir.DirPath).Run()
			cv.So(err, cv.ShouldEqual, nil)

		})
	})
}
