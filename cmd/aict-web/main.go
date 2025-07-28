package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ai-code-tracker/aict/internal/i18n"
	"github.com/ai-code-tracker/aict/internal/web"
	"github.com/ai-code-tracker/aict/internal/web/handlers"
	"github.com/ai-code-tracker/aict/internal/web/middleware"
)

const (
	// DefaultPort はデフォルトのサーバーポート
	DefaultPort = "8080"
	// DefaultDataDir はデフォルトのデータディレクトリ
	DefaultDataDir = ".git/ai-tracker"
)

func main() {
	var (
		port    = flag.String("port", DefaultPort, "Server port")
		dataDir = flag.String("data", "", "Data directory (default: .git/ai-tracker in current or parent directories)")
		lang    = flag.String("lang", "ja", "Default language (ja|en)")
		debug   = flag.Bool("debug", false, "Enable debug mode")
	)
	flag.Parse()

	// 国際化システムを初期化
	i18n.Initialize()
	if *lang != "" {
		i18n.SetLocale(i18n.Locale(*lang))
	}

	// データディレクトリを決定
	if *dataDir == "" {
		dir, err := findDataDirectory()
		if err != nil {
			log.Printf("Warning: %v. Using current directory as fallback.", err)
			*dataDir = "."
		} else {
			*dataDir = dir
		}
	}

	// Webサーバーを初期化（簡略化版を使用）
	server := web.NewSimpleWebServer(&web.Config{
		DataDir: *dataDir,
		Debug:   *debug,
		Lang:    *lang,
	})

	// ルーターを設定
	router := setupRoutes(server)

	log.Printf("🌐 AI Code Tracker Web Dashboard starting on port %s", *port)
	log.Printf("📁 Data directory: %s", *dataDir)
	log.Printf("🗣️  Language: %s", *lang)
	if *debug {
		log.Printf("🐛 Debug mode enabled")
	}

	if err := http.ListenAndServe(":"+*port, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// setupRoutes はHTTPルーターを設定する
func setupRoutes(server *web.StandaloneServer) http.Handler {
	mux := http.NewServeMux()

	// ミドルウェアを適用
	handler := middleware.Chain(
		mux,
		middleware.Logger,
		middleware.CORS,
		middleware.Security,
		middleware.I18n,
	)

	// 静的ファイル
	mux.Handle("/static/", http.StripPrefix("/static/", 
		http.FileServer(http.Dir("web/static/"))))

	// API エンドポイント
	apiHandler := handlers.NewSimpleAPIHandler(server)
	mux.Handle("/api/stats", apiHandler.HandleStats())
	mux.Handle("/api/blame/", apiHandler.HandleBlame())
	mux.Handle("/api/contributors", apiHandler.HandleContributors())
	mux.Handle("/api/timeline", apiHandler.HandleTimeline())
	mux.Handle("/api/files", apiHandler.HandleFiles())
	mux.Handle("/api/health", apiHandler.HandleHealth())

	// WebSocket エンドポイント（リアルタイム更新用）
	mux.Handle("/ws", apiHandler.HandleWebSocket())

	// ダッシュボードページ
	dashboardHandler := handlers.NewSimpleDashboardHandler(server)
	mux.Handle("/", dashboardHandler.HandleIndex())
	mux.Handle("/dashboard", dashboardHandler.HandleDashboard())
	mux.Handle("/contributors", dashboardHandler.HandleContributors())
	mux.Handle("/files", dashboardHandler.HandleFiles())
	mux.Handle("/timeline", dashboardHandler.HandleTimeline())
	mux.Handle("/settings", dashboardHandler.HandleSettings())

	return handler
}

// findDataDirectory は.git/ai-trackerディレクトリを探す
func findDataDirectory() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// 現在のディレクトリから上に向かって.git/ai-trackerを探す
	dir := currentDir
	for {
		gitDir := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			trackerDir := filepath.Join(gitDir, "ai-tracker")
			if _, err := os.Stat(trackerDir); err == nil {
				return trackerDir, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // ルートディレクトリに到達
		}
		dir = parent
	}

	return "", os.ErrNotExist
}