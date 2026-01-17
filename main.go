package main

import (
	"fmt"

	"MusicDev33/econsim/internal/sim"
)

func main() {
	// Economic parameters
	materialsCost := 1.0
	initialWage := 10.0

	// Create market
	market := sim.NewMarket(sim.MarketConfig{
		Product:       "wheat",
		MaterialsCost: materialsCost,
		InitialWage:   initialWage,
	})

	// Setup households
	numHouseholds := 100
	population := 30      // members per household (demand)
	workersPerHH := 20    // working-age members (labor supply)
	savingsBuffer := 10.0 // periods of income as initial savings

	for i := 0; i < numHouseholds; i++ {
		initialCash := float64(workersPerHH) * initialWage * savingsBuffer
		h := sim.NewHousehold(sim.HouseholdConfig{
			Population:  population,
			Workers:     workersPerHH,
			InitialCash: initialCash,
		})
		market.AddHousehold(h)
	}

	// Setup initial firms
	numFirms := 5
	for i := 0; i < numFirms; i++ {
		market.CreateFirm()
	}

	// Run simulation
	steps := 500
	for step := 0; step < steps; step++ {
		market.Step()
		fmt.Println()
	}
}
