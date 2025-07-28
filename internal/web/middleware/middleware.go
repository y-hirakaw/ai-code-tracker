package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ai-code-tracker/aict/internal/i18n"
)

// Middleware はHTTPミドルウェアの型
type Middleware func(http.Handler) http.Handler

// Chain は複数のミドルウェアを連鎖させる
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// Logger はリクエストログを出力するミドルウェア
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// カスタムResponseWriterでステータスコードをキャプチャ
		ww := &responseWriter{ResponseWriter: w}
		
		next.ServeHTTP(ww, r)
		
		duration := time.Since(start)
		
		log.Printf("%s %s %d %v %s",
			r.Method,
			r.URL.Path,
			ww.statusCode,
			duration,
			r.UserAgent(),
		)
	})
}

// responseWriter はステータスコードをキャプチャするためのラッパー
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// CORS はCORSヘッダーを設定するミドルウェア
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 開発環境でのCORS設定
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Language")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		
		// プリフライトリクエストの処理
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// Security はセキュリティヘッダーを設定するミドルウェア
func Security(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// セキュリティヘッダー
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", 
			"default-src 'self'; "+
			"script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; "+
			"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; "+
			"img-src 'self' data:; "+
			"connect-src 'self' ws: wss:")
		
		next.ServeHTTP(w, r)
	})
}

// I18n は国際化のコンテキストを設定するミドルウェア
func I18n(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 言語を決定する優先順位:
		// 1. X-Language ヘッダー
		// 2. lang クエリパラメータ
		// 3. Accept-Language ヘッダー
		// 4. デフォルト言語
		
		var locale i18n.Locale = i18n.LocaleJA // デフォルト
		
		// X-Language ヘッダーをチェック
		if lang := r.Header.Get("X-Language"); lang != "" {
			if lang == "en" {
				locale = i18n.LocaleEN
			} else if lang == "ja" {
				locale = i18n.LocaleJA
			}
		} else if lang := r.URL.Query().Get("lang"); lang != "" {
			// lang クエリパラメータをチェック
			if lang == "en" {
				locale = i18n.LocaleEN
			} else if lang == "ja" {
				locale = i18n.LocaleJA
			}
		} else if acceptLang := r.Header.Get("Accept-Language"); acceptLang != "" {
			// Accept-Language ヘッダーをパース
			if strings.HasPrefix(acceptLang, "en") {
				locale = i18n.LocaleEN
			}
		}
		
		// コンテキストに言語情報を設定
		ctx := context.WithValue(r.Context(), "locale", locale)
		
		// 一時的にグローバル言語を設定（スレッドセーフではないが、開発用）
		originalLocale := i18n.GetLocale()
		i18n.SetLocale(locale)
		defer i18n.SetLocale(originalLocale)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// JSON はJSONレスポンス用のヘッダーを設定するミドルウェア
func JSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

// RateLimit は簡単なレート制限ミドルウェア
func RateLimit(requestsPerMinute int) Middleware {
	// 簡単なin-memoryレート制限（本格運用ではRedis等を使用）
	clients := make(map[string][]time.Time)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)
			now := time.Now()
			
			// 古いエントリをクリーンアップ
			if times, exists := clients[clientIP]; exists {
				validTimes := []time.Time{}
				for _, t := range times {
					if now.Sub(t) < time.Minute {
						validTimes = append(validTimes, t)
					}
				}
				clients[clientIP] = validTimes
			}
			
			// レート制限チェック
			if len(clients[clientIP]) >= requestsPerMinute {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			
			// 現在のリクエストを記録
			clients[clientIP] = append(clients[clientIP], now)
			
			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP はクライアントのIPアドレスを取得する
func getClientIP(r *http.Request) string {
	// X-Forwarded-For ヘッダーをチェック
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// X-Real-IP ヘッダーをチェック
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}
	
	// RemoteAddr を使用
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	
	return ip
}

// Recover はパニックを捕捉するミドルウェア
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

// SetHeaders は共通ヘッダーを設定するミドルウェア
func SetHeaders(headers map[string]string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for key, value := range headers {
				w.Header().Set(key, value)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// BasicAuth は基本認証ミドルウェア
func BasicAuth(username, password string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 開発環境では認証をスキップ
			if username == "" || password == "" {
				next.ServeHTTP(w, r)
				return
			}
			
			user, pass, ok := r.BasicAuth()
			if !ok || user != username || pass != password {
				w.Header().Set("WWW-Authenticate", `Basic realm="AI Code Tracker"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// GetLocaleFromContext はコンテキストから言語情報を取得する
func GetLocaleFromContext(ctx context.Context) i18n.Locale {
	if locale, ok := ctx.Value("locale").(i18n.Locale); ok {
		return locale
	}
	return i18n.LocaleJA // デフォルト
}

// SetContentType は Content-Type ヘッダーを設定するヘルパー
func SetContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
}

// WriteJSONError はJSON形式のエラーレスポンスを書き込む
func WriteJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"error": "%s", "status": %d}`, message, statusCode)
}