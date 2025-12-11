package main

import (
	"cmp"
	"fmt"
	"math/rand"
	"slices"
)

type SimpleFirm struct {
	ID        string
	Name      string
	Product   string
	Cash      float64
	Price     float64
	Inventory int
	OpCosts   float64
	BasePrice float64
}

func NewSimpleFirm(product string, floorPrice float64, idNum int) SimpleFirm {
	startPrice := floorPrice * 1.5
	variance := startPrice * 0.2

	id := fmt.Sprintf("%s-%d", product, idNum)
	opCosts := 50.0

	f := SimpleFirm{
		ID:        id,
		Name:      id,
		Product:   product,
		Cash:      1000.0,
		Price:     randFloat(startPrice-variance, startPrice+variance),
		Inventory: 350,
		OpCosts:   opCosts,
		BasePrice: floorPrice,
	}

	return f
}

func (f *SimpleFirm) CreatePrice(lastSales int) {
	adjRate := 0.05

	pricePressure := 0.0
	invPricePressure := 0.0
	invPricePressure = float64(f.Inventory) / 1000.0

	if invPricePressure > 0.5 {
		invPricePressure = 0.5
	}

	pricePressure -= invPricePressure

	if lastSales > 300 {
		pricePressure += 0.2
	} else if lastSales > 200 {
		pricePressure += 0.1
	} else if lastSales > 100 {
		pricePressure += 0.05
	}

	if lastSales <= 100 {
		pricePressure -= 0.2
	}

	if f.Cash > 10000 {
		pricePressure -= 0.5
	} else if f.Cash > 20000 {
		pricePressure -= 0.7
	} else if f.Cash > 30000 {
		pricePressure -= 0.9
	}

	adjRateMod := 1.0
	if pricePressure < -0.5 || pricePressure > 0.5 {
		adjRateMod = 2.0
	}

	if f.Price <= f.BasePrice {
		pricePressure += 0.4
	}

	if pricePressure > 0 {
		f.Price *= 1.0 + adjRate*adjRateMod
		return
	}

	if pricePressure < 0 {
		f.Price *= 1.0 - adjRate*adjRateMod
		return
	}
}

func (f *SimpleFirm) Produce(lastSales int) {
	maxNewCap := int(float64(lastSales) * 0.2)
	maxNewCap = min(maxNewCap, 200)

	wantedInv := int(float64(lastSales) * 1.1)
	toProduce := wantedInv - f.Inventory
	if toProduce <= 0 {
		return
	}

	if lastSales == 0 {
		return
	}

	produced := min(toProduce, lastSales+maxNewCap)
	produced = max(produced, 0)
	f.Inventory += produced
	f.Cash -= f.BasePrice * float64(produced)
}

func (f *SimpleFirm) Step(res *MarketResult) {
	if res == nil {
		return
	}

	f.CreatePrice(res.FirmSales[f.ID])
	f.Produce(res.FirmSales[f.ID])

	f.Cash -= f.OpCosts
}

type SimpleHousehold struct {
	Population int

	IncomeWages       float64
	ConsumptionBudget float64
}

func (h *SimpleHousehold) Step() {
	h.Population += int(0.05 * float64(h.Population))
	h.IncomeWages += 0.05 * h.IncomeWages
	h.ConsumptionBudget = h.IncomeWages * 0.8
}

// Market for one good
type BasicMarket struct {
	Product            string
	Firms              []SimpleFirm
	FirmMap            map[string]int
	Households         []SimpleHousehold
	PrevResult         *MarketResult
	FloorPrice         float64
	TotalHistoricFirms int
}

func (b *BasicMarket) RegisterFirm(f SimpleFirm) {
	b.Firms = append(b.Firms, f)

	firmMap := map[string]int{}
	for i, f := range b.Firms {
		firmMap[f.ID] = i
	}
	b.FirmMap = firmMap

	b.TotalHistoricFirms += 1
}

func (b *BasicMarket) RemoveFirm(firmID string) {
	b.Firms = slices.DeleteFunc(b.Firms, func(f SimpleFirm) bool {
		return f.ID == firmID
	})

	firmMap := map[string]int{}
	for i, f := range b.Firms {
		firmMap[f.ID] = i
	}
	b.FirmMap = firmMap
}

func (b *BasicMarket) RegisterHousehold(h SimpleHousehold) {
	b.Households = append(b.Households, h)
}

func (b *BasicMarket) PrintLastMkt() {
	res := b.PrevResult
	if res == nil {
		fmt.Println("No prev result")
		return
	}

	fmt.Printf("Market Price for %s: %.2f\n", b.Product, res.LastPrice)
	fmt.Printf("  - Demand: %d\n", res.Demand)
	fmt.Printf("  - Supply: %d\n", res.Supply)
	fmt.Printf("  - Total Sales: %d\n", res.TotalSales)
	fmt.Println("Firms:")
	for k, v := range res.FirmSales {
		fmt.Printf("  - %s (%.2f)\n", k, b.Firms[b.FirmMap[k]].Cash)
		fmt.Printf("    - Sales: %d\n", v)
		fmt.Printf("    - Price: %.2f\n", b.Firms[b.FirmMap[k]].Price)
	}
}

