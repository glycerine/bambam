package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	cv "github.com/glycerine/goconvey/convey"
)

func Test003WriteReadThroughGeneratedTranslationCode(t *testing.T) {

	tdir := NewTempDir()
	// comment the defer out to debug any rw test failures.
	defer tdir.Cleanup()

	err := exec.Command("cp", "rw.go.txt", tdir.DirPath+"/rw.go").Run()
	if err != nil {
		fmt.Printf("cp rw.go.txt %s/rw.go failed: '%s'\n", tdir.DirPath, err)
		panic(err)
	}

	MainArgs([]string{os.Args[0], "-o", tdir.DirPath, "rw.go.txt"})

	cv.Convey("Given bambam generated go bindings, \n"+
		"        then we should be able to write to disk, and read back the same structure", t, func() {

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
}
