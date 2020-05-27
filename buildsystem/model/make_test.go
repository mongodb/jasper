package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeVariantTask(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		t.Run("FailsWithoutRef", func(t *testing.T) {
			mvt := MakeVariantTask{}
			assert.Error(t, mvt.Validate())
		})
		t.Run("SucceedsIfNameSet", func(t *testing.T) {
			mvt := MakeVariantTask{Name: "task"}
			assert.NoError(t, mvt.Validate())
		})
		t.Run("SucceedsIfTagSet", func(t *testing.T) {
			mvt := MakeVariantTask{Tag: "tag"}
			assert.NoError(t, mvt.Validate())
		})
		t.Run("FailsIfNameAndTagSet", func(t *testing.T) {
			mvt := MakeVariantTask{
				Name: "task",
				Tag:  "tag",
			}
			assert.Error(t, mvt.Validate())
		})
	})
}

func TestMakeVariantParameters(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		t.Run("SucceedsWithName", func(t *testing.T) {
			mvp := MakeVariantParameters{
				Tasks: []MakeVariantTask{
					{Name: "name"},
				},
			}
			assert.NoError(t, mvp.Validate())
		})
		t.Run("SucceedsWithTag", func(t *testing.T) {
			mvp := MakeVariantParameters{
				Tasks: []MakeVariantTask{
					{Tag: "tag"},
				},
			}
			assert.NoError(t, mvp.Validate())
		})
		t.Run("SucceedsWithMultiple", func(t *testing.T) {
			mvp := MakeVariantParameters{
				Tasks: []MakeVariantTask{
					{Name: "name1"},
					{Name: "name2"},
					{Tag: "tag1"},
					{Tag: "tag2"},
				},
			}
			assert.NoError(t, mvp.Validate())
		})
		t.Run("FailsWithEmpty", func(t *testing.T) {
			mvp := MakeVariantParameters{}
			assert.Error(t, mvp.Validate())
		})
		t.Run("FailsWithDuplicateName", func(t *testing.T) {
			mvp := MakeVariantParameters{
				Tasks: []MakeVariantTask{
					{Name: "name"},
					{Name: "name"},
				},
			}
			assert.Error(t, mvp.Validate())
		})
		t.Run("FailsWithDuplicateTag", func(t *testing.T) {
			mvp := MakeVariantParameters{
				Tasks: []MakeVariantTask{
					{Tag: "tag"},
					{Tag: "tag"},
				},
			}
			assert.Error(t, mvp.Validate())
		})
	})
}

func TestNamedMakeVariantParameters(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		t.Run("Succeeds", func(t *testing.T) {
			nmvp := NamedMakeVariantParameters{
				Name: "variant",
				MakeVariantParameters: MakeVariantParameters{
					Tasks: []MakeVariantTask{
						{Name: "task"},
					},
				},
			}
			assert.NoError(t, nmvp.Validate())
		})
		t.Run("FailsWithEmpty", func(t *testing.T) {
			nmvp := NamedMakeVariantParameters{}
			assert.Error(t, nmvp.Validate())
		})
		t.Run("FailsWithoutName", func(t *testing.T) {
			nmvp := NamedMakeVariantParameters{
				MakeVariantParameters: MakeVariantParameters{
					Tasks: []MakeVariantTask{
						{Name: "task"},
					},
				},
			}
			assert.Error(t, nmvp.Validate())
		})
		t.Run("FailsWithInvalidParameters", func(t *testing.T) {
			nmvp := NamedMakeVariantParameters{
				Name: "variant",
			}
			assert.Error(t, nmvp.Validate())
		})
	})
}

