include $(GOROOT)/src/Make.inc

TARG=exif4go
GOFMT=gofmt -s -spaces=true -tabindent=false -tabwidth=4

GOFILES=\
  exifdefs.go\
  exif.go\
  exifheader.go\

include $(GOROOT)/src/Make.pkg

format:
	${GOFMT} -w ${GOFILES}

