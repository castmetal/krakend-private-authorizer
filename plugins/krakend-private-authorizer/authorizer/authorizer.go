package authorizer

import (
	"encoding/json"
	"strings"

	"github.com/castmetal/krakend-private-authorizer/common"
	"github.com/valyala/fasthttp"
)

type Authorizer struct {
	Url           string
	Method        string
	Headers       map[string]string
	Params        []string
	RequestUri    string
	RequestMethod string
}
type AuthorizerResponse struct {
	StatusCode int
	Body       *[]byte
}

func NewAuthorizer(paramsRequest common.ParamsRequest, uri string, requestMethod string) *Authorizer {
	allowedMethods := make(map[string]bool, 5)
	allowedMethods = map[string]bool{
		"GET":    true,
		"POST":   true,
		"PUT":    true,
		"DELETE": true,
	}

	m := strings.ToUpper(requestMethod)

	if !allowedMethods[m] {
		m = "GET"
	}

	return &Authorizer{
		Url:           paramsRequest.AuthUrl,
		Method:        m,
		Headers:       paramsRequest.Headers,
		Params:        paramsRequest.Params,
		RequestUri:    uri,
		RequestMethod: requestMethod,
	}
}

func (a *Authorizer) DoRequest() AuthorizerResponse {
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("POST")
	req.Header.SetContentType(a.Headers["content-type"])
	req.SetRequestURI(a.Url)

	a.SetRequestBody(req)

	res := fasthttp.AcquireResponse()
	if err := fasthttp.Do(req, res); err != nil {
		panic("handle error")
	}

	fasthttp.ReleaseRequest(req)

	body := res.Body()
	statusCode := res.StatusCode()

	defer fasthttp.ReleaseResponse(res)

	return AuthorizerResponse{
		StatusCode: statusCode,
		Body:       &body,
	}
}

func (a *Authorizer) SetRequestBody(req *fasthttp.Request) {
	delete(a.Headers, "content-type")

	bodyStruct := a.Headers
	bodyStruct["resource_path"] = a.RequestUri
	bodyStruct["resource_method"] = a.RequestMethod

	bodyBytes, err := json.Marshal(bodyStruct)
	if err != nil {
		bodyBytes = []byte("")
	}

	req.SetBody(bodyBytes)
}