func TestMakeVariant(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		for testName, testCase := range map[string]func(t *testing.T, v *MakeVariant){
			"Succeeds": func(t *testing.T, mv *MakeVariant) {
				assert.NoError(t, mv.Validate())
			},
			"FailsWithoutName": func(t *testing.T, mv *MakeVariant) {
				mv.Name = ""
				assert.Error(t, mv.Validate())
			},
			"FailsWithoutDistros": func(t *testing.T, mv *MakeVariant) {
				mv.Distros = nil
				assert.Error(t, mv.Validate())
			},
			"FailsWithoutTasks": func(t *testing.T, mv *MakeVariant) {
				mv.Tasks = nil
				assert.Error(t, mv.Validate())
			},
			"FailsWithInvalidTask": func(t *testing.T, mv *MakeVariant) {
				mv.Tasks = []MakeVariantTask{{}}
				assert.Error(t, mv.Validate())
			},
			"FailsWithDuplicateTaskName": func(t *testing.T, mv *MakeVariant) {
				mv.Tasks = []MakeVariantTask{
					{Name: "name"},
					{Name: "name"},
				}
				assert.Error(t, mv.Validate())
			},
			"FailsWithDuplicateTaskTag": func(t *testing.T, mv *MakeVariant) {
				mv.Tasks = []MakeVariantTask{
					{Tag: "tag"},
					{Tag: "tag"},
				}
				assert.Error(t, mv.Validate())
			},
		} {
			t.Run(testName, func(t *testing.T) {
				mv := MakeVariant{
					VariantDistro: VariantDistro{
						Name:    "var_name",
						Distros: []string{"distro1", "distro2"},
					},
					MakeVariantParameters: MakeVariantParameters{
						Tasks: []MakeVariantTask{
							{Name: "name"},
							{Tag: "tag"},
						},
					},
				}
				testCase(t, &mv)
			})
		}
	})
}

func TestMakeRuntimeOptions(t *testing.T) {
	t.Run("Merge", func(t *testing.T) {
		t.Run("ReturnsIdenticalWithNoArguments", func(t *testing.T) {
			mro := MakeRuntimeOptions([]string{"-i", "-k"})
			assert.Equal(t, mro, mro.Merge())
		})
		t.Run("ReturnsConcatenatedArguments", func(t *testing.T) {
			mro := MakeRuntimeOptions([]string{"-i", "-k"})
			merged := mro.Merge([]string{"-n"}, []string{"-w", "-k"})
			for i, expected := range []string{"-i", "-k", "-n", "-w", "-k"} {
				assert.Equal(t, expected, merged[i])
			}
		})
	})
}

func TestMakeTask(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		for testName, testCase := range map[string]func(t *testing.T, mt *MakeTask){
			"Succeeds": func(t *testing.T, mt *MakeTask) {
				assert.NoError(t, mt.Validate())
			},
			"FailsWithoutName": func(t *testing.T, mt *MakeTask) {
				mt.Name = ""
				assert.Error(t, mt.Validate())
			},
			"FailsWithoutTargets": func(t *testing.T, mt *MakeTask) {
				mt.Targets = nil
				assert.Error(t, mt.Validate())
			},
			"FailsWithDuplicateTags": func(t *testing.T, mt *MakeTask) {
				mt.Tags = []string{"tag1", "tag1"}
			},
		} {
			t.Run(testName, func(t *testing.T) {
				mt := MakeTask{
					Name: "name",
					Targets: []MakeTaskTarget{
						{Name: "target"},
						{Sequence: "sequence"},
					},
					Tags: []string{"tag1"},
				}
				testCase(t, &mt)
			})
		}
	})
}

func TestMakeTaskTarget(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		t.Run("SucceedsWithName", func(t *testing.T) {
			mtt := MakeTaskTarget{
				Name: "name",
			}
			assert.NoError(t, mtt.Validate())
		})
		t.Run("SucceedsWithSequence", func(t *testing.T) {
			mtt := MakeTaskTarget{
				Sequence: "sequence",
			}
			assert.NoError(t, mtt.Validate())
		})
		t.Run("FailsWithEmpty", func(t *testing.T) {
			mtt := MakeTaskTarget{}
			assert.Error(t, mtt.Validate())
		})
		t.Run("FailsWithNameAndSequence", func(t *testing.T) {
			mtt := MakeTaskTarget{
				Name:     "name",
				Sequence: "sequence",
			}
			assert.Error(t, mtt.Validate())
		})
	})
}

func TestMakeTargetSequence(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		t.Run("Succeeds", func(t *testing.T) {
			mts := MakeTargetSequence{
				Name:    "name",
				Targets: []string{"target"},
			}
			assert.NoError(t, mts.Validate())
		})
		t.Run("FailsWithEmpty", func(t *testing.T) {
			mts := MakeTargetSequence{}
			assert.Error(t, mts.Validate())
		})
		t.Run("FailsWithoutName", func(t *testing.T) {
			mts := MakeTargetSequence{
				Targets: []string{"target"},
			}
			assert.Error(t, mts.Validate())
		})
		t.Run("FailsWithoutTargets", func(t *testing.T) {
			mts := MakeTargetSequence{
				Name: "name",
			}
			assert.Error(t, mts.Validate())
		})
	})
}

