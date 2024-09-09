package generator

import (
	_ "go.trulyao.dev/mirror/v2"
)

type Schema[Request, Response any] struct {
	Type           string   `mirror:"name:_type,type:'query'|'mutation'"`
	Name           string   `mirror:"name"`
	RequestSchema  Request  `mirror:"name:request"`
	ResponseSchema Response `mirror:"name:response"`
}

type Exported[Request, Response any] struct {
	Procedures []Schema[Request, Response] `mirror:"name:procedures"`
}
