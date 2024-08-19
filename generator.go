package robin

type Schema[Request, Response any] struct {
	Type           ProcedureType `mirror:"name:_type,type:'query'|'mutation'"`
	Name           string        `mirror:"name"`
	RequestSchema  Request       `mirror:"name:requestSchema"`
	ResponseSchema Response      `mirror:"name:responseSchema"`
}

type Exported[Request, Response any] struct {
	Procedures []Schema[Request, Response] `mirror:"name:procedures"`
}
