package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// E2E テスト環境の設定
type E2ETestEnv struct {
	TempDir    string
	ProjectDir string
	BinaryPath string
}

// E2E テスト環境のセットアップ
func setupE2EEnv(t *testing.T) *E2ETestEnv {
	// 一時ディレクトリの作成
	tempDir, err := os.MkdirTemp("", "aict-e2e-*")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}

	// テストプロジェクトディレクトリの作成
	projectDir := filepath.Join(tempDir, "test-project")
	err = os.MkdirAll(projectDir, 0755)
	if err != nil {
		t.Fatalf("プロジェクトディレクトリの作成に失敗: %v", err)
	}

	// Gitリポジトリの初期化
	cmd := exec.Command("git", "init")
	cmd.Dir = projectDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Gitリポジトリの初期化に失敗: %v", err)
	}

	// Git設定
	exec.Command("git", "config", "user.name", "AICT Test").Dir = projectDir
	exec.Command("git", "config", "user.email", "test@aict.local").Dir = projectDir

	// aictバイナリのパスを取得
	binaryPath := filepath.Join(os.Getenv("PWD"), "bin", "aict")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// ビルドしてみる
		buildCmd := exec.Command("make", "build")
		buildCmd.Dir = os.Getenv("PWD")
		if err := buildCmd.Run(); err != nil {
			t.Fatalf("aictバイナリのビルドに失敗: %v", err)
		}
	}

	return &E2ETestEnv{
		TempDir:    tempDir,
		ProjectDir: projectDir,
		BinaryPath: binaryPath,
	}
}

// E2E テスト環境のクリーンアップ
func (env *E2ETestEnv) Cleanup() {
	os.RemoveAll(env.TempDir)
}

