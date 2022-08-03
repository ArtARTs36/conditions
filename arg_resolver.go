package conditions

import "fmt"

type ArgResolver interface {
	Resolve(key string) (interface{}, error)
}

type MapArgResolver struct {
	args map[string]interface{}
}

func NewMapArgResolver(args map[string]interface{}) ArgResolver {
	return &MapArgResolver{args: args}
}

func (r *MapArgResolver) Resolve(key string) (interface{}, error) {
	arg, exists := r.args[key]

	if exists {
		return arg, nil
	}

	return nil, fmt.Errorf("argument by key %s not found", key)
}
