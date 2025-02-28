package domain

type QueryType string

type QueryParams struct {
	QueryType QueryType
	Email     string
	Limit     int
	Offset    int
}

const (
	QueryTypeList  QueryType = "list"
	QueryTypeEmail QueryType = "email"
	DefaultLimit   int       = 10
	DefaultOffset  int       = 0
)
