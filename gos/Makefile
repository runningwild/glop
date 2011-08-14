include $(GOROOT)/src/Make.inc

TARG=glop/gos

CGOFILES=cgo_os.go
CGO_LDFLAGS=-lglop -framework Cocoa -framework OpenGL -mmacosx-version-min=10.5

include $(GOROOT)/src/Make.pkg

%: install %.go
	$(GC) $*.go
	$(LD) -o $@ $*.$O

