package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/castmetal/krakend-private-authorizer/authorizer"
	"github.com/castmetal/krakend-private-authorizer/common"
	uuid "github.com/satori/go.uuid"
)

// download code on github.com/castmetal/krakend-private-authorizer
const Namespace = "krakend-private-authorizer"

type statusRecorder struct {
	http.ResponseWriter
	status  int
	buf     *bytes.Buffer
	written bool
}

// HandlerRegisterer is the symbol the plugin loader will try to load. It must implement the Registerer interface
var HandlerRegisterer = registerer(Namespace)

type registerer string

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

func (registerer) RegisterLogger(v interface{}) {
	l, ok := v.(common.Logger)
	if !ok {
		return
	}
	common.FireLogger = l

	common.FireLogger.Debug(fmt.Sprintf("[PLUGIN: %s] Logger loaded", HandlerRegisterer))
}

func (r registerer) registerHandlers(ctx context.Context, extra map[string]interface{}, handler http.Handler) (http.Handler, error) {
	// return the actual handler wrapping or your custom logic so it can be used as a replacement for the default http handler
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var requestId string

		requestArgs := r.getParamsToRequest(extra, req)

		w.Header().Set("Content-Type", "application/json")

		requestId = uuid.NewV4().String()

		uriSplit := strings.Split(req.RequestURI, "?")
		uri := uriSplit[0]

		newReq := req.WithContext(ctx)
		newReq.Header.Set("x-origin", uri)
		newReq.Header.Set("x-request-id", requestId)
		newReq.Header.Set("x-remote-ip", req.RemoteAddr)

		auth := r.InterceptWithAuthorizer(w, req, requestArgs)
		if !auth {
			return
		}

		rec := statusRecorder{w, 401, &bytes.Buffer{}, false}

		handler.ServeHTTP(&rec, newReq)

		m := make(map[string]interface{})
		json.Unmarshal(rec.buf.Bytes(), &m)

		messageError := make(map[string]interface{})
		status := rec.status

		if requestArgs.ModifyErrors == true && m[requestArgs.ErrorFlag] != nil {
			mapMessage := m[requestArgs.ErrorFlag].(map[string]interface{})
			statusFloat := mapMessage["http_status_code"].(float64)
			messageStr := mapMessage["http_body"].(string)

			err := json.Unmarshal([]byte(messageStr), &messageError)
			if err != nil {
				messageError = make(map[string]interface{})
			}

			status = int(statusFloat)
		}

		w.WriteHeader(status)

		if status >= 500 {
			resBody := common.ErrorResponse{
				Message: requestArgs.DefaultErrorMessage,
			}

			json.NewEncoder(w).Encode(resBody)
			return
		} else if status >= 300 {
			json.NewEncoder(w).Encode(messageError)
			return
		}

		json.NewEncoder(w).Encode(m)
	}), nil
}

func (r registerer) InterceptWithAuthorizer(w http.ResponseWriter, req *http.Request, requestArgs common.ParamsRequest) bool {
	if strings.Contains(req.RequestURI, requestArgs.PublicFlag) {
		return true
	}

	uriSplit := strings.Split(req.RequestURI, "?")
	uri := uriSplit[0]

	authorizer := authorizer.NewAuthorizer(requestArgs, uri, req.Method)
	response := authorizer.DoRequest()

	authResponse := make(map[string]interface{})

	err := json.Unmarshal(*response.Body, &authResponse)
	if err != nil {
		authResponse["message"] = requestArgs.DefaultErrorMessage
	}

	if response.StatusCode <= 299 {
		return true
	}

	w.WriteHeader(response.StatusCode)
	json.NewEncoder(w).Encode(authResponse)

	return false
}

func (r registerer) getParamsToRequest(extra map[string]interface{}, req *http.Request) common.ParamsRequest {
	var paramsMap map[string]interface{}

	paramsMap = extra[Namespace].(map[string]interface{})

	tokenHeader := fmt.Sprint(paramsMap["token_header"])
	apiIdHeader := fmt.Sprint(paramsMap["api_id_header"])
	clientIdHeader := fmt.Sprint(paramsMap["client_id_header"])
	headers := make(map[string]string)

	headers["content-type"] = "application/json"
	headers[tokenHeader] = req.Header.Get(tokenHeader)
	headers[apiIdHeader] = req.Header.Get(apiIdHeader)
	headers[clientIdHeader] = req.Header.Get(clientIdHeader)

	params := paramsMap["params"].([]interface{})

	paramsArr := make([]string, len(params))
	for i, v := range params {
		paramsArr[i] = fmt.Sprint(v)
	}

	modifyErrors := paramsMap["modify_errors"].(bool)

	return common.ParamsRequest{
		Params:              paramsArr,
		Headers:             headers,
		AuthUrl:             fmt.Sprint(paramsMap["auth_url"]),
		AuthMethod:          fmt.Sprint(paramsMap["auth_method"]),
		ApiIdHeader:         fmt.Sprint(paramsMap["api_id_header"]),
		ClientIdHeader:      fmt.Sprint(paramsMap["client_id_header"]),
		PublicFlag:          fmt.Sprint(paramsMap["public_flag"]),
		DefaultErrorMessage: fmt.Sprint(paramsMap["default_error_message"]),
		ErrorFlag:           fmt.Sprintf("error_%s", paramsMap["error_flag"]),
		ModifyErrors:        modifyErrors,
	}
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.written = true
	rec.status = code
}

func (rec *statusRecorder) Write(p []byte) (int, error) {
	return rec.buf.Write(p)
}

func main() {
}
