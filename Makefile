# Copyright 2009 The Go Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include $(GOROOT)/src/Make.$(GOARCH)

TARG=glop

# Can have plain GOFILES too, but this example doesn't.

CGOFILES=\
	glop.go

CGO_LDFLAGS=-L/Users/jonathanwills/go-glop/lib -lglop -I. -framework Cocoa

# To add flags necessary for locating the library or its include files,
# set CGO_CFLAGS or CGO_LDFLAGS.  For example, to use an
# alternate installation of the library:
#	CGO_CFLAGS=-I/home/rsc/gmp32/include
#	CGO_LDFLAGS+=-L/home/rsc/gmp32/lib
# Note the += on the second line.

#CLEANFILES+=pi fib

include $(GOROOT)/src/Make.pkg

# Simple test programs

# Computes 1000 digits of pi; single-threaded.
#pi: install pi.go
#	$(GC) pi.go
#	$(LD) -o $@ pi.$O

# Computes 200 Fibonacci numbers; multi-threaded.
#fib: install fib.go
#	$(GC) fib.go
#	$(LD) -o $@ fib.$O

