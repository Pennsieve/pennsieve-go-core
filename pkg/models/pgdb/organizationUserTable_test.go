package pgdb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDbPermission_AsRoleString(t *testing.T) {
	assert.Equal(t, NoPermission.AsRoleString(), "none")
	assert.Equal(t, Guest.AsRoleString(), "guest")
	assert.Equal(t, Read.AsRoleString(), "viewer")
	assert.Equal(t, Write.AsRoleString(), "editor")
	assert.Equal(t, Delete.AsRoleString(), "editor")
	assert.Equal(t, Administer.AsRoleString(), "manager")
	assert.Equal(t, Owner.AsRoleString(), "owner")
	assert.Equal(t, DbPermission(999).AsRoleString(), "none")
}
