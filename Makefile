
GOFMT=gofmt -tabs=false -tabwidth=4

GOFILES=\
	httplib.go\

format:
	${GOFMT} -w httplib.go
	${GOFMT} -w httplib_test.go
