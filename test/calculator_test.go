package test

import "testing"

func TestCalculatorBasicOperations(t *testing.T) {
	calc := NewCalculator()
	
	result := calc.Add(10).Subtract(5).Multiply(2).GetResult()
	expected := 10.0
	
	if result != expected {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestCalculatorDivision(t *testing.T) {
	calc := NewCalculator()
	
	result := calc.Add(20).Divide(4).GetResult()
	expected := 5.0
	
	if result != expected {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestCalculatorReset(t *testing.T) {
	calc := NewCalculator()
	
	calc.Add(100).Reset()
	result := calc.GetResult()
	
	if result != 0.0 {
		t.Errorf("Expected 0, got %f", result)
	}
}