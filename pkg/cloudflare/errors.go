package cloudflare

import "github.com/pkg/errors"

var (
	ErrNoRouteMatches = errors.New("NO_ROUTE_MATCHES") // ErrNoRoute
)
