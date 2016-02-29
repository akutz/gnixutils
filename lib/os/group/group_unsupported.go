// +build !darwin !dragonfly !freebsd !android !linux !netbsd !openbsd !solaris
// +build !cgo

package group

func lookupGroupID(gid string) (*Group, error) {
	return nil, ErrUnsupported
}
