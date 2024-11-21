package pgdb

import (
	"github.com/pennsieve/pennsieve-go-core/pkg/models/role"
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

func TestDbPermission_ImpliesRole(t *testing.T) {

	allRoles := []role.Role{role.None, role.Guest, role.Viewer, role.Editor, role.Manager, role.Owner}
	for _, testParams := range []struct {
		name            string
		permission      DbPermission
		expectedToImply map[role.Role]bool
	}{
		{
			"NoPermission implies",
			NoPermission,
			map[role.Role]bool{role.None: true},
		},
		{
			"Guest implies",
			Guest,
			map[role.Role]bool{role.None: true, role.Guest: true},
		},
		{
			"Read implies",
			Read,
			map[role.Role]bool{role.None: true, role.Guest: true, role.Viewer: true},
		},
		{
			"Write implies",
			Write,
			map[role.Role]bool{role.None: true, role.Guest: true, role.Viewer: true, role.Editor: true},
		},
		{
			"Delete implies",
			Delete,
			map[role.Role]bool{role.None: true, role.Guest: true, role.Viewer: true, role.Editor: true},
		},
		{
			"Administer implies",
			Administer,
			map[role.Role]bool{role.None: true, role.Guest: true, role.Viewer: true, role.Editor: true, role.Manager: true},
		},
		{
			"Owner implies",
			Owner,
			map[role.Role]bool{role.None: true, role.Guest: true, role.Viewer: true, role.Editor: true, role.Manager: true, role.Owner: true},
		},
	} {
		t.Run(testParams.name, func(t *testing.T) {
			for _, requiredRole := range allRoles {
				expected := testParams.expectedToImply[requiredRole]
				actual := testParams.permission.ImpliesRole(requiredRole)
				assert.Equal(t, expected, actual, "expected '%s implies %s' to be %v, not %v", testParams.permission, requiredRole, expected, actual)
			}
		})
	}
}
