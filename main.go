package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pokochi/bitFlyer"
)

type bitFlyerOpe struct {
	api  bitFlyer.PublicApi
	tick chan string
}

func newBitFlyerOpe() *bitFlyerOpe {
	b := bitFlyer.PublicApi{}
	err, s := b.GetHealth()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(s)

	return &bitFlyerOpe{
		api:  b,
		tick: make(chan string),
	}
}

func (b *bitFlyerOpe) getTicker(code string) string {
	m := make(map[string]string)
	m["product_code"] = code

	err, data := b.api.GetTicker(m)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return data
}

func (b *bitFlyerOpe) getTickers(markets []string) []string {
	results := []string{}

	for _, market := range markets {
		go func(market string) {
			b.tick <- b.getTicker(market)
		}(market)
	}

	for i := 0; i <= len(markets)-1; i++ {
		results = append(results, <-b.tick)
	}

	return results
}

type market struct {
	ProductCode     string  `json:"product_code"`
	Timestamp       string  `json:"timestamp"`
	TickId          int32   `json:"tick_id"`
	BestBid         float64 `json:"best_bid"`
	BestAsk         float64 `json:"best_ask"`
	BestBidSize     float64 `json:"best_bid_size"`
	BestAskSize     float64 `json:"best_ask_size"`
	TotalBidDepth   float64 `json:"total_bid_depth"`
	TotalAskDepth   float64 `json:"total_ask_depth"`
	Ltp             float64 `json:"ltp"`
	Volume          float64 `json:"volume"`
	VolumeByProduct float64 `json:"volume_by_product"`
}

func main() {
	b := newBitFlyerOpe()

	markets := []string{
		"BTC_JPY",
		"FX_BTC_JPY",
	}

	results := b.getTickers(markets)

	for _, result := range results {
		var m market
		err := json.Unmarshal([]byte(result), &m)
		if err != nil {
			fmt.Println("JSON Unmarshal error:", err)
			return
		}
		fmt.Printf("%+v\n", m)
	}
}
