package sim

import (
	"cmp"
	"fmt"
	"math/rand"
	"slices"
)

// Market orchestrates the economy for a single product
type Market struct {
	Product       string
	Firms         []*Firm
	firmIndex     map[string]int // ID -> index for fast lookup
	Households    []*Household
	PrevResult    *MarketResult
	MaterialsCost float64
	PriceFloor    float64
	Labor         *LaborMarket

	totalHistoricFirms int
	firmConfig         FirmConfig
}

// MarketConfig holds parameters for creating a market
type MarketConfig struct {
	Product       string
	MaterialsCost float64
	InitialWage   float64
}

// NewMarket creates a new market
func NewMarket(cfg MarketConfig) *Market {
	laborProductivity := 1.0
	priceFloor := cfg.MaterialsCost + (cfg.InitialWage / laborProductivity)

	return &Market{
		Product:       cfg.Product,
		Firms:         []*Firm{},
		firmIndex:     map[string]int{},
		Households:    []*Household{},
		MaterialsCost: cfg.MaterialsCost,
		PriceFloor:    priceFloor,
		Labor:         NewLaborMarket(cfg.InitialWage),
		firmConfig:    DefaultFirmConfig(cfg.Product, cfg.MaterialsCost, priceFloor),
	}
}

// AddFirm registers a new firm in the market
func (m *Market) AddFirm(f *Firm) {
	m.Firms = append(m.Firms, f)
	m.rebuildFirmIndex()
	m.totalHistoricFirms++
}

// RemoveFirm removes a firm from the market
func (m *Market) RemoveFirm(firmID string) {
	m.Firms = slices.DeleteFunc(m.Firms, func(f *Firm) bool {
		return f.ID == firmID
	})
	m.rebuildFirmIndex()
}

func (m *Market) rebuildFirmIndex() {
	m.firmIndex = make(map[string]int, len(m.Firms))
	for i, f := range m.Firms {
		m.firmIndex[f.ID] = i
	}
}

// AddHousehold registers a new household in the market
func (m *Market) AddHousehold(h *Household) {
	m.Households = append(m.Households, h)
}

// CreateFirm creates and adds a new firm to the market
func (m *Market) CreateFirm() *Firm {
	f := NewFirm(m.firmConfig, m.totalHistoricFirms+1)
	m.AddFirm(f)
	return f
}

// Step advances the simulation by one tick
func (m *Market) Step() {
	m.clearLaborMarket()
	m.clearGoodsMarket()
	m.handleFirmExitAndEntry()
	m.updateHouseholds()
}

func (m *Market) clearLaborMarket() {
	m.Labor.Clear(m.Firms, m.Households)

	// Assign workers to firms and pay wages
	for _, f := range m.Firms {
		f.Workers = m.Labor.Employment[f.ID]
		wageCost := float64(f.Workers) * m.Labor.WageRate
		f.Cash -= wageCost
	}

	// Distribute employment to households
	m.distributeEmployment()

	fmt.Printf("Labor: wage=%.2f, demand=%d, supply=%d, employed=%d\n",
		m.Labor.WageRate, m.Labor.TotalLaborDemand, m.Labor.TotalLaborSupply, m.Labor.TotalJobs())
}

func (m *Market) distributeEmployment() {
	totalJobs := m.Labor.TotalJobs()
	remainingJobs := totalJobs

	for _, h := range m.Households {
		if m.Labor.TotalLaborSupply <= 0 || remainingJobs <= 0 {
			h.Employed = 0
			continue
		}

		// Proportional allocation with rounding
		share := float64(h.Workers) / float64(m.Labor.TotalLaborSupply)
		employed := int(float64(totalJobs)*share + 0.5)
		employed = min(employed, h.Workers)
		employed = min(employed, remainingJobs)

		h.Employed = employed
		remainingJobs -= employed
	}
}

func (m *Market) clearGoodsMarket() {
	offers := m.collectOffers()
	sales, totalSupply := m.matchBuyersAndSellers(offers)
	m.settleTransactions(offers, sales)
	m.recordMarketResult(offers, sales, totalSupply)
	m.updateFirms()
}

func (m *Market) collectOffers() []FirmOffer {
	offers := make([]FirmOffer, 0, len(m.Firms))
	for _, f := range m.Firms {
		offers = append(offers, FirmOffer{
			FirmID:       f.ID,
			PricePerUnit: f.Price,
			Qty:          f.Inventory,
		})
	}

	// Sort by price (cheapest first)
	slices.SortFunc(offers, func(a, b FirmOffer) int {
		return cmp.Compare(a.PricePerUnit, b.PricePerUnit)
	})

	return offers
}

