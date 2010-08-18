include $(GOROOT)/src/Make.inc

TARG=glop

CGOFILES=glop.go

CGO_LDFLAGS=-lglop -framework Cocoa -framework OpenGL

include $(GOROOT)/src/Make.pkg

%: install %.go
	$(GC) $*.go
	$(LD) -o $@ $*.$O

