include $(GOROOT)/src/Make.inc

TARG=glop

CGOFILES=glop_carbon.go
CGO_CFLAGS=-I inc2
#CGO_LDFLAGS=-lglop -L lib -framework Cocoa -framework OpenGL -mmacosx-version-min=10.4
CGO_LDFLAGS=-lglop -L lib -framework Carbon -framework OpenGL -framework IOKit -framework AGL -framework ApplicationServices

include $(GOROOT)/src/Make.pkg

%: install %.go
	$(GC) $*.go
	$(LD) -o $@ $*.$O

