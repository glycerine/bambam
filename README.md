bambam: auto-generate capnproto schema from your golang source files.
======

Adding [capnproto serialization](https://github.com/glycerine/go-capnproto) to an existing Go project can mean writing alot of boilerplate.

Not anymore.

Given a golang source file, bambam will generate a [capnproto](http://kentonv.github.io/capnproto/) schema. Even better: bambam will also generate translation functions to readily convert between golang structs and capnproto structs.

use
---

`$ bambam <sourcefile.go>`

prereqs
-------

[Capnproto](http://kentonv.github.io/capnproto/) and [go-capnproto](https://github.com/glycerine/go-capnproto) should both be installed and on your PATH.

to install
--------

`go get -t github.com/glycerine/bambam`  # the -t pulls in the test dependencies.


Copyright (c) 2014, Jason E. Aten, Ph.D.

