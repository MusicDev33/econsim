package sim

import "fmt"

// Firm represents a producer in the economy
type Firm struct {
	ID      string
	Name    string
	Product string

	// Financials
	Cash            float64
	Price           float64
	TargetPrice     float64
	Inventory       int
	OpCosts         float64 // fixed operating costs per tick
	MaterialsCost   float64 // raw materials cost per unit
	PriceFloor      float64 // minimum viable selling price
	ConsecutiveLoss int     // ticks with declining cash (for voluntary exit)
	PrevCash        float64

	// Labor
	Workers           int     // current employees
	TargetWorkers     int     // desired employees
	LaborProductivity float64 // units produced per worker per tick
}

// FirmConfig holds parameters for creating new firms
type FirmConfig struct {
	Product           string
	MaterialsCost     float64
	PriceFloor        float64
	InitialCash       float64
	InitialWorkers    int
	LaborProductivity float64
	OpCosts           float64
}

// DefaultFirmConfig returns sensible defaults
func DefaultFirmConfig(product string, materialsCost, priceFloor float64) FirmConfig {
	return FirmConfig{
		Product:           product,
		MaterialsCost:     materialsCost,
		PriceFloor:        priceFloor,
		InitialCash:       5000.0,
		InitialWorkers:    400,
		LaborProductivity: 1.0,
		OpCosts:           100.0,
	}
}

// NewFirm creates a new firm with the given configuration
func NewFirm(cfg FirmConfig, idNum int) *Firm {
	id := fmt.Sprintf("%s-%d", cfg.Product, idNum)

	initialPrice := cfg.PriceFloor * RandFloat(1.0, 1.2)
	initialInventory := cfg.InitialWorkers * int(cfg.LaborProductivity)

	return &Firm{
		ID:                id,
		Name:              id,
		Product:           cfg.Product,
		Cash:              cfg.InitialCash,
		Price:             initialPrice,
		TargetPrice:       initialPrice,
		Inventory:         initialInventory,
		OpCosts:           cfg.OpCosts,
		MaterialsCost:     cfg.MaterialsCost,
		PriceFloor:        cfg.PriceFloor,
		ConsecutiveLoss:   0,
		PrevCash:          cfg.InitialCash,
		Workers:           cfg.InitialWorkers,
		TargetWorkers:     cfg.InitialWorkers,
		LaborProductivity: cfg.LaborProductivity,
	}
}

// UpdatePrice adjusts the firm's price based on market conditions
func (f *Firm) UpdatePrice(lastSales int, currentWage float64) {
	// Dynamic price floor based on current wages
	f.PriceFloor = f.MaterialsCost + (currentWage / f.LaborProductivity)

	// Calculate price pressure
	pressure := f.calculatePricePressure(lastSales)

	// Update target price
	magnitude := 0.05
	if pressure < -0.5 || pressure > 0.5 {
		magnitude = 0.10
	}

	if pressure > 0 {
		f.TargetPrice *= 1.0 + magnitude
	} else if pressure < 0 {
		f.TargetPrice *= 1.0 - magnitude
	}

	// Enforce floor on target
	if f.TargetPrice < f.PriceFloor {
		f.TargetPrice = f.PriceFloor
	}

	// Price stickiness: move toward target at max 5% per tick
	f.applyPriceStickiness()

	// Hard floor on actual price
	if f.Price < f.PriceFloor {
		f.Price = f.PriceFloor
	}
}

func (f *Firm) calculatePricePressure(lastSales int) float64 {
	pressure := 0.0

	// Inventory pressure: high inventory pushes price down
	invPressure := float64(f.Inventory) / 1000.0
	if invPressure > 0.5 {
		invPressure = 0.5
	}
	pressure -= invPressure

	// Sales pressure: high sales push price up
	switch {
	case lastSales > 300:
		pressure += 0.2
	case lastSales > 200:
		pressure += 0.1
	case lastSales > 100:
		pressure += 0.05
	case lastSales <= 50:
		pressure -= 0.4
	}

	// Cash pressure: rich firms lower prices to gain market share
	switch {
	case f.Cash > 30000:
		pressure -= 0.7
	case f.Cash > 20000:
		pressure -= 0.6
	case f.Cash > 10000:
		pressure -= 0.5
	}

	// Floor protection
	if f.TargetPrice <= f.PriceFloor {
		pressure += 0.6
	}

	return pressure
}

func (f *Firm) applyPriceStickiness() {
	maxAdjustment := 0.05
	diff := f.TargetPrice - f.Price
	maxChange := f.Price * maxAdjustment

	switch {
	case diff > maxChange:
		f.Price += maxChange
	case diff < -maxChange:
		f.Price -= maxChange
	default:
		f.Price = f.TargetPrice
	}
}

// Produce creates inventory based on sales and worker capacity
func (f *Firm) Produce(lastSales int) {
	// Production limited by workers
	maxProduction := f.Workers * int(f.LaborProductivity)

	// Want 1.1x of last sales in inventory
	wantedInv := int(float64(lastSales) * 1.1)
	toProduce := max(wantedInv-f.Inventory, 0)
	produced := min(toProduce, maxProduction)

	f.Inventory += produced
	f.Cash -= f.MaterialsCost * float64(produced)

	// Update labor demand for next tick
	neededProduction := int(float64(lastSales) * 1.1)
	f.TargetWorkers = (neededProduction / int(f.LaborProductivity)) + 1
	f.TargetWorkers = max(f.TargetWorkers, 50) // minimum viable operation
}

// Step performs end-of-tick updates
func (f *Firm) Step(result *MarketResult, currentWage float64) {
	if result == nil {
		return
	}

	sales := result.FirmSales[f.ID]
	f.UpdatePrice(sales, currentWage)
	f.Produce(sales)

	f.Cash -= f.OpCosts

	// Track consecutive losses
	if f.Cash < f.PrevCash {
		f.ConsecutiveLoss++
	} else {
		f.ConsecutiveLoss = 0
	}
	f.PrevCash = f.Cash
}

// IsBankrupt returns true if the firm has no cash
func (f *Firm) IsBankrupt() bool {
	return f.Cash <= 0
}

// WantsToExit returns true if the firm has sustained losses
func (f *Firm) WantsToExit() bool {
	return f.ConsecutiveLoss >= 10
}
