package test

// AI-generated Calculator struct
type Calculator struct {
	result float64
}

// AI-generated constructor
func NewCalculator() *Calculator {
	return &Calculator{result: 0.0}
}

// AI-generated methods for basic operations
func (c *Calculator) Add(value float64) *Calculator {
	c.result += value
	return c
}

func (c *Calculator) Subtract(value float64) *Calculator {
	c.result -= value
	return c
}

func (c *Calculator) Multiply(value float64) *Calculator {
	c.result *= value
	return c
}

func (c *Calculator) Divide(value float64) *Calculator {
	if value != 0 {
		c.result /= value
	}
	return c
}

func (c *Calculator) GetResult() float64 {
	return c.result
}

func (c *Calculator) Reset() *Calculator {
	c.result = 0.0
	return c
}