func TestMakeGetVariantIndexByName(t *testing.T) {
	for testName, testCase := range map[string]func(t *testing.T, m *Make){
		"Succeeds": func(t *testing.T, m *Make) {
			mv, i, err := m.GetVariantIndexByName("variant")
			require.NoError(t, err)
			assert.Equal(t, 0, i)
			assert.Equal(t, "variant", mv.Name)
		},
		"FailsIfVariantNotFound": func(t *testing.T, m *Make) {
			mv, i, err := m.GetVariantIndexByName("foo")
			assert.Error(t, err)
			assert.Equal(t, -1, i)
			assert.Zero(t, mv)
		},
	} {
		t.Run(testName, func(t *testing.T) {
			m := Make{
				Variants: []MakeVariant{
					{VariantDistro: VariantDistro{Name: "variant"}},
				},
			}
			testCase(t, &m)
		})
	}
}

func TestMakeGetTasksByTag(t *testing.T) {
	for testName, testCase := range map[string]func(t *testing.T, m *Make){
		"Succeeds": func(t *testing.T, m *Make) {
			mts := m.GetTasksByTag("tag2")
			require.Len(t, mts, 1)
			assert.Equal(t, "task1", mts[0].Name)
		},
		"FailsIfTaskNotFound": func(t *testing.T, m *Make) {
			mts := m.GetTasksByTag("foo")
			assert.Empty(t, mts)
		},
	} {
		t.Run(testName, func(t *testing.T) {
			g := Make{
				Tasks: []MakeTask{
					{Name: "task1", Tags: []string{"tag1", "tag2"}},
					{Name: "task2", Tags: []string{"tag1", "tag3"}},
				},
			}
			testCase(t, &g)
		})
	}
}

func TestMakeGetTaskIndexByName(t *testing.T) {
	for testName, testCase := range map[string]func(t *testing.T, m *Make){
		"Succeeds": func(t *testing.T, m *Make) {
			mt, i, err := m.GetTaskIndexByName("task1")
			require.NoError(t, err)
			assert.Equal(t, 0, i)
			assert.Equal(t, "task1", mt.Name)
		},
		"FailsIfTaskNotFound": func(t *testing.T, m *Make) {
			mt, i, err := m.GetTaskIndexByName("foo")
			assert.Error(t, err)
			assert.Equal(t, -1, i)
			assert.Zero(t, mt)
		},
	} {
		t.Run(testName, func(t *testing.T) {
			m := Make{
				Tasks: []MakeTask{
					{Name: "task1"},
					{Name: "task2"},
				},
			}
			testCase(t, &m)
		})
	}
}

func TestMakeGetTasksFromRef(t *testing.T) {
	for testName, testCase := range map[string]func(t *testing.T, m *Make){
		"SucceedsWithName": func(t *testing.T, m *Make) {
			mts, err := m.GetTasksFromRef(MakeVariantTask{Name: "task1"})
			require.NoError(t, err)
			require.Len(t, mts, 1)
			assert.Equal(t, m.Tasks[0], mts[0])
		},
		"SucceedsWithTag": func(t *testing.T, m *Make) {
			mts, err := m.GetTasksFromRef(MakeVariantTask{Tag: "tag"})
			require.NoError(t, err)
			require.Len(t, mts, 2)
			assert.Equal(t, m.Tasks[0], mts[0])
		},
		"FailsWithEmpty": func(t *testing.T, m *Make) {
			mts, err := m.GetTasksFromRef(MakeVariantTask{})
			assert.Error(t, err)
			assert.Zero(t, mts)
		},
		"FailsWithUnmatchedName": func(t *testing.T, m *Make) {
			mts, err := m.GetTasksFromRef(MakeVariantTask{Name: "foo"})
			assert.Error(t, err)
			assert.Zero(t, mts)
		},
		"FailsWithUnmatchedTag": func(t *testing.T, m *Make) {
			mts, err := m.GetTasksFromRef(MakeVariantTask{Tag: "foo"})
			assert.Error(t, err)
			assert.Zero(t, mts)
		},
	} {
		t.Run(testName, func(t *testing.T) {
			m := Make{
				Tasks: []MakeTask{
					{Name: "task1", Tags: []string{"tag"}},
					{Name: "task2", Tags: []string{"tag"}},
				},
			}
			testCase(t, &m)
		})
	}
}

