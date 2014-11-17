bambam: auto-generate capnproto schema from your golang source files.
======

Adding [capnproto serialization](https://github.com/glycerine/go-capnproto) to an existing Go project used to mean writing alot of boilerplate.

Not anymore.

Given a set of golang (Go) source files, bambam will generate a [capnproto](http://kentonv.github.io/capnproto/) schema. Even better: bambam will also generate translation functions to readily convert between your golang structs and the new capnproto structs.

prereqs
-------

[Capnproto](http://kentonv.github.io/capnproto/) and [go-capnproto](https://github.com/glycerine/go-capnproto) should both be installed and on your PATH.

to install
--------

`go get -t github.com/glycerine/bambam`  # the -t pulls in the test dependencies.

use
---------

~~~
use: bambam -o outdir -p package myGoSourceFile.go myGoSourceFile2.go ...
     # Bambam makes it easy to use Capnproto serialization[1] from Go.
     # Bambam reads .go files and writes a .capnp schema and Go bindings.
     # options:
     #   -o="odir" specifies the directory to write to (created if need be).
     #   -p="main" specifies the package header to write (e.g. main, mypkg).
     #   -X exports private fields of Go structs. Default only maps public fields.
     #   -version   shows build version with git commit hash
     # required: at least one .go source file for struct definitions. Must be last, after options.
     #
     # [1] https://github.com/glycerine/go-capnproto 
~~~

demo
-----

See rw.go.txt. Also: after running `go test`, cd into testdir_* and look at the sample project files there.

Here is what use looks like. You end up with a Save() and Load() function for each of your structs. Simple!

~~~
package main

import (
    "bytes"
)

type MyStruct struct {
	Hello    []string
	World    []int
}

func main() {

	rw := MyStruct{
		Hello:    []string{"one", "two", "three"},
		World:    []int{1, 2, 3},
	}

    // any io.ReadWriter will work here (os.File, etc)
	var o bytes.Buffer

	rw.Save(&o)
    // now we have saved!


    rw2 := &MyStruct{}
	rw2.Load(&o)
    // now we have restored!

}

~~~

what Go types does bambam recognize?
----------------------------------------

Supported: structs, slices, and primitve/scalar types are supported. Structs that contain structs are supported. You have both slices of scalars and slices of structs available.

Currently unsupported (at the moment; pull requests welcome): Go maps.  

Also: some pointers work, but pointers in the inner-most struct do not. This is not a big limitation, as it is rarely meaningful to pass a pointer value to a different process.

Not planned (likely never supported): Go interfaces, Go channels.

capid tags on go structs
--------------------------

When you run `bambam`, it will generate a modified copy of your go source files in the output directory.

These new versions include capid tags on all public fields of structs. You should inspect the copy of the source file in the output directory, and then replace your original source with the tagged version.  You can also manually add capid tags to fields, if you need to manually specify a field number (e.g. you are matching an pre-existing capnproto definition).

Only public fields (with Captial first letter in their name) are tagged. The -X flag ignores the public/private distinction, and tags all fields.

The capid tags allow the capnproto schema evolution to function properly as you add new fields to structs. If you don't include the capid tags, your serialization code won't be backwards compatible as you change your structs.

example of capid annotion use
~~~
type Job struct { 
   C int `capid:"2"`  // we added C later, thus it is numbered higher.
   A int `capid:"0"`
   B int `capid:"1"` 
}
~~~

-----
-----

Copyright (c) 2014, Jason E. Aten, Ph.D.

