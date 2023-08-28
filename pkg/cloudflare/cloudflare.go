package cloudflare

import (
	"go.uber.org/zap"

	"goland-ddns/pkg/http"
)

type CloudflareAPI struct {
	logger  *zap.SugaredLogger
	request http.APIClient
}

// NewCloudflareClient create new api connection
func NewCloudflareClient(logger *zap.SugaredLogger, client http.APIClient) (*CloudflareAPI, error) {

	return &CloudflareAPI{
		logger:  logger,
		request: client,
	}, nil
}
