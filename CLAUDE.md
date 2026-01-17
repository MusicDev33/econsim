# CLAUDE.md

## Project Overview

EconSim is an economic simulation engine in Go that models market dynamics between households and firms trading wheat. It demonstrates emergent economic behaviors (monopolistic pricing, price competition, market equilibrium, business cycles) through agent-based modeling with behavioral rules and a labor market.

## Tech Stack

- **Go 1.21.1** - Core simulation
- **Plotly.js** - Price visualization in HTML output

## Project Structure

```
econsim/
├── main.go                 # Entry point, simulation setup
├── internal/
│   └── sim/                # Core simulation package
│       ├── types.go        # Shared types (MarketResult, FirmOffer)
│       ├── util.go         # Helper functions (RandFloat)
│       ├── firm.go         # Firm: production, pricing, labor demand
│       ├── household.go    # Household: consumption, labor supply, savings
│       ├── labor.go        # LaborMarket: wage equilibrium, employment
│       └── market.go       # Market: orchestrates labor + goods markets
├── mkt.sh                  # Build/run script with visualization
├── prices.html             # Plotly.js visualization template
└── internal/
    ├── config/             # YAML config (unused)
    └── llm/                # LLM integration (unused)
```

## Running the Simulation

```bash
go run main.go     # Direct execution
./mkt.sh           # Run + generate price visualization
```

## Core Types (internal/sim)

### Firm (`firm.go`)
Producer agent with:
- **Pricing**: Dynamic floor based on wages, price stickiness (5%/tick max change)
- **Production**: Capped by workers × productivity
- **Labor demand**: Based on sales targets, minimum 50 workers

Key methods: `UpdatePrice()`, `Produce()`, `Step()`

### Household (`household.go`)
Consumer agent with:
- **Demand**: Population determines goods demand
- **Labor supply**: Workers available for employment
- **Savings**: 20% of cash saved, 80% available for spending

Key methods: `Step()`, `Spend()`

### LaborMarket (`labor.go`)
Equilibrium wage model:
- Wage adjusts based on supply/demand (max 5%/tick)
- Wage floor: $5
- Workers allocated proportionally to firm demand

Key methods: `Clear()`, `TotalJobs()`

### Market (`market.go`)
Orchestrates the economy:
1. Clear labor market (wages, employment)
2. Clear goods market (prices, sales)
3. Handle firm exit (bankruptcy, voluntary) and entry (probabilistic)
4. Update households (income)

Key methods: `Step()`, `AddFirm()`, `AddHousehold()`

## Simulation Parameters

Configured in `main.go`:
- 100 households, 5 initial firms
- 500 ticks
- Materials cost: $1/unit
- Initial wage: $10/worker
- Savings buffer: 10x income

## Economic Mechanics

### Circular Flow
Firms pay wages → Households earn income → Households buy goods → Firms earn revenue

### Stabilizers
- **Savings**: Households save 20%, buffer against income shocks
- **Minimum workforce**: Firms maintain 50+ workers
- **Dynamic price floor**: Tracks wages to prevent below-cost sales
- **Wage floor**: $5 minimum prevents wage collapse

### Known Dynamics
- Business cycles emerge from feedback loops
- Can get stuck in low-employment equilibrium
- High wages can trigger profit squeeze and recessions
