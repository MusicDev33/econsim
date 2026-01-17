package sim

const savingsRate = 0.2 // households save 20% of cash

// Household represents a consumer unit in the economy
type Household struct {
	Population int // total members (determines demand for goods)
	Workers    int // working-age members (labor supply)

	Employed          int     // currently employed workers
	Cash              float64 // total savings
	ConsumptionBudget float64 // amount willing to spend this tick
}

// HouseholdConfig holds parameters for creating households
type HouseholdConfig struct {
	Population  int
	Workers     int
	InitialCash float64
}

// NewHousehold creates a new household
func NewHousehold(cfg HouseholdConfig) *Household {
	return &Household{
		Population:        cfg.Population,
		Workers:           cfg.Workers,
		Employed:          cfg.Workers, // start fully employed
		Cash:              cfg.InitialCash,
		ConsumptionBudget: cfg.InitialCash * (1 - savingsRate),
	}
}

// Step performs end-of-tick updates: receive income, update budget
func (h *Household) Step(wageRate float64) {
	income := float64(h.Employed) * wageRate
	h.Cash += income
	h.ConsumptionBudget = h.Cash * (1 - savingsRate)
}

// Spend deducts the given amount from cash
func (h *Household) Spend(amount float64) {
	h.Cash -= amount
}
