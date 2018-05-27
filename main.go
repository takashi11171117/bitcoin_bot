package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/datastore"
	"github.com/pokochi/bitFlyer"
	"golang.org/x/net/context"
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
	projectID := "bitcoin-takashi1117"

	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	kind := "Market"
	name := "sampletask1"
	key := datastore.NameKey(kind, name, nil)
	if _, err := client.Put(ctx, key, m); err != nil {
		log.Fatalf("Failed to save task: %v", err)
	}

	fmt.Printf("Saved %v: %v\n", key, m)
}

func main() {
	b := newBitFlyerOpe()

	markets := []string{
		"BTC_JPY",
		"FX_BTC_JPY",
	}

	results := b.getTickers(markets)

	ctx := context.Background()

	for _, result := range results {
		var m market
		err := json.Unmarshal([]byte(result), &m)
		if err != nil {
			fmt.Println("JSON Unmarshal error:", err)
			return
		}
		m.putDatastore(ctx)
	}
}
