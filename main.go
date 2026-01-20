package main

import (
	"log"

	"MusicDev33/econsim/internal/config"
	"MusicDev33/econsim/internal/sim"
)

func main() {
	cfg := config.Get()

	// Setup event emitter
	events, err := sim.NewEventEmitter(cfg.EventFile)
	if err != nil {
		log.Fatalf("Failed to create event emitter: %v", err)
	}
	defer events.Close()

	// Economic parameters
	materialsCost := 1.0
	initialWage := 10.0

	// Create market
	market := sim.NewMarket(sim.MarketConfig{
		Product:       "wheat",
		MaterialsCost: materialsCost,
		InitialWage:   initialWage,
	})
	market.Events = events

	// Setup households
	workersPerHH := (cfg.PopPerHousehold * 2) / 3 // ~2/3 of population are workers
	savingsBuffer := 10.0                         // periods of income as initial savings

	for i := 0; i < cfg.NumHouseholds; i++ {
		initialCash := float64(workersPerHH) * initialWage * savingsBuffer
		h := sim.NewHousehold(sim.HouseholdConfig{
			Population:  cfg.PopPerHousehold,
			Workers:     workersPerHH,
			InitialCash: initialCash,
		})
		market.AddHousehold(h)
	}

	// Setup initial firms
	for i := 0; i < cfg.NumStartingFirms; i++ {
		market.CreateFirm()
	}

	// Run simulation
	for step := 0; step < cfg.NumTicks; step++ {
		market.Step()
	}
}
