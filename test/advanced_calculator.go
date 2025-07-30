package test

import (
	"fmt"
	"math"
)

// AI-generated advanced calculator with scientific functions
type AdvancedCalculator struct {
	*Calculator
	history []float64
}

// AI-generated constructor for advanced calculator
func NewAdvancedCalculator() *AdvancedCalculator {
	return &AdvancedCalculator{
		Calculator: NewCalculator(),
		history:    make([]float64, 0),
	}
}

// AI-generated scientific operations
func (ac *AdvancedCalculator) Power(exponent float64) *AdvancedCalculator {
	ac.result = math.Pow(ac.result, exponent)
	ac.recordHistory()
	return ac
}

func (ac *AdvancedCalculator) SquareRoot() *AdvancedCalculator {
	if ac.result >= 0 {
		ac.result = math.Sqrt(ac.result)
		ac.recordHistory()
	}
	return ac
}

func (ac *AdvancedCalculator) Logarithm() *AdvancedCalculator {
	if ac.result > 0 {
		ac.result = math.Log(ac.result)
		ac.recordHistory()
	}
	return ac
}

func (ac *AdvancedCalculator) Sin() *AdvancedCalculator {
	ac.result = math.Sin(ac.result)
	ac.recordHistory()
	return ac
}

func (ac *AdvancedCalculator) Cos() *AdvancedCalculator {
	ac.result = math.Cos(ac.result)
	ac.recordHistory()
	return ac
}

func (ac *AdvancedCalculator) Tan() *AdvancedCalculator {
	ac.result = math.Tan(ac.result)
	ac.recordHistory()
	return ac
}

// AI-generated history management
func (ac *AdvancedCalculator) recordHistory() {
	ac.history = append(ac.history, ac.result)
}

func (ac *AdvancedCalculator) GetHistory() []float64 {
	return ac.history
}

func (ac *AdvancedCalculator) ClearHistory() *AdvancedCalculator {
	ac.history = make([]float64, 0)
	return ac
}

func (ac *AdvancedCalculator) PrintHistory() {
	fmt.Println("Calculation History:")
	for i, val := range ac.history {
		fmt.Printf("Step %d: %f\n", i+1, val)
	}
}