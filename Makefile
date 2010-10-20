include $(GOROOT)/src/Make.inc

TARG=httplib
GOFMT=gofmt -spaces=true -tabindent=false -tabwidth=4

GOFILES=\
	httplib.go\

include $(GOROOT)/src/Make.pkg

format:
	${GOFMT} -w httplib.go
