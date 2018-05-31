package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type bitFlyerOpe struct {
	api  PublicApi
	tick chan string
}

func newBitFlyerOpe(ctx context.Context) *bitFlyerOpe {
	b := PublicApi{}
	err, _ := b.GetHealth(ctx)
	if err != nil {
		log.Errorf(ctx, "Get Health of bitflyer: %v", err)
		os.Exit(1)
	}

	return &bitFlyerOpe{
		api:  b,
		tick: make(chan string),
	}
}

func (b *bitFlyerOpe) getTicker(ctx context.Context, code string) string {
	m := make(map[string]string)
	m["product_code"] = code

	err, s := b.api.GetTicker(m, ctx)
	if err != nil {
		log.Errorf(ctx, "Get ticker fail: %v", err)
		return ""
	}

	return s
}

func (b *bitFlyerOpe) getTickers(ctx context.Context, markets []string) []string {
	results := []string{}

	for _, market := range markets {
		go func(market string) {
			b.tick <- b.getTicker(ctx, market)
		}(market)
	}

	for i := 0; i <= len(markets)-1; i++ {
		results = append(results, <-b.tick)
	}

	return results
}

type market struct {
	ProductCode string `json:"product_code"`
	Timestamp   string `json:"timestamp"`
	// TickID          int32   `json:"tick_id"`
	// BestBid         float64 `json:"best_bid"`
	// BestAsk         float64 `json:"best_ask"`
	// BestBidSize     float64 `json:"best_bid_size"`
	// BestAskSize     float64 `json:"best_ask_size"`
	// TotalBidDepth   float64 `json:"total_bid_depth"`
	// TotalAskDepth   float64 `json:"total_ask_depth"`
	Ltp float64 `json:"ltp"`
	//Volume          float64 `json:"volume"`
	//VolumeByProduct float64 `json:"volume_by_product"`
}

func (m *market) putDatastore(ctx context.Context) {
	kind := "Market"
	key := datastore.NewIncompleteKey(ctx, kind, nil)
	if _, err := datastore.Put(ctx, key, m); err != nil {
		log.Errorf(ctx, "datastore.Put: %v", err)

		return
	}
}

func main() {
	http.HandleFunc("/bitcoin", bitcoin)
	appengine.Main()
}

func bitcoin(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	b := newBitFlyerOpe(ctx)

	//s := b.getTicker(ctx, "FX_BTC_JPY")
	//log.Debugf(ctx, "a: %v", s)

	markets := []string{
		"BTC_JPY",
		"FX_BTC_JPY",
	}

	results := b.getTickers(ctx, markets)

	for _, result := range results {
		var m market

		if string(result) != "" {
			err := json.Unmarshal([]byte(result), &m)
			if err != nil {
				log.Errorf(ctx, "JSON Unmarshal error: %v", err)
				return
			}
			m.putDatastore(ctx)
		}
	}
}
