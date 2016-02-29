// +build darwin
// +build cgo

package group

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDarwinLookupID(t *testing.T) {
	gid := "0"
	name := "wheel"
	grp, err := LookupGroupID(gid)
	assert.NoError(t, err)
	assert.NotNil(t, grp)
	assert.Equal(t, gid, grp.ID)
	assert.Equal(t, name, grp.Name)
}
