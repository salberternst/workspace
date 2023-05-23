package builder

import (
	"strings"
)

type Values struct {
	values map[string]interface{}
}

func NewValues() Values {
	return Values{
		values: map[string]interface{}{},
	}
}

func (o *Values) set(value interface{}, paths ...string) {
	var currentPath interface{}

	currentPath = o.values

	for index, path := range paths {
		node := currentPath.(map[string]interface{})
		if index < len(paths)-1 {
			if node[path] == nil {
				node[path] = map[string]interface{}{}
			}
			currentPath = node[path]
		} else {
			node[path] = value
		}
	}
}

func (o *Values) Set(value interface{}, path string) {
	o.set(value, strings.Split(path, ".")...)
}

func (o *Values) GetMap() map[string]interface{} {
	return o.values
}