type FirmOffer struct {
	FirmID       string
	PricePerUnit float64
	Qty          int
}

func (b *BasicMarket) Step() {
	offers := []FirmOffer{}
	totalSupply := 0
	for _, f := range b.Firms {
		offers = append(offers, FirmOffer{
			FirmID:       f.ID,
			PricePerUnit: f.Price,
			Qty:          f.Inventory,
		})

		totalSupply += f.Inventory

		slices.SortFunc(offers, func(a, b FirmOffer) int {
			return cmp.Compare(a.PricePerUnit, b.PricePerUnit)
		})
	}

	totalDemand := 0
	remainingDemand := 0
	hhDemand := 0
	for _, h := range b.Households {
		totalDemand += h.Population
		hhDemand = 0
		budget := h.ConsumptionBudget

		for _, o := range offers {
			desired := int(budget / o.PricePerUnit)
			// For now, consume one wheat per pop
			if desired > h.Population {
				hhDemand = h.Population
				budget -= float64(h.Population) * o.PricePerUnit
				break
			}

			hhDemand += desired
			if hhDemand > h.Population {
				purchased := hhDemand - h.Population
				budget -= float64(purchased) * o.PricePerUnit
				hhDemand = h.Population
				break
			}

			budget -= float64(desired) * o.PricePerUnit
		}

		if budget < 0 {
			panic(fmt.Errorf("simulation err: budget went below 0"))
		}

		remainingDemand += hhDemand
	}

	sales := map[string]int{}
	for _, o := range offers {
		if remainingDemand <= 0 {
			sales[o.FirmID] = 0
			continue
		}

		sold := min(o.Qty, remainingDemand)
		sales[o.FirmID] = sold
		remainingDemand -= sold
	}

	dollarsTransacted := 0.0
	totalSold := 0
	for _, o := range offers {
		sold := sales[o.FirmID]
		fIndex := b.FirmMap[o.FirmID]
		b.Firms[fIndex].Cash += float64(sold) * o.PricePerUnit
		b.Firms[fIndex].Inventory -= sold

		totalSold += sold
		dollarsTransacted += float64(sold) * o.PricePerUnit
	}

	marketPrice := dollarsTransacted / float64(totalSold)
	mktRes := MarketResult{
		LastPrice:  marketPrice,
		Supply:     totalSupply,
		Demand:     totalDemand,
		TotalSales: totalSold,
		FirmSales:  sales,
	}

	b.PrevResult = &mktRes
	b.PrintLastMkt()

	toRmv := []string{}
	for i := range b.Firms {
		b.Firms[i].Step(&mktRes)

		if b.Firms[i].Cash <= 0 {
			toRmv = append(toRmv, b.Firms[i].ID)
		}
	}

	for _, v := range toRmv {
		b.RemoveFirm(v)
	}

	if float64(totalSold)/float64(totalDemand) < 0.66 {
		newFirm := NewSimpleFirm(b.Product, b.FloorPrice, b.TotalHistoricFirms+1)
		b.RegisterFirm(newFirm)
	}

	// for i := range b.Households {
	// 	b.Households[i].Step()
	// }
}

func (b *BasicMarket) PrintInfo() {
	fmt.Println("Firms:")
	for _, f := range b.Firms {
		fmt.Printf("  - %s: %f\n", f.Name, f.Price)
	}

	fmt.Println("Households:")
	for _, h := range b.Households {
		fmt.Printf("  - %f\n", h.IncomeWages)
	}
}

type MarketInfo struct {
	LastPrice float64
	Supply    int
	Demand    int
}

type MarketResult struct {
	LastPrice  float64
	Supply     int
	Demand     int
	TotalSales int
	FirmSales  map[string]int
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func main() {
	floorPrice := 11.0
	product := "wheat"

	bm := BasicMarket{
		Product:    product,
		Firms:      []SimpleFirm{},
		FirmMap:    map[string]int{},
		Households: []SimpleHousehold{},
		PrevResult: nil,
		FloorPrice: floorPrice,
	}

	// Setup households
	hhs := 50
	hhPop := 30
	i := 0
	for i < hhs {
		wages := randFloat(200.0, 400.0)
		h := SimpleHousehold{
			Population:        hhPop,
			IncomeWages:       wages,
			ConsumptionBudget: wages * 0.8,
		}

		bm.RegisterHousehold(h)
		i++
	}

	// Setup firms
	firms := 5
	i = 0
	for i < firms {
		f := NewSimpleFirm(bm.Product, bm.FloorPrice, i+1)

		bm.RegisterFirm(f)
		i++
	}

	steps := 100
	curStep := 0
	for curStep < steps {
		bm.Step()
		fmt.Println()
		curStep += 1
	}
}
