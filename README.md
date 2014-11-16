bambam: auto-generate capnproto schema from your golang source files.
======

Adding [capnproto serialization](https://github.com/glycerine/go-capnproto) to an existing Go project can mean writing alot of boilerplate.

Not anymore.

Given a golang source file, bambam will generate a [capnproto](http://kentonv.github.io/capnproto/) schema. Even better: bambam will also generate translation functions to readily convert between golang structs and capnproto structs.

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

Copyright (c) 2014, Jason E. Aten, Ph.D.

