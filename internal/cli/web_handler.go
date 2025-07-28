package cli

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/errors"
	"github.com/y-hirakaw/ai-code-tracker/internal/i18n"
)

// WebHandler はWebダッシュボードコマンドを処理する
type WebHandler struct{}

// NewWebHandler は新しいWebHandlerを作成する
func NewWebHandler() *WebHandler {
	return &WebHandler{}
}

// Handle はWebダッシュボードコマンドを実行する
func (h *WebHandler) Handle(args []string) error {
	var (
		port     = "8080"
		lang     = "ja"
		debug    = false
		dataDir  = ""
		openBrowser = true
	)

	// コマンドライン引数をパース
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--port", "-p":
			if i+1 < len(args) {
				port = args[i+1]
				i++
			}
		case "--lang", "-l":
			if i+1 < len(args) {
				lang = args[i+1]
				i++
			}
		case "--debug", "-d":
			debug = true
		case "--data":
			if i+1 < len(args) {
				dataDir = args[i+1]
				i++
			}
		case "--no-browser":
			openBrowser = false
		case "--help", "-h":
			return h.showHelp()
		}
	}

	// ポート番号の検証
	if portNum, err := strconv.Atoi(port); err != nil || portNum < 1 || portNum > 65535 {
		return errors.NewError(errors.ErrorTypeCommand, "invalid_port", port).
			WithSuggestions(i18n.T("suggestion_valid_port", "有効なポート番号は 1-65535 です"))
	}

	// 言語の検証
	if lang != "ja" && lang != "en" {
		return errors.NewError(errors.ErrorTypeCommand, "invalid_language", lang).
			WithSuggestions(i18n.T("suggestion_valid_languages", "利用可能な言語: ja, en"))
	}

	// Web サーバーの起動確認
	fmt.Printf("🌐 %s\n", i18n.T("starting_web_dashboard", "Webダッシュボードを起動しています"))
	fmt.Printf("   %s: %s\n", i18n.T("port", "ポート"), port)
	fmt.Printf("   %s: %s\n", i18n.T("language", "言語"), lang)
	if debug {
		fmt.Printf("   %s: %s\n", i18n.T("debug_mode", "デバッグモード"), i18n.T("enabled", "有効"))
	}
	if dataDir != "" {
		fmt.Printf("   %s: %s\n", i18n.T("data_directory", "データディレクトリ"), dataDir)
	}

	// aict-web コマンドを実行
	cmdArgs := []string{
		"-port", port,
		"-lang", lang,
	}
	
	if debug {
		cmdArgs = append(cmdArgs, "-debug")
	}
	
	if dataDir != "" {
		cmdArgs = append(cmdArgs, "-data", dataDir)
	}

	// aict-web バイナリのパスを取得
	webBinary := "aict-web"
	if runtime.GOOS == "windows" {
		webBinary = "aict-web.exe"
	}

	// バイナリが存在するかチェック
	if _, err := exec.LookPath(webBinary); err != nil {
		// 同じディレクトリにあるかチェック
		if _, err := os.Stat("./" + webBinary); err != nil {
			return errors.NewError(errors.ErrorTypeFile, "web_binary_not_found", webBinary).
				WithSuggestions(i18n.T("suggestion_build_web", "go build ./cmd/aict-web でWebバイナリをビルドしてください"))
		}
		webBinary = "./" + webBinary
	}

	// ブラウザを開く
	if openBrowser {
		go func() {
			// サーバー起動を待つ
			time.Sleep(2 * time.Second)
			if h.isServerRunning(port) {
				h.openBrowser(fmt.Sprintf("http://localhost:%s", port))
			}
		}()
	}

	// Web サーバーを実行
	cmd := exec.Command(webBinary, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Printf("🚀 %s http://localhost:%s\n", 
		i18n.T("dashboard_available", "ダッシュボードが利用可能です:"), port)
	fmt.Printf("💡 %s Ctrl+C\n", 
		i18n.T("stop_server_hint", "サーバーを停止するには"))

	return cmd.Run()
}

// showHelp はWebコマンドのヘルプを表示する
func (h *WebHandler) showHelp() error {
	fmt.Printf(`🌐 %s - Webダッシュボード

%s:
  aict web [options]

%s:
  -p, --port <port>     サーバーポート (デフォルト: 8080)
  -l, --lang <lang>     表示言語 (ja|en, デフォルト: ja)
  -d, --debug          デバッグモードを有効化
      --data <dir>     データディレクトリを指定
      --no-browser     ブラウザを自動で開かない
  -h, --help           このヘルプを表示

%s:
  aict web                          # デフォルト設定で起動
  aict web -p 3000                  # ポート3000で起動
  aict web -l en --debug            # 英語+デバッグモードで起動
  aict web --data /path/to/data     # カスタムデータディレクトリで起動
  aict web --no-browser             # ブラウザを開かずに起動

%s:
  - リアルタイム統計更新
  - 多言語対応インターフェース (日本語/英語)
  - レスポンシブデザイン
  - ファイル別詳細分析
  - 貢献者別統計
  - タイムライン表示
  - インタラクティブチャート

%s:
  http://localhost:8080             # デフォルトアクセスURL
  
`, 
		i18n.T("web_command_title", "aict web"),
		i18n.T("usage", "使用方法"),
		i18n.T("options", "オプション"),
		i18n.T("examples", "例"),
		i18n.T("features", "機能"),
		i18n.T("access", "アクセス"))

	return nil
}

// isServerRunning はサーバーが起動しているかチェックする
func (h *WebHandler) isServerRunning(port string) bool {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/api/health", port))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// openBrowser はブラウザでURLを開く
func (h *WebHandler) openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		log.Printf("Cannot open browser on %s platform", runtime.GOOS)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to open browser: %v", err)
	} else {
		fmt.Printf("🌐 %s %s\n", 
			i18n.T("browser_opened", "ブラウザで開きました:"), url)
	}
}

// TestWebConnection はWeb接続をテストする
func (h *WebHandler) TestWebConnection(port string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	
	// ヘルスチェック
	resp, err := client.Get(fmt.Sprintf("http://localhost:%s/api/health", port))
	if err != nil {
		return errors.NewError(errors.ErrorTypeNetwork, "web_connection_failed", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewError(errors.ErrorTypeNetwork, "web_server_error", 
			fmt.Sprintf("Status: %d", resp.StatusCode))
	}

	fmt.Printf("✅ %s\n", i18n.T("web_connection_ok", "Web接続テスト成功"))
	return nil
}