package cloudflare

import (
	"fmt"
	"goland-ddns/pkg/config"
	"goland-ddns/pkg/http"
)

var cloudflareBaseEndpoint = "https://api.cloudflare.com/client/v4"

type BaseCloudflareResponse[T any] struct {
	Messages   []ErrorResponseData `json:"messages"`
	Success    bool                `json:"success"`
	ResultInfo *ResultInfo         `json:"result_info,omitempty"`
	Result     T                   `json:"result"`
	Errors     []interface{}       `json:"errors"`
}

type ResultInfo struct {
	Count      int `json:"count"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalCount int `json:"total_count"`
}

// ErrorResponseData provide error in response api
type ErrorResponseData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *API) commonHeader() []http.APIHeader {
	return []http.APIHeader{
		{
			Name:  "Content-Type",
			Value: []string{"application/json"},
		},
		{
			Name: "Authorization",
			Value: []string{
				fmt.Sprintf("Bearer %s", config.ENV.Cloudflare.API.ID),
			},
		},
	}
}
