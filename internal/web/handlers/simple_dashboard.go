package handlers

import (
	"fmt"
	"net/http"

	"github.com/y-hirakaw/ai-code-tracker/internal/i18n"
	"github.com/y-hirakaw/ai-code-tracker/internal/web"
	"github.com/y-hirakaw/ai-code-tracker/internal/web/middleware"
)

// SimpleDashboardHandler は簡易ダッシュボードページを処理する
type SimpleDashboardHandler struct {
	server *web.StandaloneServer
}

// NewSimpleDashboardHandler は新しい簡易ダッシュボードハンドラーを作成する
func NewSimpleDashboardHandler(server *web.StandaloneServer) *SimpleDashboardHandler {
	return &SimpleDashboardHandler{
		server: server,
	}
}

// HandleIndex はインデックスページを処理する
func (h *SimpleDashboardHandler) HandleIndex() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ダッシュボードにリダイレクト
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	})
}

// HandleDashboard はダッシュボードページを処理する
func (h *SimpleDashboardHandler) HandleDashboard() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale := middleware.GetLocaleFromContext(r.Context())
		
		// 統計データを取得
		stats, err := h.server.GetStats(r.Context(), nil)
		if err != nil {
			http.Error(w, "Failed to load stats", http.StatusInternalServerError)
			return
		}
		
		h.renderDashboard(w, r, stats, string(locale))
	})
}

// HandleContributors は貢献者ページを処理する
func (h *SimpleDashboardHandler) HandleContributors() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale := middleware.GetLocaleFromContext(r.Context())
		
		contributors, err := h.server.GetContributors(r.Context())
		if err != nil {
			http.Error(w, "Failed to load contributors", http.StatusInternalServerError)
			return
		}
		
		h.renderContributors(w, r, contributors, string(locale))
	})
}

// HandleFiles はファイルページを処理する
func (h *SimpleDashboardHandler) HandleFiles() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale := middleware.GetLocaleFromContext(r.Context())
		
		fileStats, err := h.server.GetFileStats(r.Context())
		if err != nil {
			http.Error(w, "Failed to load file stats", http.StatusInternalServerError)
			return
		}
		
		h.renderFiles(w, r, fileStats, string(locale))
	})
}

// HandleTimeline はタイムラインページを処理する
func (h *SimpleDashboardHandler) HandleTimeline() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale := middleware.GetLocaleFromContext(r.Context())
		
		events, err := h.server.GetTimeline(r.Context(), 100)
		if err != nil {
			http.Error(w, "Failed to load timeline", http.StatusInternalServerError)
			return
		}
		
		h.renderTimeline(w, r, events, string(locale))
	})
}

// HandleSettings は設定ページを処理する
func (h *SimpleDashboardHandler) HandleSettings() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale := middleware.GetLocaleFromContext(r.Context())
		
		config := h.server.GetConfig()
		h.renderSettings(w, r, config, string(locale))
	})
}

