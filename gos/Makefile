include $(GOROOT)/src/Make.inc

TARG=glop/gos

ifeq ($(GOOS),darwin)
CGOFILES=cgo_darwin.go
CGO_LDFLAGS=-Ldarwin/lib -lglop -framework Cocoa -framework OpenGL -mmacosx-version-min=10.5
endif 

ifeq ($(GOOS),windows)
CGOFILES=cgo_windows.go
CGO_LDFLAGS=-Lwindows/lib -lglop
endif

ifeq ($(GOOS),linux)
CGOFILES=cgo_linux.go
CGO_LDFLAGS=-Llinux/lib -lglop -lX11 -lGL
endif

# gb: this is the local install
GBROOT=..

# gb: compile/link against local install
GC+= -I $(GBROOT)/_obj
LD+= -L $(GBROOT)/_obj
# gb: compile/link against GOPATH entries
GC+= -I $(GOROOT)/pkg/$(GOOS)_$(GOARCH)
LD+= -L $(GOROOT)/pkg/$(GOOS)_$(GOARCH)

# gb: copy to local install
$(GBROOT)/_obj/$(TARG).a: _obj/$(TARG).a
	mkdir -p $(dir $@); cp -f $< $@

package: $(GBROOT)/_obj/$(TARG).a

include $(GOROOT)/src/Make.pkg