func TestGetTargetSequenceIndexByName(t *testing.T) {
	for testName, testCase := range map[string]func(t *testing.T, m *Make){
		"Succeeds": func(t *testing.T, m *Make) {
			mts, i, err := m.GetTargetSequenceIndexByName("sequence")
			require.NoError(t, err)
			assert.Equal(t, 0, i)
			assert.Equal(t, m.TargetSequences[0], *mts)
		},
		"FailsIfTaskNotFound": func(t *testing.T, m *Make) {
			mts, i, err := m.GetTaskIndexByName("foo")
			assert.Error(t, err)
			assert.Equal(t, -1, i)
			assert.Zero(t, mts)
		},
	} {
		t.Run(testName, func(t *testing.T) {
			m := Make{
				TargetSequences: []MakeTargetSequence{
					{Name: "sequence", Targets: []string{"target1"}},
				},
			}
			testCase(t, &m)
		})
	}
}

func TestMakeGetTargetsFromRef(t *testing.T) {
	for testName, testCase := range map[string]func(t *testing.T, m *Make){
		"SucceedsWithName": func(t *testing.T, m *Make) {
			mts, err := m.GetTasksFromRef(MakeVariantTask{Name: "task1"})
			require.NoError(t, err)
			require.Len(t, mts, 1)
			assert.Equal(t, m.Tasks[0], mts[0])
		},
		"SucceedsWithTag": func(t *testing.T, m *Make) {
			mts, err := m.GetTasksFromRef(MakeVariantTask{Tag: "tag"})
			require.NoError(t, err)
			require.Len(t, mts, 2)
			assert.Equal(t, m.Tasks[0], mts[0])
		},
		"FailsWithEmpty": func(t *testing.T, m *Make) {
			mts, err := m.GetTasksFromRef(MakeVariantTask{})
			assert.Error(t, err)
			assert.Zero(t, mts)
		},
		"FailsWithUnmatchedName": func(t *testing.T, m *Make) {
			mts, err := m.GetTasksFromRef(MakeVariantTask{Name: "foo"})
			assert.Error(t, err)
			assert.Zero(t, mts)
		},
		"FailsWithUnmatchedTag": func(t *testing.T, m *Make) {
			mts, err := m.GetTasksFromRef(MakeVariantTask{Tag: "foo"})
			assert.Error(t, err)
			assert.Zero(t, mts)
		},
	} {
		t.Run(testName, func(t *testing.T) {
			m := Make{
				Tasks: []MakeTask{
					{Name: "task1", Tags: []string{"tag"}},
					{Name: "task2", Tags: []string{"tag"}},
				},
			}
			testCase(t, &m)
		})
	}
}

