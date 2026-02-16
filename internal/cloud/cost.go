package cloud

import (
	"fmt"
	"time"
)

// CostTracker tracks ongoing infrastructure costs
type CostTracker struct {
	startTime     time.Time
	currentCost   float64
	hourlyRate    float64
	maxTotalSpend float64
}

// NewCostTracker creates a new cost tracker
func NewCostTracker(hourlyRate, maxTotalSpend float64) *CostTracker {
	return &CostTracker{
		startTime:     time.Now(),
		currentCost:   0,
		hourlyRate:    hourlyRate,
		maxTotalSpend: maxTotalSpend,
	}
}

// UpdateCost calculates the current cost based on elapsed time
func (ct *CostTracker) UpdateCost() float64 {
	elapsed := time.Since(ct.startTime).Hours()
	ct.currentCost = elapsed * ct.hourlyRate
	return ct.currentCost
}

// GetCurrentCost returns the current accumulated cost
func (ct *CostTracker) GetCurrentCost() float64 {
	return ct.UpdateCost()
}

// GetElapsedTime returns the elapsed time since tracking started
func (ct *CostTracker) GetElapsedTime() time.Duration {
	return time.Since(ct.startTime)
}

// CheckLimits validates that current cost is within limits
func (ct *CostTracker) CheckLimits() error {
	currentCost := ct.UpdateCost()

	if ct.maxTotalSpend > 0 && currentCost > ct.maxTotalSpend {
		return fmt.Errorf("current cost ($%.2f) exceeds maximum total spend ($%.2f)",
			currentCost, ct.maxTotalSpend)
	}

	return nil
}

// GetEstimatedFinalCost estimates final cost for a given duration
func (ct *CostTracker) GetEstimatedFinalCost(duration time.Duration) float64 {
	return duration.Hours() * ct.hourlyRate
}

// GetCostSummary returns a formatted cost summary
func (ct *CostTracker) GetCostSummary() string {
	elapsed := ct.GetElapsedTime()
	currentCost := ct.GetCurrentCost()

	hours := int(elapsed.Hours())
	minutes := int(elapsed.Minutes()) % 60

	return fmt.Sprintf("Elapsed: %dh %dm | Current Cost: $%.2f | Hourly Rate: $%.2f/hr",
		hours, minutes, currentCost, ct.hourlyRate)
}