// renderDashboard はダッシュボードHTMLを生成する
func (h *SimpleDashboardHandler) renderDashboard(w http.ResponseWriter, r *http.Request, stats *web.StandaloneStats, locale string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	title := i18n.T("dashboard")
	if locale == "en" {
		title = "Dashboard"
	}
	
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html lang="%s">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - AI Code Tracker</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.7.2/font/bootstrap-icons.css" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        body { background-color: #f8f9fa; }
        .card-header { background-color: #007bff; color: white; }
        .metric-card { transition: transform 0.2s; }
        .metric-card:hover { transform: translateY(-2px); }
        .sidebar { min-height: 100vh; background-color: #343a40; }
        .sidebar .nav-link { color: #adb5bd; }
        .sidebar .nav-link:hover { color: white; }
        .sidebar .nav-link.active { color: white; background-color: #007bff; }
    </style>
</head>
<body>
    <div class="container-fluid">
        <div class="row">
            <!-- サイドバー -->
            <nav class="col-md-2 d-none d-md-block sidebar">
                <div class="sidebar-sticky pt-3">
                    <h5 class="text-white px-3 mb-3">
                        <i class="bi bi-robot"></i> AI Code Tracker
                    </h5>
                    <ul class="nav flex-column">
                        <li class="nav-item">
                            <a class="nav-link active" href="/dashboard">
                                <i class="bi bi-speedometer2"></i> %s
                            </a>
                        </li>
                        <li class="nav-item">
                            <a class="nav-link" href="/contributors">
                                <i class="bi bi-people"></i> %s
                            </a>
                        </li>
                        <li class="nav-item">
                            <a class="nav-link" href="/files">
                                <i class="bi bi-file-code"></i> %s
                            </a>
                        </li>
                        <li class="nav-item">
                            <a class="nav-link" href="/timeline">
                                <i class="bi bi-clock-history"></i> %s
                            </a>
                        </li>
                        <li class="nav-item">
                            <a class="nav-link" href="/settings">
                                <i class="bi bi-gear"></i> %s
                            </a>
                        </li>
                    </ul>
                </div>
            </nav>

            <!-- メインコンテンツ -->
            <main class="col-md-10 ml-sm-auto px-md-4">
                <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
                    <h1 class="h2">%s</h1>
                    <div class="btn-toolbar mb-2 mb-md-0">
                        <div class="btn-group me-2">
                            <button type="button" class="btn btn-sm btn-outline-secondary" onclick="refreshData()">
                                <i class="bi bi-arrow-clockwise"></i> %s
                            </button>
                        </div>
                        <div class="btn-group">
                            <button type="button" class="btn btn-sm btn-outline-secondary dropdown-toggle" data-bs-toggle="dropdown">
                                <i class="bi bi-globe"></i> %s
                            </button>
                            <ul class="dropdown-menu">
                                <li><a class="dropdown-item" href="?lang=ja">🇯🇵 日本語</a></li>
                                <li><a class="dropdown-item" href="?lang=en">🇺🇸 English</a></li>
                            </ul>
                        </div>
                    </div>
                </div>

                <!-- メトリクスカード -->
                <div class="row mb-4">
                    <div class="col-md-3 mb-3">
                        <div class="card metric-card h-100">
                            <div class="card-body text-center">
                                <h5 class="card-title">%s</h5>
                                <h2 class="text-primary" id="total-lines">%d</h2>
                                <small class="text-muted">%s</small>
                            </div>
                        </div>
                    </div>
                    <div class="col-md-3 mb-3">
                        <div class="card metric-card h-100">
                            <div class="card-body text-center">
                                <h5 class="card-title">%s</h5>
                                <h2 class="text-success" id="ai-lines">%d</h2>
                                <small class="text-muted">%.1f%%</small>
                            </div>
                        </div>
                    </div>
                    <div class="col-md-3 mb-3">
                        <div class="card metric-card h-100">
                            <div class="card-body text-center">
                                <h5 class="card-title">%s</h5>
                                <h2 class="text-info" id="human-lines">%d</h2>
                                <small class="text-muted">%.1f%%</small>
                            </div>
                        </div>
                    </div>
                    <div class="col-md-3 mb-3">
                        <div class="card metric-card h-100">
                            <div class="card-body text-center">
                                <h5 class="card-title">%s</h5>
                                <h2 class="text-warning" id="total-files">%d</h2>
                                <small class="text-muted">%s</small>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- チャート -->
                <div class="row">
                    <div class="col-md-6 mb-4">
                        <div class="card">
                            <div class="card-header">
                                <h5 class="mb-0">%s</h5>
                            </div>
                            <div class="card-body">
                                <canvas id="ratio-chart"></canvas>
                            </div>
                        </div>
                    </div>
                    <div class="col-md-6 mb-4">
                        <div class="card">
                            <div class="card-header">
                                <h5 class="mb-0">%s</h5>
                            </div>
                            <div class="card-body">
                                <div class="table-responsive">
                                    <table class="table table-sm">
                                        <thead>
                                            <tr>
                                                <th>%s</th>
                                                <th>%s</th>
                                                <th>%s</th>
                                            </tr>
                                        </thead>
                                        <tbody>`,
		locale, title,
		i18n.T("dashboard"), i18n.T("contributors"), i18n.T("files"), i18n.T("timeline"), i18n.T("settings"),
		title, i18n.T("refresh"), i18n.T("language"),
		i18n.T("total_lines"), stats.TotalLines, i18n.T("lines_of_code"),
		i18n.T("ai_lines"), stats.AILines, float64(stats.AILines)/float64(stats.TotalLines)*100,
		i18n.T("human_lines"), stats.HumanLines, float64(stats.HumanLines)/float64(stats.TotalLines)*100,
		i18n.T("total_files"), stats.FileCount, i18n.T("tracked_files"),
		i18n.T("ai_human_ratio"), i18n.T("top_files"),
		i18n.T("file_path"), i18n.T("total_lines"), i18n.T("ai_percentage"))

	// ファイル統計をテーブルに表示
	for i, file := range stats.FileStats {
		if i >= 5 { // 上位5つのみ表示
			break
		}
		aiPercentage := float64(file.AILines) / float64(file.TotalLines) * 100
		fmt.Fprintf(w, `
                                            <tr>
                                                <td><a href="/api/blame/%s">%s</a></td>
                                                <td>%d</td>
                                                <td>
                                                    <div class="progress" style="height: 20px;">
                                                        <div class="progress-bar bg-success" role="progressbar" 
                                                             style="width: %.1f%%">
                                                            %.1f%%
                                                        </div>
                                                    </div>
                                                </td>
                                            </tr>`,
			file.Path, file.Path, file.TotalLines, aiPercentage, aiPercentage)
	}

	fmt.Fprintf(w, `
                                        </tbody>
                                    </table>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    </div>

    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js"></script>
    <script>
        // AIコード比率チャート
        const ratioCtx = document.getElementById('ratio-chart').getContext('2d');
        new Chart(ratioCtx, {
            type: 'doughnut',
            data: {
                labels: ['%s', '%s'],
                datasets: [{
                    data: [%d, %d],
                    backgroundColor: ['#28a745', '#17a2b8'],
                    borderWidth: 2
                }]
            },
            options: {
                responsive: true,
                plugins: {
                    legend: {
                        position: 'bottom',
                    }
                }
            }
        });
        
        // WebSocket接続
        function connectWebSocket() {
            const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
            const ws = new WebSocket(protocol + '//' + location.host + '/ws');
            
            ws.onmessage = function(event) {
                const data = JSON.parse(event.data);
                updateUI(data);
            };
            
            ws.onclose = function() {
                setTimeout(connectWebSocket, 5000);
            };
        }
        
        function updateUI(data) {
            if (data.type === 'stats_updated') {
                const stats = data.data;
                document.getElementById('total-lines').textContent = stats.total_lines;
                document.getElementById('ai-lines').textContent = stats.ai_lines;
                document.getElementById('human-lines').textContent = stats.human_lines;
                document.getElementById('total-files').textContent = stats.file_count;
            }
        }
        
        function refreshData() {
            location.reload();
        }
        
        connectWebSocket();
    </script>
</body>
</html>`,
		i18n.T("ai_code"), i18n.T("human_code"),
		stats.AILines, stats.HumanLines)
}

// 他のrender関数も同様に簡単な実装を提供
func (h *SimpleDashboardHandler) renderContributors(w http.ResponseWriter, r *http.Request, contributors []web.StandaloneContributor, locale string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<html><body><h1>Contributors</h1><p>Feature coming soon...</p><a href="/dashboard">Back to Dashboard</a></body></html>`)
}

func (h *SimpleDashboardHandler) renderFiles(w http.ResponseWriter, r *http.Request, files []web.StandaloneFileInfo, locale string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<html><body><h1>Files</h1><p>Feature coming soon...</p><a href="/dashboard">Back to Dashboard</a></body></html>`)
}

func (h *SimpleDashboardHandler) renderTimeline(w http.ResponseWriter, r *http.Request, events []web.StandaloneEvent, locale string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<html><body><h1>Timeline</h1><p>Feature coming soon...</p><a href="/dashboard">Back to Dashboard</a></body></html>`)
}

func (h *SimpleDashboardHandler) renderSettings(w http.ResponseWriter, r *http.Request, config *web.Config, locale string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<html><body><h1>Settings</h1><p>Feature coming soon...</p><a href="/dashboard">Back to Dashboard</a></body></html>`)
}