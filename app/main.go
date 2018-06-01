package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/BurntSushi/toml"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type config struct {
	Slack slackConf
}

type slackConf struct {
	SlackURL string
}

var configVar config

type bitFlyerOpe struct {
	api  PublicApi
	tick chan string
}

func newBitFlyerOpe(ctx context.Context) *bitFlyerOpe {
	b := PublicApi{}
	err, _ := b.GetHealth(ctx)
	if err != nil {
		log.Errorf(ctx, "Get Health of bitflyer: %v", err)
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
	}
}

func getDatastore(ctx context.Context, productCode string) []market {
	var marketArray []market
	q := datastore.NewQuery("Market").Filter("ProductCode =", productCode).Order("-Timestamp").Limit(5)
	if _, err := q.GetAll(ctx, &marketArray); err != nil {
		log.Errorf(ctx, "datastore.Get: %v", err)
	}

	return marketArray
}

func main() {
	http.HandleFunc("/bitcoin", bitcoin)
	appengine.Main()
}

func bitcoin(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	_, err := toml.DecodeFile("config.toml", &configVar)
	if err != nil {
		log.Errorf(ctx, "Setup file can not be read : %v", err)
	}

	b := newBitFlyerOpe(ctx)

	//s := b.getTicker(ctx, "FX_BTC_JPY")
	//log.Debugf(ctx, "a: %v", s)

	markets := []string{
		"BTC_JPY",
		"FX_BTC_JPY",
	}

	results := b.getTickers(ctx, markets)
	marketResults := make(map[string]market)
	marketArray := make(map[string][]market)

	for _, result := range results {
		var m market

		if string(result) != "" {
			err := json.Unmarshal([]byte(result), &m)
			if err != nil {
				log.Errorf(ctx, "JSON Unmarshal error: %v", err)
				return
			}
			m.putDatastore(ctx)
			marketResults[m.ProductCode] = m
			marketArray[m.ProductCode] = getDatastore(ctx, m.ProductCode)
		}
	}

	basicRatio := 3.5

	ratio := ((marketResults["FX_BTC_JPY"].Ltp / marketResults["BTC_JPY"].Ltp) * 100) - 100
	if ratio > basicRatio {
		FxBtcJpy := marketArray["FX_BTC_JPY"][1]
		BtcJpy := marketArray["BTC_JPY"][1]
		ratio2 := ((FxBtcJpy.Ltp / BtcJpy.Ltp) * 100) - 100
		if ratio2 <= basicRatio {
			payload := "{'text':'乖離率が" + strconv.FormatFloat(basicRatio, 'f', 2, 64) + "%を超えました現在の乖離率は" + strconv.FormatFloat(ratio, 'f', 4, 64) + "%です', 'username':'bitcoin-bot', 'channel':'bitcoin', 'icon_emoji':':kityune:'}"
			data := url.Values{}
			data.Set("payload", payload)

			// log.Debugf(ctx, "e: %v", configVar.Slack.SlackURL)
			// log.Debugf(ctx, "e: %v", strconv.FormatFloat(ratio, 'f', 4, 64))

			req, err := http.NewRequest("POST", configVar.Slack.SlackURL, bytes.NewBufferString(data.Encode()))
			if err != nil {
				log.Errorf(ctx, "Fail post to slack: %v", err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			client := urlfetch.Client(ctx)
			_, err = client.Do(req)

			if err != nil {
				log.Errorf(ctx, "Fail post to slack: %v", err)
			}
		}
	}
}
