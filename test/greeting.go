package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Greeting は挨拶メッセージを管理する構造体
type Greeting struct {
	Message   string
	Timestamp time.Time
	Language  string
}

// NewGreeting は新しい挨拶を作成する
func NewGreeting(message, language string) *Greeting {
	return &Greeting{
		Message:   message,
		Timestamp: time.Now(),
		Language:  language,
	}
}

// Display は挨拶を表示する
func (g *Greeting) Display() {
	fmt.Printf("[%s] %s (%s)\n", 
		g.Timestamp.Format("2006-01-02 15:04:05"), 
		g.Message, 
		g.Language)
}

// LogToFile は挨拶をファイルにログ出力する
func (g *Greeting) LogToFile(filename string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("ファイルのオープンに失敗: %w", err)
	}
	defer file.Close()

	logger := log.New(file, "", log.LstdFlags)
	logger.Printf("%s (%s)", g.Message, g.Language)
	
	return nil
}

func main() {
	// 複数言語の挨拶を作成
	greetings := []*Greeting{
		NewGreeting("こんにちは、世界！", "日本語"),
		NewGreeting("Hello, World!", "English"),
		NewGreeting("Bonjour, le monde!", "Français"),
		NewGreeting("¡Hola, mundo!", "Español"),
		NewGreeting("Hallo, Welt!", "Deutsch"),
	}

	// コンソールに表示
	fmt.Println("=== AI Code Tracker テスト用挨拶プログラム ===")
	for _, greeting := range greetings {
		greeting.Display()
	}

	// ログファイルに出力
	logFile := "test/greetings.log"
	fmt.Printf("\nログファイルに出力中: %s\n", logFile)
	
	for _, greeting := range greetings {
		if err := greeting.LogToFile(logFile); err != nil {
			fmt.Printf("ログ出力エラー: %v\n", err)
		}
	}

	fmt.Println("✅ 挨拶プログラム実行完了")
	
	// 追加: 実行時間の表示
	fmt.Printf("実行完了時刻: %s\n", time.Now().Format("2006-01-02 15:04:05"))
}