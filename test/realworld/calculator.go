package realworld

import "fmt"

// Calculator は基本的な計算機能を提供します
type Calculator struct {
	result float64
}

// NewCalculator は新しい計算機インスタンスを作成します
func NewCalculator() *Calculator {
	return &Calculator{result: 0}
}

// Add は加算を実行します
func (c *Calculator) Add(value float64) *Calculator {
	c.result += value
	return c
}

// Multiply は乗算を実行します（人間が実装）
func (c *Calculator) Multiply(value float64) *Calculator {
	c.result *= value
	return c
}

// GetResult は現在の結果を取得します
func (c *Calculator) GetResult() float64 {
	return c.result
}

// Subtract は減算を実行します（AIが実装）
func (c *Calculator) Subtract(value float64) *Calculator {
	c.result -= value
	return c
}

// Divide は除算を実行します（AIが実装）
func (c *Calculator) Divide(value float64) *Calculator {
	if value != 0 {
		c.result /= value
	} else {
		fmt.Println("エラー: ゼロで割ることはできません")
	}
	return c
}

// Reset は計算機をリセットします（AIが実装）
func (c *Calculator) Reset() *Calculator {
	c.result = 0
	return c
}

// Display は結果を表示します
func (c *Calculator) Display() {
	fmt.Printf("計算結果: %.2f\n", c.result)
}