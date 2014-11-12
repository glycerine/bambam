all:
	rm -f translateCapn.go schema.capnp.go && go test -v && go build && go install

