package hook

import (
	"testing"

	"github.com/loft-sh/devspace/pkg/devspace/config/versions/latest"
	"github.com/loft-sh/devspace/pkg/util/log"
)

func TestHookWithoutExecution(t *testing.T) {
	//Execute 0 hooks
	executer := NewExecuter(&latest.Config{})
	err := executer.Execute("", "", "", Context{}, log.Discard)
	if err != nil {
		t.Fatalf("Failed to execute 0 hooks with error: %v", err)
	}

	//Execute 1 hook with no when
	executer = NewExecuter(&latest.Config{
		Hooks: []*latest.HookConfig{
			&latest.HookConfig{},
		},
	})
	err = executer.Execute("", "", "", Context{}, log.Discard)
	if err != nil {
		t.Fatalf("Failed to execute 1 hook without when with error: %v", err)
	}

	//Execute 1 hook with no When.Before and no When.After
	executer = NewExecuter(&latest.Config{
		Hooks: []*latest.HookConfig{
			&latest.HookConfig{
				When: &latest.HookWhenConfig{},
			},
		},
	})
	err = executer.Execute("", "", "", Context{}, log.Discard)
	if err != nil {
		t.Fatalf("Failed to execute 1 hook without When.Before and When.After with error: %v", err)
	}

	//Execute 1 hook with empty When.Before
	executer = NewExecuter(&latest.Config{
		Hooks: []*latest.HookConfig{
			&latest.HookConfig{
				When: &latest.HookWhenConfig{
					Before: &latest.HookWhenAtConfig{},
				},
			},
		},
	})
	err = executer.Execute(Before, "", "", Context{}, log.Discard)
	if err != nil {
		t.Fatalf("Failed to execute 1 hook with empty When.Before: %v", err)
	}

	//Execute 1 hook with empty When.After
	executer = NewExecuter(&latest.Config{
		Hooks: []*latest.HookConfig{
			&latest.HookConfig{
				When: &latest.HookWhenConfig{
					After: &latest.HookWhenAtConfig{},
				},
			},
		},
	})
	err = executer.Execute(After, "", "", Context{}, log.Discard)
	if err != nil {
		t.Fatalf("Failed to execute 1 hook with empty When.After: %v", err)
	}

}

func TestHookWithExecution(t *testing.T) {
	executer := NewExecuter(&latest.Config{
		Hooks: []*latest.HookConfig{
			&latest.HookConfig{
				When: &latest.HookWhenConfig{
					Before: &latest.HookWhenAtConfig{
						Deployments: "theseDeployments",
					},
				},
				Command: "echo",
				Args:    []string{"hello"},
			},
		},
	})
	err := executer.Execute(Before, StageDeployments, "theseDeployments", Context{}, log.Discard)
	if err != nil {
		t.Fatalf("Failed to execute 1 hook with empty When.After: %v", err)
	}

}
