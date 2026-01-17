package sim

// MarketResult captures the outcome of a single market tick
type MarketResult struct {
	LastPrice       float64
	Supply          int
	Demand          int
	TotalSales      int
	FirmSales       map[string]int
	AvgProfitMargin float64 // (avgPrice - floorPrice) / floorPrice
}

// FirmOffer represents a firm's offer to sell goods
type FirmOffer struct {
	FirmID       string
	PricePerUnit float64
	Qty          int
}
