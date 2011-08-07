include $(GOROOT)/src/Make.inc

TARG=glop

CGOFILES=glop.go
CGO_LDFLAGS=-lglop -L lib -framework Cocoa -framework OpenGL -mmacosx-version-min=10.4

include $(GOROOT)/src/Make.pkg

%: install %.go
	$(GC) $*.go
	$(LD) -o $@ $*.$O