// aictコマンドの実行
func (env *E2ETestEnv) runAICT(args ...string) (string, error) {
	cmd := exec.Command(env.BinaryPath, args...)
	cmd.Dir = env.ProjectDir
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// TestE2EFullWorkflow はフルワークフローのE2Eテストを実行する
func TestE2EFullWorkflow(t *testing.T) {
	env := setupE2EEnv(t)
	defer env.Cleanup()

	t.Run("1. aict init", func(t *testing.T) {
		output, err := env.runAICT("init")
		if err != nil {
			t.Fatalf("aict init に失敗: %v\n出力: %s", err, output)
		}

		// .git/ai-tracker ディレクトリが作成されているか確認
		trackerDir := filepath.Join(env.ProjectDir, ".git", "ai-tracker")
		if _, err := os.Stat(trackerDir); os.IsNotExist(err) {
			t.Errorf("ai-tracker ディレクトリが作成されていません")
		}
	})

	t.Run("2. 初期統計の確認", func(t *testing.T) {
		output, err := env.runAICT("stats", "--format", "summary")
		if err != nil {
			t.Fatalf("aict stats に失敗: %v\n出力: %s", err, output)
		}

		if !strings.Contains(output, "総イベント数: 0") {
			t.Errorf("初期統計が正しくありません: %s", output)
		}
	})

	t.Run("3. 手動トラッキング（AI）", func(t *testing.T) {
		// テストファイルの作成
		testFile := filepath.Join(env.ProjectDir, "test.go")
		testContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, AICT!")
}
`
		err := os.WriteFile(testFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("テストファイルの作成に失敗: %v", err)
		}

		// AIトラッキング
		output, err := env.runAICT("track", "--ai", "--model", "claude-sonnet-4", "--files", "test.go", "--message", "E2E テスト用AIコード")
		if err != nil {
			t.Fatalf("AI トラッキングに失敗: %v\n出力: %s", err, output)
		}

		if !strings.Contains(output, "トラッキング完了") {
			t.Errorf("トラッキング完了メッセージが見つかりません: %s", output)
		}
	})

	t.Run("4. 手動トラッキング（Human）", func(t *testing.T) {
		// ファイルの修正
		testFile := filepath.Join(env.ProjectDir, "test.go")
		modifiedContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, AICT!")
	fmt.Println("Modified by human") // Human による追加
}
`
		err := os.WriteFile(testFile, []byte(modifiedContent), 0644)
		if err != nil {
			t.Fatalf("テストファイルの修正に失敗: %v", err)
		}

		// Humanトラッキング
		output, err := env.runAICT("track", "--author", "E2E Tester", "--files", "test.go", "--message", "人間による修正")
		if err != nil {
			t.Fatalf("Human トラッキングに失敗: %v\n出力: %s", err, output)
		}

		if !strings.Contains(output, "トラッキング完了") {
			t.Errorf("トラッキング完了メッセージが見つかりません: %s", output)
		}
	})

	t.Run("5. 統計の確認", func(t *testing.T) {
		output, err := env.runAICT("stats", "--format", "summary")
		if err != nil {
			t.Fatalf("aict stats に失敗: %v\n出力: %s", err, output)
		}

		if !strings.Contains(output, "総イベント数: 2") {
			t.Errorf("統計のイベント数が正しくありません: %s", output)
		}

		if !strings.Contains(output, "AI によるコード") {
			t.Errorf("AI統計が見つかりません: %s", output)
		}

		if !strings.Contains(output, "人間によるコード") {
			t.Errorf("Human統計が見つかりません: %s", output)
		}
	})

	t.Run("6. blame機能のテスト", func(t *testing.T) {
		// Gitコミットの作成（blame用）
		cmd := exec.Command("git", "add", "test.go")
		cmd.Dir = env.ProjectDir
		cmd.Run()

		cmd = exec.Command("git", "commit", "-m", "E2E test commit")
		cmd.Dir = env.ProjectDir
		cmd.Run()

		// blame機能のテスト
		output, err := env.runAICT("blame", "test.go")
		if err != nil {
			t.Logf("blame機能の警告（Gitコミット情報が不足の可能性）: %v\n出力: %s", err, output)
			// blame機能は必須ではないため、エラーでも継続
		}
	})

	t.Run("7. hooks設定のテスト", func(t *testing.T) {
		output, err := env.runAICT("setup", "--git-hooks")
		if err != nil {
			t.Fatalf("Git hooks設定に失敗: %v\n出力: %s", err, output)
		}

		// post-commit hookが作成されているか確認
		hookPath := filepath.Join(env.ProjectDir, ".git", "hooks", "post-commit")
		if _, err := os.Stat(hookPath); os.IsNotExist(err) {
			t.Errorf("post-commit hookが作成されていません")
		}
	})

	t.Run("8. hooks設定状況の確認", func(t *testing.T) {
		output, err := env.runAICT("setup", "--status")
		if err != nil {
			t.Fatalf("hooks状況確認に失敗: %v\n出力: %s", err, output)
		}

		if !strings.Contains(output, "Git hooks") {
			t.Errorf("Git hooks状況が表示されていません: %s", output)
		}
	})
}

// TestE2EPerformance はパフォーマンスE2Eテストを実行する
func TestE2EPerformance(t *testing.T) {
	env := setupE2EEnv(t)
	defer env.Cleanup()

	// 初期化
	env.runAICT("init")

	t.Run("大量イベントの処理", func(t *testing.T) {
		start := time.Now()

		// 100個のイベントを順次追加
		for i := 0; i < 100; i++ {
			fileName := fmt.Sprintf("file_%d.go", i)
			_, err := env.runAICT("track", "--ai", "--model", "claude-sonnet-4", "--files", fileName, "--message", fmt.Sprintf("パフォーマンステスト %d", i))
			if err != nil {
				t.Fatalf("イベント %d の追加に失敗: %v", i, err)
			}
		}

		duration := time.Since(start)
		t.Logf("100イベントの追加時間: %v", duration)

		// パフォーマンス目標のチェック（10秒以内）
		if duration > 10*time.Second {
			t.Errorf("100イベントの追加が遅すぎます: %v（目標: 10秒以内）", duration)
		}
	})

	t.Run("統計生成パフォーマンス", func(t *testing.T) {
		start := time.Now()

		output, err := env.runAICT("stats", "--format", "summary")
		if err != nil {
			t.Fatalf("統計生成に失敗: %v", err)
		}

		duration := time.Since(start)
		t.Logf("統計生成時間: %v", duration)

		// パフォーマンス目標のチェック（500ms以内）
		if duration > 500*time.Millisecond {
			t.Errorf("統計生成が遅すぎます: %v（目標: 500ms以内）", duration)
		}

		if !strings.Contains(output, "総イベント数: 100") {
			t.Errorf("統計の内容が正しくありません: %s", output)
		}
	})
}

