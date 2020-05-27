package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVariantDistro(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		t.Run("Succeeds", func(t *testing.T) {
			vd := VariantDistro{
				Name:    "variant",
				Distros: []string{"ubuntu"},
			}
			assert.NoError(t, vd.Validate())
		})
		t.Run("FailsForEmpty", func(t *testing.T) {
			vd := VariantDistro{}
			assert.Error(t, vd.Validate())
		})
		t.Run("FailsWithoutName", func(t *testing.T) {
			vd := VariantDistro{
				Distros: []string{"ubuntu"},
			}
			assert.Error(t, vd.Validate())
		})
		t.Run("FailsWithoutDistros", func(t *testing.T) {
			vd := VariantDistro{
				Name: "variant",
			}
			assert.Error(t, vd.Validate())
		})
	})
}

func TestMergeEnvironments(t *testing.T) {
	t.Run("AddsUniqueVars", func(t *testing.T) {
		envs := []map[string]string{
			{"foo": "bar"},
			{"bat": "baz", "qux": "quux"},
		}
		env := MergeEnvironments(envs...)
		assert.Len(t, env, 3)
		assert.Equal(t, "bar", env["foo"])
		assert.Equal(t, "baz", env["bat"])
		assert.Equal(t, "quux", env["qux"])
	})
	t.Run("OverwritesExistingByOrder", func(t *testing.T) {
		envs := []map[string]string{
			{"foo": "bar"},
			{"foo": "bat"},
			{"foo": "baz"},
		}
		env := MergeEnvironments(envs...)
		assert.Equal(t, "baz", env["foo"])
	})
}
