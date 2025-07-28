package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ai-code-tracker/aict/internal/i18n"
	"github.com/ai-code-tracker/aict/internal/web"
	"github.com/ai-code-tracker/aict/internal/web/middleware"
	"github.com/gorilla/websocket"
)

// generateSimpleClientID は簡易クライアントIDを生成する
func generateSimpleClientID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// SimpleAPIHandler は独立したAPI エンドポイントを処理する
type SimpleAPIHandler struct {
	server   *web.StandaloneServer
	upgrader websocket.Upgrader
}

// NewSimpleAPIHandler は新しい独立APIハンドラーを作成する
func NewSimpleAPIHandler(server *web.StandaloneServer) *SimpleAPIHandler {
	return &SimpleAPIHandler{
		server: server,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// 開発環境では全てのオリジンを許可
				return true
			},
		},
	}
}

// HandleStats は統計データを返すAPIエンドポイント
func (h *SimpleAPIHandler) HandleStats() http.Handler {
	return middleware.JSON(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 統計を取得
		stats, err := h.server.GetStats(r.Context(), nil)
		if err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, 
				i18n.T("statistics_fetch_failed"))
			return
		}
		
		// レスポンスを作成
		response := map[string]interface{}{
			"stats":     stats,
			"timestamp": time.Now(),
		}
		
		json.NewEncoder(w).Encode(response)
	}))
}

// HandleBlame はファイルのblame情報を返すAPIエンドポイント
func (h *SimpleAPIHandler) HandleBlame() http.Handler {
	return middleware.JSON(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// URLパスからファイルパスを取得
		path := strings.TrimPrefix(r.URL.Path, "/api/blame/")
		if path == "" {
			middleware.WriteJSONError(w, http.StatusBadRequest, 
				i18n.T("missing_required_argument", "file path"))
			return
		}
		
		// blame情報を取得
		blame, err := h.server.GetBlame(r.Context(), path)
		if err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, 
				i18n.T("blame_fetch_failed"))
			return
		}
		
		// レスポンスを作成
		response := map[string]interface{}{
			"blame":     blame,
			"file_path": path,
			"timestamp": time.Now(),
		}
		
		json.NewEncoder(w).Encode(response)
	}))
}

// HandleContributors は貢献者リストを返すAPIエンドポイント
func (h *SimpleAPIHandler) HandleContributors() http.Handler {
	return middleware.JSON(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contributors, err := h.server.GetContributors(r.Context())
		if err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, 
				i18n.T("contributor_fetch_failed"))
			return
		}
		
		// レスポンスを作成
		response := map[string]interface{}{
			"contributors": contributors,
			"count":        len(contributors),
			"timestamp":    time.Now(),
		}
		
		json.NewEncoder(w).Encode(response)
	}))
}

// HandleTimeline はタイムライン情報を返すAPIエンドポイント
func (h *SimpleAPIHandler) HandleTimeline() http.Handler {
	return middleware.JSON(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 制限数を取得（デフォルト: 100）
		limit := 100
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}
		
		events, err := h.server.GetTimeline(r.Context(), limit)
		if err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, 
				"Failed to fetch timeline")
			return
		}
		
		// レスポンスを作成
		response := map[string]interface{}{
			"events":    events,
			"count":     len(events),
			"limit":     limit,
			"timestamp": time.Now(),
		}
		
		json.NewEncoder(w).Encode(response)
	}))
}

// HandleFiles はファイル統計を返すAPIエンドポイント
func (h *SimpleAPIHandler) HandleFiles() http.Handler {
	return middleware.JSON(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fileStats, err := h.server.GetFileStats(r.Context())
		if err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, 
				"Failed to fetch file stats")
			return
		}
		
		// レスポンスを作成
		response := map[string]interface{}{
			"files":     fileStats,
			"count":     len(fileStats),
			"timestamp": time.Now(),
		}
		
		json.NewEncoder(w).Encode(response)
	}))
}

// HandleHealth はヘルスチェックエンドポイント
func (h *SimpleAPIHandler) HandleHealth() http.Handler {
	return middleware.JSON(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		healthy := h.server.IsHealthy()
		status := "ok"
		statusCode := http.StatusOK
		
		if !healthy {
			status = "error"  
			statusCode = http.StatusServiceUnavailable
		}
		
		response := map[string]interface{}{
			"status":    status,
			"timestamp": time.Now(),
			"version":   "0.1.0",
		}
		
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(response)
	}))
}

// HandleWebSocket はWebSocket接続を処理する
func (h *SimpleAPIHandler) HandleWebSocket() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// WebSocket接続にアップグレード
		conn, err := h.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		
		// クライアントID生成（簡易実装）
		clientID := generateSimpleClientID()
		
		// リアルタイム更新を購読
		updates := h.server.Subscribe(clientID)
		defer h.server.Unsubscribe(clientID)
		
		// 初期データを送信
		h.sendInitialData(conn)
		
		// WebSocketメッセージループ
		go h.handleWebSocketMessages(conn, clientID)
		
		// 更新イベントを送信
		for update := range updates {
			if err := conn.WriteJSON(update); err != nil {
				break
			}
		}
	})
}

// sendInitialData は初期データをWebSocketで送信する
func (h *SimpleAPIHandler) sendInitialData(conn *websocket.Conn) {
	// 統計データを送信
	if stats, err := h.server.GetStats(nil, nil); err == nil {
		event := &web.UpdateEvent{
			Type:      "initial_stats",
			Timestamp: time.Now(),
			Data:      stats,
		}
		conn.WriteJSON(event)
	}
	
	// 貢献者データを送信
	if contributors, err := h.server.GetContributors(nil); err == nil {
		event := &web.UpdateEvent{
			Type:      "initial_contributors", 
			Timestamp: time.Now(),
			Data:      contributors,
		}
		conn.WriteJSON(event)
	}
}

// handleWebSocketMessages はWebSocketメッセージを処理する
func (h *SimpleAPIHandler) handleWebSocketMessages(conn *websocket.Conn, clientID string) {
	for {
		var message map[string]interface{}
		if err := conn.ReadJSON(&message); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// ログは出力しない（開発用）
			}
			break
		}
		
		// メッセージタイプに応じて処理
		switch message["type"] {
		case "ping":
			// Pongを送信
			conn.WriteJSON(map[string]interface{}{
				"type":      "pong",
				"timestamp": time.Now(),
			})
		}
	}
}