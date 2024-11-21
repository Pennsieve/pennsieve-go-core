package role

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRole_Implies(t *testing.T) {

	for scenario, testFunc := range map[string]func(t *testing.T){
		"role None only implies None":                      testNoneImplies,
		"every role implies None":                          testEverythingImpliesNone,
		"role Owner implies every role":                    testOwnerImpliesEveryRole,
		"only Owner implies Owner":                         testOnlyOwnerImpliesOwner,
		"role Viewer only implies itself, Guest, and None": testViewerImplies,
	} {
		t.Run(scenario, func(t *testing.T) {
			testFunc(t)
		})
	}
}

func testNoneImplies(t *testing.T) {
	assert.True(t, None.Implies(None))

	assert.False(t, None.Implies(Guest))
	assert.False(t, None.Implies(Viewer))
	assert.False(t, None.Implies(Editor))
	assert.False(t, None.Implies(Manager))
	assert.False(t, None.Implies(Owner))
}

func testEverythingImpliesNone(t *testing.T) {
	assert.True(t, None.Implies(None))
	assert.True(t, Guest.Implies(None))
	assert.True(t, Viewer.Implies(None))
	assert.True(t, Editor.Implies(None))
	assert.True(t, Manager.Implies(None))
	assert.True(t, Owner.Implies(None))
}

func testOwnerImpliesEveryRole(t *testing.T) {
	assert.True(t, Owner.Implies(None))
	assert.True(t, Owner.Implies(Guest))
	assert.True(t, Owner.Implies(Viewer))
	assert.True(t, Owner.Implies(Editor))
	assert.True(t, Owner.Implies(Manager))
	assert.True(t, Owner.Implies(Owner))
}

func testViewerImplies(t *testing.T) {
	assert.True(t, Viewer.Implies(None))
	assert.True(t, Viewer.Implies(Guest))
	assert.True(t, Viewer.Implies(Viewer))

	assert.False(t, Viewer.Implies(Editor))
	assert.False(t, Viewer.Implies(Manager))
	assert.False(t, Viewer.Implies(Owner))
}

func testOnlyOwnerImpliesOwner(t *testing.T) {
	assert.False(t, None.Implies(Owner))
	assert.False(t, Guest.Implies(Owner))
	assert.False(t, Viewer.Implies(Owner))
	assert.False(t, Editor.Implies(Owner))
	assert.False(t, Manager.Implies(Owner))

	assert.True(t, Owner.Implies(Owner))
}
