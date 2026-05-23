package api

// HttpMethod is a type used to enforce method correctness in Route
type HttpMethod int

const (
	HttpMethodGet HttpMethod = iota
	HttpMethodPost
	HttpMethodPut
	HttpMethodDelete
	HttpMethodInvalid
)

func (m HttpMethod) String() string {
	switch m {
	case HttpMethodGet:
		return "GET"
	case HttpMethodPost:
		return "POST"
	case HttpMethodPut:
		return "PUT"
	case HttpMethodDelete:
		return "DELETE"
	default:
		return "INVALID"
	}
}

func (m HttpMethod) Valid() bool {
	return m < HttpMethodInvalid
}
