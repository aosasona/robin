package robin

import (
	"errors"
	"net/http"
	"strings"
)

// Technically this may not be a "builder" but the name sort of fits

type Instance struct {
	// Path to the generated typescript schema
	bindingsPath string

	// Internal pointer to the current robin instance
	robin *Robin
}

func (b *Instance) Handler() http.HandlerFunc {
	return b.robin.serveHTTP
}

func (b *Instance) ExportTsBindings(optPath ...string) error {
	path := b.bindingsPath

	if len(optPath) > 0 {
		path = optPath[0]
	}

	if strings.TrimSpace(path) == "" {
		return errors.New("no bindings export path provided, you can pass this to the `ExportTsBindings` method after calling `Build()` or as part of the opts during the instance creation with `robin.New(...)`")
	}

	return nil
}
