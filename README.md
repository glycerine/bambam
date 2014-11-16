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

warning (anti-foot-shooting advice)
--------
Beginners should use bambam only to get started with Capnproto. You will loose the schema evolution benefits of Capnproto if you don't preserve the struct field numbers in the schema once they have been first assigned. As a safe guard, also check-in to version control your Capnproto schema (this should be obvious).

Here is where this matters. Suppose that you have a Go struct called Job, and you have added Capnproto serialization to your project using bambam. Now you are happily serializing data at the speed of light to and from your Job struct.

Fast-forward to a month later: It is a month later, and now you decide to add an additional Go field at the beginning of the Job struct. If you now *run bambam* again, you will generate a new capnproto schema that is almost surely incompatible with the first schema. This is because the field numbers (@0, @1, @2) will have changed. bambam has no way to know what the old definition of Job was. So it is up to you to insure schemas are evolved in a compatible way.

While we can't do it all for you, bambam does try to help.  Go struct fields can be annotated with the `capid:"3"` struct annotation to insist that bambam assign @3 to a paritcular field in the generated Capnproto schema. In fact you'll need to insist on the numbering of all the already-in-use fields if you want to stay backwards compatible.  

The conclusion and recommendation for avoiding trouble is simple: put capid tags on all your Go struct fields. In the future we imagine being able to add the tags to your go file auto-magically at bootstrap time, but for now this is your responsibility.

example of capid annotion use
~~~
type Job struct { 
   // In the scenario above, C was added later, after Job with only A and P has been in use.
   // So here we insist on a back-compatible field numbering with the capid tag.
   C int `capid:"3"`  // we added C later. 
   A int `capid:"0"`
   P int `capid:"1"` 
   N int `capid:"2"` 
}
~~~

-----
-----

Copyright (c) 2014, Jason E. Aten, Ph.D.

