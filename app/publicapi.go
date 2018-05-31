package main

import "context"

const (
	HEALTH_PATH     = "/v1/gethealth"
	BOARD_PATH      = "/v1/getboard"
	TICKER_PATH     = "/v1/getticker"
	EXECUTIONS_PATH = "/v1/getexecutions"
)

type PublicApi struct {
	Lightning
}

//Ticker
func (p *PublicApi) GetTicker(m map[string]string, ctx context.Context) (error, string) {
	return p.RequestPublic(GET_METHOD, TICKER_PATH, m, ctx)
}

//取引所の状態
func (p *PublicApi) GetHealth(ctx context.Context) (error, string) {
	return p.RequestPublic(GET_METHOD, HEALTH_PATH, nil, ctx)
}
