package main

/*
func TestWriteReadThroughGeneratedTranslationCode(t *testing.T) {

	cv.Convey("Given bambam generated go bindings", t, func() {
		cv.Convey("then we should be able to write to disk, and read back the same structure", func() {

			tdir := NewTempDir()
			//defer tdir.Cleanup()

			MainArgs([]string{os.Args[0], "-o", tdir.DirPath, "rw.go.txt"})

			err := exec.Command("cp", "rw.go.txt", tdir.DirPath+"/rw.go").Run()
			cv.So(err, cv.ShouldEqual, nil)

			tdir.MoveTo()

			err = exec.Command("capnpc", "-ogo", "schema.capnp").Run()
			cv.So(err, cv.ShouldEqual, nil)

			err = exec.Command("go", "build").Run()
			cv.So(err, cv.ShouldEqual, nil)

		})
	})
}
*/
