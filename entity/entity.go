package entity

type KlineStreamObject struct {
	KlineStartTime   int64  `json:"t"`
	KlineCloseTime   int64  `json:"T"`
	Symbol           string `json:"s"`
	Interval         string `json:"i"`
	FirstTradeID     int64  `json:"f"`
	LastTradeID      int64  `json:"L"`
	OpenPrice        string `json:"o"`
	ClosePrice       string `json:"c"`
	HighPrice        string `json:"h"`
	LowPrice         string `json:"l"`
	BaseAssetVolume  string `json:"v"`
	NumberOfTrade    int64  `json:"n"`
	KlineClosed      bool   `json:"x"`
	QuoteAssetVolume string `json:"q"`
	TakerBase        string `json:"V"`
	TakerQuote       string `json:"Q"`
	Ignore           string `json:"B"`
}

type KlineItem struct {
	EventType string            `json:"e"`
	EventTime int64             `json:"E"`
	Symbol    string            `json:"s"`
	Kline     KlineStreamObject `json:"k"`
}
