/*
Package group is made possible by the golang stdlib patch at
https://codereview.appspot.com/101310044/
*/
package group

import (
	"fmt"
	"runtime"
)

var (
	// ErrUnsupported is the error for when a function is unsupported on the
	// current OS.
	ErrUnsupported = fmt.Errorf(
		"Unsupported on %s_%s", runtime.GOOS, runtime.GOARCH)
)

// UnknownGroupIDError is returned by LookupID when a group cannot be found.
type UnknownGroupIDError string

func (e UnknownGroupIDError) Error() string {
	return "group: unknown groupid " + string(e)
}

// Group represents a group database entry.
//
// On posix systems Gid contains a decimal number
// representing the group gid.
type Group struct {
	// ID is the group's ID.
	ID string

	// Name is the group's name.
	Name string
}

// LookupGroupID looks up a group by a group's ID. If the group cannot be
// found, the returned error is of type UnknownGroupIdError.
func LookupGroupID(gid string) (*Group, error) {
	grp, err := lookupGroupID(gid)
	if err == ErrUnsupported {
		return &Group{
			ID:   gid,
			Name: "",
		}, nil
	}
	return grp, err
}
