package cloudflare

import (
	"go.uber.org/zap"

	"goland-ddns/pkg/http"
)

type API struct {
	logger  *zap.SugaredLogger
	request http.APIClient
}

// NewCloudflareClient create new api connection
func NewCloudflareClient(logger *zap.SugaredLogger, client http.APIClient) (*API, error) {

	return &API{
		logger:  logger,
		request: client,
	}, nil
}
