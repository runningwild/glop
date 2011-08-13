include $(GOROOT)/src/Make.inc

TARG=glop

CGOFILES=glop.go
CGO_CFLAGS=-I include
CGO_LDFLAGS=-arch x86_64 -lglop -Llib -framework Cocoa -framework OpenGL -mmacosx-version-min=10.4

include $(GOROOT)/src/Make.pkg

%: install %.go
	$(GC) $*.go
	$(LD) -o $@ $*.$O

