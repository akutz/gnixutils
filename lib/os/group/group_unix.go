// +build darwin dragonfly freebsd !android,linux netbsd openbsd solaris
// +build cgo

package group

import (
	"fmt"
	"runtime"
	"strconv"
	"syscall"
	"unsafe"
)

/*
#cgo solaris CFLAGS: -D_POSIX_PTHREAD_SEMANTICS
#include <unistd.h>
#include <sys/types.h>
#include <pwd.h>
#include <grp.h>
#include <stdlib.h>

static int mygetgrgid_r(int gid, struct group *grp,
           char *buf, size_t buflen, struct group **result) {
    return getgrgid_r(gid, grp, buf, buflen, result);
}
*/
import "C"

const (
	userBuffer = iota
	groupBuffer
)

func lookupGroupID(gid string) (*Group, error) {
	i, e := strconv.Atoi(gid)
	if e != nil {
		return nil, e
	}
	return lookupUnixGroup(i, buildGroup)
}

// LookupID looks up a group name based on its group ID.
func lookupUnixGroup(gid int, f func(*C.struct_group) *Group) (*Group, error) {

	var grp C.struct_group
	var result *C.struct_group

	buf, bufSize, err := allocBuffer(groupBuffer)
	if err != nil {
		return nil, err
	}
	defer C.free(buf)

	// mygetgrgid_r is a wrapper around getgrgid_r to
	// to avoid using gid_t because C.gid_t(gid) for
	// unknown reasons doesn't work on linux.
	rv := C.mygetgrgid_r(C.int(gid),
		&grp,
		(*C.char)(buf),
		C.size_t(bufSize),
		&result)
	if rv != 0 {
		return nil, fmt.Errorf(
			"group: lookup groupid %d: %s", gid, syscall.Errno(rv))
	}
	if result == nil {
		return nil, UnknownGroupIDError(gid)
	}

	g := f(&grp)
	return g, nil
}

func allocBuffer(bufType int) (unsafe.Pointer, C.long, error) {
	var bufSize C.long

	if runtime.GOOS == "freebsd" {
		// FreeBSD doesn't have _SC_GETPW_R_SIZE_MAX
		// or SC_GETGR_R_SIZE_MAX and just returns -1.
		// So just use the same size that Linux returns
		bufSize = 1024
	} else {
		var size C.int
		var constName string
		switch bufType {
		case userBuffer:
			size = C._SC_GETPW_R_SIZE_MAX
			constName = "_SC_GETPW_R_SIZE_MAX"
		case groupBuffer:
			size = C._SC_GETGR_R_SIZE_MAX
			constName = "_SC_GETGR_R_SIZE_MAX"
		}
		bufSize = C.sysconf(size)
		if bufSize <= 0 || bufSize > 1<<20 {
			return nil, bufSize,
				fmt.Errorf("user: unreasonable %s of %d", constName, bufSize)
		}
	}
	return C.malloc(C.size_t(bufSize)), bufSize, nil
}

func buildGroup(grp *C.struct_group) *Group {
	g := &Group{
		ID:   strconv.Itoa(int(grp.gr_gid)),
		Name: C.GoString(grp.gr_name),
	}
	return g
}
