package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// APIClient struct
type APIClient struct {
	logger     *zap.SugaredLogger
	httpClient *http.Client
}

// NewAPIRequestClient constructor
func NewAPIRequestClient(logger *zap.SugaredLogger) APIClient {
	return APIClient{
		logger:     logger,
		httpClient: &http.Client{},
	}
}

// APIHeader header for request
type APIHeader struct {
	Name  string
	Value []string
}

// RequestParams input for Http request
type RequestParams struct {
	URL    string
	Header []APIHeader
	Body   interface{}
}

// Get create http get request
func (r *APIClient) Get(params RequestParams) ([]byte, error) {
	request, err := http.NewRequest("GET", params.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid url format %s, details: %s", params.URL, err)
	}
	request.Header = r.convertHeader(params.Header...)

	r.logger.Debugf("%+v", request)

	response, err := r.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("r.httpClient.Do: get %s has error %s", params.URL, err.Error())
	}

	r.logger.Infof("httpClient.Get: get %s", params.URL)

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	responseByteArr, err := io.ReadAll(response.Body)

	if err != nil {
		r.logger.Error("io.ReadAll", err.Error())
		return nil, err
	}

	return responseByteArr, nil
}

func (r *APIClient) convertHeader(headers ...APIHeader) http.Header {
	if len(headers) == 0 {
		return nil
	}
	responseHeader := make(http.Header, len(headers))
	for _, header := range headers {
		if len(header.Value) > 0 {
			join := strings.Join(header.Value, ",")
			r.logger.Debugf(join)
			responseHeader.Add(header.Name, join)
		}
	}

	r.logger.Debug(responseHeader)
	return responseHeader
}

// Post create http post request
func (r *APIClient) Post(params RequestParams) ([]byte, error) {
	return r.send("POST", params)
}

// Patch create http post request
func (r *APIClient) Patch(params RequestParams) ([]byte, error) {
	return r.send("PATCH", params)
}

func (r *APIClient) send(method string, params RequestParams) ([]byte, error) {
	if params.Body == nil {
		return nil, fmt.Errorf("request body must be not nil")
	}

	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(params.Body)
	if err != nil {
		return nil, fmt.Errorf("error when convert %#v to bytes.Buffer, details: %s", params.Body, err)
	}

	request, err := http.NewRequest(method, params.URL, &buf)
	if err != nil {
		return nil, fmt.Errorf("invalid format url: %s, body: %s, details: %s", params.URL, &buf, err)
	}
	request.Header = r.convertHeader(params.Header...)

	r.logger.Debugf("%+v", request)

	response, err := r.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("r.httpClient.Do: %s %s has error %s", method, params.URL, err.Error())
	}

	r.logger.Infof("httpClient.%s: url %s, body: %#v", method, params.URL, params.Body)

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	responseByteArr, err := io.ReadAll(response.Body)

	if err != nil {
		r.logger.Error("io.ReadAll", err.Error())
		return nil, err
	}

	return responseByteArr, nil
}
