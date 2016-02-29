// +build dragonfly freebsd !android,linux netbsd openbsd solaris
// +build cgo

package group

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnixLookupID(t *testing.T) {
	gid := "0"
	name := "root"
	grp, err := LookupGroupID(gid)
	assert.NoError(t, err)
	assert.NotNil(t, grp)
	assert.Equal(t, gid, grp.ID)
	assert.Equal(t, name, grp.Name)
}
