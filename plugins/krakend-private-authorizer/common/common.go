package common

type ParamsRequest struct {
	Headers             map[string]string
	Params              []string
	AuthUrl             string
	AuthMethod          string
	ClientIdHeader      string
	ApiIdHeader         string
	DefaultErrorMessage string
	PublicFlag          string
	ModifyErrors        bool
	ErrorFlag           string
}

type ErrorResponse struct {
	Message string `json:"message"`
}
