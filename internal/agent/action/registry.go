package action

import (
	"fmt"
	"sort"
	"strings"
)

var builtinActions []Action

func RegisterBuiltin(action Action) {
	if action == nil {
		return
	}

	builtinActions = append(builtinActions, action)
}

type Registry struct {
	actions map[string]Action
}

func NewRegistry(actions ...Action) *Registry {
	registry := &Registry{
		actions: make(map[string]Action),
	}

	for _, action := range actions {
		registry.Register(action)
	}

	return registry
}

func NewDefaultRegistry() *Registry {
	registry := NewRegistry()

	for _, action := range builtinActions {
		registry.Register(action)
	}

	return registry
}

func (r *Registry) Register(action Action) {
	if action == nil {
		return
	}

	name := NormalizeName(action.Name())
	if name == "" {
		return
	}

	r.actions[name] = action
}

func (r *Registry) Get(name string) (Action, bool) {
	if r == nil {
		return nil, false
	}

	action, ok := r.actions[NormalizeName(name)]
	return action, ok
}

func (r *Registry) MustGet(name string) (Action, error) {
	handler, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("unknown action: %s", name)
	}

	return handler, nil
}

func (r *Registry) HasAction(name string) bool {
	_, ok := r.Get(name)
	return ok
}

func (r *Registry) List() []Info {
	if r == nil {
		return nil
	}

	result := make([]Info, 0, len(r.actions))

	for _, handler := range r.actions {
		result = append(result, Info{
			Name:        handler.Name(),
			Description: handler.Description(),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

func NormalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
