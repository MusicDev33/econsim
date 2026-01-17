package sim

const (
	maxWageChange = 0.05 // wages adjust max 5% per tick
	wageFloor     = 5.0  // minimum wage
)

// LaborMarket matches workers with firms
type LaborMarket struct {
	WageRate         float64        // equilibrium wage
	TotalLaborSupply int            // sum of household workers
	TotalLaborDemand int            // sum of firm labor needs
	Employment       map[string]int // firmID -> workers allocated
}

// NewLaborMarket creates a labor market with the given initial wage
func NewLaborMarket(initialWage float64) *LaborMarket {
	return &LaborMarket{
		WageRate:   initialWage,
		Employment: make(map[string]int),
	}
}

// Clear matches labor supply with demand, updates wages, allocates workers
func (lm *LaborMarket) Clear(firms []*Firm, households []*Household) {
	lm.calculateSupplyAndDemand(firms, households)
	lm.adjustWage()
	lm.allocateWorkers(firms)
}

func (lm *LaborMarket) calculateSupplyAndDemand(firms []*Firm, households []*Household) {
	lm.TotalLaborSupply = 0
	for _, h := range households {
		lm.TotalLaborSupply += h.Workers
	}

	lm.TotalLaborDemand = 0
	for _, f := range firms {
		lm.TotalLaborDemand += f.TargetWorkers
	}
}

func (lm *LaborMarket) adjustWage() {
	if lm.TotalLaborDemand > lm.TotalLaborSupply {
		// Labor shortage: wages rise
		shortage := float64(lm.TotalLaborDemand-lm.TotalLaborSupply) / float64(lm.TotalLaborSupply+1)
		increase := min(shortage*0.1, maxWageChange)
		lm.WageRate *= 1.0 + increase
	} else if lm.TotalLaborDemand < lm.TotalLaborSupply {
		// Labor surplus: wages fall
		surplus := float64(lm.TotalLaborSupply-lm.TotalLaborDemand) / float64(lm.TotalLaborSupply+1)
		decrease := min(surplus*0.1, maxWageChange)
		lm.WageRate *= 1.0 - decrease
	}

	if lm.WageRate < wageFloor {
		lm.WageRate = wageFloor
	}
}

func (lm *LaborMarket) allocateWorkers(firms []*Firm) {
	lm.Employment = make(map[string]int)

	if lm.TotalLaborDemand <= 0 {
		return
	}

	for _, f := range firms {
		// Proportional allocation
		share := float64(f.TargetWorkers) / float64(lm.TotalLaborDemand)
		allocated := int(float64(lm.TotalLaborSupply) * share)
		allocated = min(allocated, f.TargetWorkers) // can't get more than requested
		lm.Employment[f.ID] = allocated
	}
}

// TotalJobs returns the actual number of filled positions
func (lm *LaborMarket) TotalJobs() int {
	return min(lm.TotalLaborDemand, lm.TotalLaborSupply)
}