func TestMakeValidate(t *testing.T) {
	for testName, testCase := range map[string]func(t *testing.T, m *Make){
		"Succeeds": func(t *testing.T, m *Make) {
			assert.NoError(t, m.Validate())
		},
		"SucceedsWithValidTargetSequenceName": func(t *testing.T, m *Make) {
			m.TargetSequences = []MakeTargetSequence{
				{
					Name:    "sequence",
					Targets: []string{"target1"},
				},
			}
			m.Tasks = []MakeTask{
				{
					Name:    "task",
					Targets: []MakeTaskTarget{},
				},
			}
		},
		"FailsWithInvalidTargetSequenceName": func(t *testing.T, m *Make) {
			m.TargetSequences = nil
			m.Tasks = []MakeTask{
				{
					Name: "task",
					Targets: []MakeTaskTarget{
						{Sequence: "foo"},
					},
					Tags: []string{"tag1"},
				},
			}
			assert.Error(t, m.Validate())
		},
		"FailsWithoutTasks": func(t *testing.T, m *Make) {
			m.Tasks = nil
			assert.Error(t, m.Validate())
		},
		"FailsWithInvalidTask": func(t *testing.T, m *Make) {
			m.Tasks = []MakeTask{{}}
			assert.Error(t, m.Validate())
		},
		"FailsWithDuplicateTaskName": func(t *testing.T, m *Make) {
			m.Tasks = []MakeTask{
				{
					Name: "name",
					Targets: []MakeTaskTarget{
						{Name: "target1"},
					},
				}, {
					Name: "name",
					Targets: []MakeTaskTarget{
						{Name: "target2"},
					},
				},
			}
			assert.Error(t, m.Validate())
		},
		"FailsWithoutVariants": func(t *testing.T, m *Make) {
			m.Variants = nil
			assert.Error(t, m.Validate())
		},
		"FailsWithInvalidVariant": func(t *testing.T, m *Make) {
			m.Variants = []MakeVariant{{}}
			assert.Error(t, m.Validate())
		},
		"FailsWithDuplicateVariantName": func(t *testing.T, m *Make) {
			m.Variants = []MakeVariant{
				{
					VariantDistro: VariantDistro{
						Name:    "variant",
						Distros: []string{"distro"},
					},
					MakeVariantParameters: MakeVariantParameters{
						Tasks: []MakeVariantTask{
							{Name: "task"},
						},
					},
				}, {
					VariantDistro: VariantDistro{
						Name:    "variant",
						Distros: []string{"distro"},
					},
					MakeVariantParameters: MakeVariantParameters{
						Tasks: []MakeVariantTask{
							{Name: "task"},
						},
					},
				},
			}
			assert.Error(t, m.Validate())
		},
		"SucceedsWithValidVariantTaskName": func(t *testing.T, m *Make) {
			m.Tasks = []MakeTask{
				{
					Name: "task",
					Targets: []MakeTaskTarget{
						{Name: "target1"},
					},
				},
			}
			m.Variants = []MakeVariant{
				{
					VariantDistro: VariantDistro{
						Name:    "variant",
						Distros: []string{"distro"},
					},
					MakeVariantParameters: MakeVariantParameters{
						Tasks: []MakeVariantTask{
							{Name: "task"},
						},
					},
				},
			}
			assert.NoError(t, m.Validate())
		},
		"FailsWithInvalidVariantTaskName": func(t *testing.T, m *Make) {
			m.Variants = []MakeVariant{
				{
					VariantDistro: VariantDistro{
						Name:    "variant",
						Distros: []string{"distro"},
					},
					MakeVariantParameters: MakeVariantParameters{
						Tasks: []MakeVariantTask{
							{Name: "foo"},
						},
					},
				},
			}
			assert.Error(t, m.Validate())
		},
		"FailsWithDuplicateMakeTaskReferences": func(t *testing.T, m *Make) {
			m.Tasks = []MakeTask{
				{
					Name: "task",
					Tags: []string{"tag"},
				},
			}
			m.Variants = []MakeVariant{
				{
					VariantDistro: VariantDistro{
						Name:    "variant",
						Distros: []string{"distro"},
					},
					MakeVariantParameters: MakeVariantParameters{
						Tasks: []MakeVariantTask{
							{Name: "task"},
							{Tag: "tag"},
						},
					},
				},
			}
			assert.Error(t, m.Validate())
		},
		"SucceedsWithValidVariantTaskTag": func(t *testing.T, m *Make) {
			m.Tasks = []MakeTask{
				{
					Name: "task",
					Targets: []MakeTaskTarget{
						{Sequence: "sequence1"},
					},
					Tags: []string{"tag"},
				},
			}
			m.Variants = []MakeVariant{
				{
					VariantDistro: VariantDistro{
						Name:    "variant",
						Distros: []string{"distro"},
					},
					MakeVariantParameters: MakeVariantParameters{
						Tasks: []MakeVariantTask{
							{Tag: "tag"},
						},
					},
				},
			}
			assert.NoError(t, m.Validate())
		},
		"FailsWithInvalidVariantTaskTag": func(t *testing.T, m *Make) {
			m.Variants = []MakeVariant{
				{
					VariantDistro: VariantDistro{
						Name:    "variant",
						Distros: []string{"distro"},
					},
					MakeVariantParameters: MakeVariantParameters{
						Tasks: []MakeVariantTask{
							{Tag: "nonexistent"},
						},
					},
				},
			}
			assert.Error(t, m.Validate())
		},
	} {
		t.Run(testName, func(t *testing.T) {
			m := Make{
				TargetSequences: []MakeTargetSequence{
					{
						Name:    "sequence1",
						Targets: []string{"target1"},
					}, {
						Name:    "sequence2",
						Targets: []string{"target1", "target3"},
					},
				},
				Tasks: []MakeTask{
					{
						Name: "task",
						Targets: []MakeTaskTarget{
							{Name: "target2"},
							{Sequence: "sequence2"},
						},
						Tags: []string{"tag1", "tag2"},
					},
				},
				Variants: []MakeVariant{
					{
						VariantDistro: VariantDistro{
							Name:    "variant",
							Distros: []string{"distro"},
						},
						MakeVariantParameters: MakeVariantParameters{
							Tasks: []MakeVariantTask{
								{Name: "task"},
							},
						},
					},
				},
			}
			testCase(t, &m)
		})
	}
}

// kim: TODO: test merging
