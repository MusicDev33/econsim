# CLAUDE.md

## Project Overview

EconSim is an economic simulation engine in Go that models market dynamics between households and firms trading wheat. It demonstrates emergent economic behaviors (monopolistic pricing, price competition, market equilibrium) through agent-based modeling with simple behavioral rules.

## Tech Stack

- **Go 1.21.1** - Core simulation
- **Plotly.js** - Price visualization in HTML output
- **YAML** - Config support via `go-yaml` (infrastructure ready, not actively used)

## Key Files

| File | Purpose |
|------|---------|
| `main.go` | All simulation logic: firms, households, market clearing |
| `mkt.sh` | Build/run script: runs sim, extracts prices, generates HTML viz |
| `prices.html` | Visualization template with Plotly.js charts |
| `internal/config/config.go` | YAML config loader (expects `akGrok` key) |
| `internal/llm/llm.go` | Grok LLM API client (infrastructure, not integrated) |

## Running the Simulation

```bash
./mkt.sh           # Full pipeline: run sim + generate visualization
go run main.go     # Direct execution (outputs to stdout)
```

Output goes to `simrun/` directory. Open `simrun/prices.html` for interactive price chart.

## Core Data Structures

**SimpleFirm** (lines 10-42): Agent with Cash, Price, Inventory, OpCosts, BasePrice
- `CreatePrice()` - Adaptive pricing based on inventory, sales history, cash reserves
- `Produce()` - Produces 1.1x of last sales (capped at 20% capacity growth)

**SimpleHousehold** (lines 128-139): Consumer with Population, Income, ConsumptionBudget

**BasicMarket** (lines 142-326): Orchestrates market clearing
- Sorts firm offers by price (ascending)
- Households buy from cheapest first (budget-constrained)
- Removes bankrupt firms (cash <= 0)
- Spawns new firm if sales < 66% of demand

**MarketResult** (lines 333-339): Records LastPrice, Supply, Demand, TotalSales

## Simulation Parameters (in main.go)

- 100 households, 5 initial firms
- 500 simulation ticks
- Household income: 10-20 (random), 80% budget for consumption
- Firm operating cost: 10/tick
- Base price floor: 1.0

## Economic Mechanics

1. **Pricing**: Inventory pressure (high inv → lower price), sales history (high sales → raise price), cash reserves (rich firms lower price for market share), floor price protection
2. **Production**: Target 1.1x last sales, max 20% growth per tick
3. **Market Entry**: New firm enters when market is undersupplied (sales < 66% demand)
4. **Exit**: Firms with cash <= 0 are removed

## Code Patterns

- Agent-based modeling with simple behavioral rules
- Factory pattern: `NewSimpleFirm()`
- Pointer receivers for state modification
- Maps for O(1) firm lookup by ID
- Monolithic main.go (educational simplicity)
