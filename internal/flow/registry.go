package flow

import "fmt"

var registry = make(map[string]Flow)

func Register(f Flow) {
	registry[f.Name()] = f
}

func Get(name string) (Flow, error) {
	f, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown flow: %s (available: %v)", name, List())
	}
	return f, nil
}

func List() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
