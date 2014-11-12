all:
	rm -f translateCapn.go schema.capnp.go && go test -v && go build && go install

test:
	./bambam testpkg/t.go
	capnp compile -ogo schema.capnp
	mv schema.capnp* testpkg
	perl -pi -e 's/main/testpkg/' translateCapn.go
	mv translateCapn.go testpkg/
	cd testpkg; go build
