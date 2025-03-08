package domain

type QueryType string

const (
	QueryTypeList  QueryType = "list"
	QueryTypeEmail QueryType = "email"
	DefaultLimit   int       = 10
	DefaultOffset  int       = 0
)

type GetQueryParams struct {
	QueryType QueryType
	Email     string
	Limit     int
	Offset    int
}