func (m *Market) matchBuyersAndSellers(offers []FirmOffer) (map[string]int, int) {
	remainingInv := make(map[string]int)
	totalSupply := 0
	for _, o := range offers {
		remainingInv[o.FirmID] = o.Qty
		totalSupply += o.Qty
	}

	sales := make(map[string]int)
	for _, o := range offers {
		sales[o.FirmID] = 0
	}

	// Households buy from cheapest sellers first
	for _, h := range m.Households {
		budget := h.ConsumptionBudget
		toBuy := h.Population

		for _, o := range offers {
			if toBuy <= 0 || budget <= 0 {
				break
			}

			avail := remainingInv[o.FirmID]
			if avail <= 0 {
				continue
			}

			canAfford := int(budget / o.PricePerUnit)
			want := min(toBuy, canAfford, avail)
			if want <= 0 {
				continue
			}

			cost := float64(want) * o.PricePerUnit
			budget -= cost
			toBuy -= want
			remainingInv[o.FirmID] -= want
			sales[o.FirmID] += want
		}

		// Deduct spending from cash (savings remain)
		spent := h.ConsumptionBudget - budget
		h.Spend(spent)
	}

	return sales, totalSupply
}

func (m *Market) settleTransactions(offers []FirmOffer, sales map[string]int) {
	for _, o := range offers {
		sold := sales[o.FirmID]
		idx := m.firmIndex[o.FirmID]

		m.Firms[idx].Cash += float64(sold) * o.PricePerUnit
		m.Firms[idx].Inventory -= sold
	}
}

func (m *Market) recordMarketResult(offers []FirmOffer, sales map[string]int, totalSupply int) {
	var totalSold int
	var dollarsTransacted float64

	for _, o := range offers {
		sold := sales[o.FirmID]
		totalSold += sold
		dollarsTransacted += float64(sold) * o.PricePerUnit
	}

	totalDemand := 0
	for _, h := range m.Households {
		totalDemand += h.Population
	}

	// Calculate market price
	var marketPrice float64
	if totalSold > 0 {
		marketPrice = dollarsTransacted / float64(totalSold)
	} else if m.PrevResult != nil {
		marketPrice = m.PrevResult.LastPrice
	} else {
		marketPrice = m.PriceFloor
	}

	profitMargin := (marketPrice - m.PriceFloor) / m.PriceFloor

	m.PrevResult = &MarketResult{
		LastPrice:       marketPrice,
		Supply:          totalSupply,
		Demand:          totalDemand,
		TotalSales:      totalSold,
		FirmSales:       sales,
		AvgProfitMargin: profitMargin,
	}

	m.printMarketStatus()
}

func (m *Market) printMarketStatus() {
	res := m.PrevResult
	fmt.Printf("Market Price for %s: %.2f\n", m.Product, res.LastPrice)
	fmt.Printf("  - Demand: %d\n", res.Demand)
	fmt.Printf("  - Supply: %d\n", res.Supply)
	fmt.Printf("  - Total Sales: %d\n", res.TotalSales)
	fmt.Println("Firms:")
	for id, sales := range res.FirmSales {
		f := m.Firms[m.firmIndex[id]]
		fmt.Printf("  - %s ($%.2f, %d units)\n", id, f.Cash, f.Inventory)
		fmt.Printf("    - Sales: %d\n", sales)
		fmt.Printf("    - Price: %.2f\n", f.Price)
	}
}

func (m *Market) updateFirms() {
	for _, f := range m.Firms {
		f.Step(m.PrevResult, m.Labor.WageRate)
	}
}

func (m *Market) handleFirmExitAndEntry() {
	m.handleExits()
	m.handleEntry()
}

func (m *Market) handleExits() {
	var toRemove []string

	for _, f := range m.Firms {
		if f.IsBankrupt() {
			toRemove = append(toRemove, f.ID)
			fmt.Printf("  [EXIT] %s: bankruptcy\n", f.ID)
		} else if f.WantsToExit() {
			toRemove = append(toRemove, f.ID)
			fmt.Printf("  [EXIT] %s: voluntary (sustained losses)\n", f.ID)
		}
	}

	for _, id := range toRemove {
		m.RemoveFirm(id)
	}
}

func (m *Market) handleEntry() {
	if m.PrevResult == nil {
		return
	}

	// Calculate entry probability
	entryProb := 0.0

	margin := m.PrevResult.AvgProfitMargin
	if margin > 0.3 {
		entryProb += 0.3
	} else if margin > 0.15 {
		entryProb += 0.15
	}

	satisfaction := float64(m.PrevResult.TotalSales) / float64(m.PrevResult.Demand)
	if satisfaction < 0.8 {
		entryProb += 0.2
	}

	if rand.Float64() < entryProb {
		newFirm := m.CreateFirm()
		fmt.Printf("  [ENTRY] %s: new firm entered (margin=%.2f, satisfaction=%.2f)\n",
			newFirm.ID, margin, satisfaction)
	}
}

func (m *Market) updateHouseholds() {
	for _, h := range m.Households {
		h.Step(m.Labor.WageRate)
	}
}
