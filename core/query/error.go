package query

type QueryError struct {
	Message string
}

func (e *QueryError) Error() string {
	return e.Message
}