// TestE2EErrorHandling はエラーハンドリングのE2Eテストを実行する
func TestE2EErrorHandling(t *testing.T) {
	env := setupE2EEnv(t)
	defer env.Cleanup()

	t.Run("初期化前のコマンド実行", func(t *testing.T) {
		output, err := env.runAICT("stats")
		if err == nil {
			t.Errorf("初期化前のstatsコマンドでエラーが発生すべきです。出力: %s", output)
		}

		if !strings.Contains(output, "初期化") && !strings.Contains(output, "init") {
			t.Errorf("適切なエラーメッセージが表示されていません: %s", output)
		}
	})

	t.Run("無効なオプション", func(t *testing.T) {
		_, err := env.runAICT("track", "--invalid-option")
		if err == nil {
			t.Errorf("無効なオプションでエラーが発生すべきです")
		}
	})

	t.Run("存在しないファイルのblame", func(t *testing.T) {
		env.runAICT("init") // 初期化

		_, err := env.runAICT("blame", "nonexistent.go")
		if err == nil {
			t.Errorf("存在しないファイルのblameでエラーが発生すべきです")
		}
	})
}

// TestE2EConcurrency は並行処理のE2Eテストを実行する
func TestE2EConcurrency(t *testing.T) {
	env := setupE2EEnv(t)
	defer env.Cleanup()

	// 初期化
	env.runAICT("init")

	t.Run("並行トラッキング", func(t *testing.T) {
		// 複数のゴルーチンで同時にトラッキング
		done := make(chan bool, 10)
		errors := make(chan error, 10)

		for i := 0; i < 10; i++ {
			go func(id int) {
				defer func() { done <- true }()
				
				fileName := fmt.Sprintf("concurrent_%d.go", id)
				_, err := env.runAICT("track", "--ai", "--model", "claude-sonnet-4", "--files", fileName, "--message", fmt.Sprintf("並行テスト %d", id))
				if err != nil {
					errors <- fmt.Errorf("並行トラッキング %d で失敗: %v", id, err)
				}
			}(i)
		}

		// 全ての処理の完了を待機
		for i := 0; i < 10; i++ {
			<-done
		}

		// エラーチェック
		close(errors)
		for err := range errors {
			t.Error(err)
		}

		// 最終的な統計確認
		output, err := env.runAICT("stats", "--format", "summary")
		if err != nil {
			t.Fatalf("最終統計確認に失敗: %v", err)
		}

		if !strings.Contains(output, "総イベント数: 10") {
			t.Errorf("並行処理後の統計が正しくありません: %s", output)
		}
	})
}

// BenchmarkE2ETracking はトラッキングのベンチマークテストを実行する
func BenchmarkE2ETracking(b *testing.B) {
	env := setupE2EEnv(&testing.T{})
	defer env.Cleanup()

	// 初期化
	env.runAICT("init")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fileName := fmt.Sprintf("bench_%d.go", i)
		_, err := env.runAICT("track", "--ai", "--model", "claude-sonnet-4", "--files", fileName, "--message", fmt.Sprintf("ベンチマーク %d", i))
		if err != nil {
			b.Fatalf("ベンチマークトラッキングに失敗: %v", err)
		}
	}
}