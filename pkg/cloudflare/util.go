package cloudflare

import (
	"fmt"

	"goland-ddns/pkg/config"
	"goland-ddns/pkg/http"
)

var cloudflareBaseEndpoint = "https://api.cloudflare.com/client/v4"

func (c *CloudflareAPI) commonHeader() []http.APIHeader {
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
