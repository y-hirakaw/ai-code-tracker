# AI Code Tracker - REST API リファレンス

## 📋 目次

- [概要](#概要)
- [認証](#認証)
- [エンドポイント一覧](#エンドポイント一覧)
- [データ形式](#データ形式)
- [WebSocket](#websocket)
- [エラーハンドリング](#エラーハンドリング)
- [使用例](#使用例)

## 概要

AI Code Tracker Web Dashboard は、統計データとリアルタイム更新にアクセスするための REST API を提供します。

### ベースURL
```
http://localhost:8080/api
```

### Content-Type
```
application/json
```

## 認証

現在のバージョンでは認証は不要です（ローカル実行のため）。

## エンドポイント一覧

### ヘルスチェック

#### `GET /api/health`

サーバーの健全性を確認します。

**レスポンス:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-20T15:30:45Z",
  "version": "0.1.0"
}
```

**ステータスコード:**
- `200 OK`: サーバーが正常
- `503 Service Unavailable`: サーバーに問題がある場合

### 統計データ

#### `GET /api/stats`

プロジェクトの統計データを取得します。

**レスポンス:**
```json
{
  "stats": {
    "total_lines": 1500,
    "ai_lines": 900,
    "human_lines": 600,
    "file_count": 25,
    "last_updated": "2024-01-20T15:30:45Z",
    "file_stats": [
      {
        "path": "src/main.go",
        "total_lines": 250,
        "ai_lines": 150,
        "human_lines": 100
      }
    ],
    "contributors": [
      {
        "name": "Claude Sonnet 4",
        "type": "ai",
        "lines": 900
      },
      {
        "name": "Developer",
        "type": "human", 
        "lines": 600
      }
    ]
  },
  "timestamp": "2024-01-20T15:30:45Z"
}
```

### 貢献者リスト

#### `GET /api/contributors`

プロジェクトの貢献者情報を取得します。

**レスポンス:**
```json
{
  "contributors": [
    {
      "name": "Claude Sonnet 4",
      "type": "ai",
      "lines": 900
    },
    {
      "name": "Developer",
      "type": "human",
      "lines": 600
    }
  ],
  "count": 2,
  "timestamp": "2024-01-20T15:30:45Z"
}
```

### ファイル統計

#### `GET /api/files`

ファイル別の統計データを取得します。

**レスポンス:**
```json
{
  "files": [
    {
      "path": "src/main.go",
      "total_lines": 250,
      "ai_lines": 150,
      "human_lines": 100
    },
    {
      "path": "src/handler.go",
      "total_lines": 180,
      "ai_lines": 120,
      "human_lines": 60
    }
  ],
  "count": 2,
  "timestamp": "2024-01-20T15:30:45Z"
}
```

### タイムライン

#### `GET /api/timeline`

開発タイムラインを取得します。

**クエリパラメータ:**
- `limit` (optional): 取得する最大件数（デフォルト: 100）

**例:** `GET /api/timeline?limit=50`

**レスポンス:**
```json
{
  "events": [
    {
      "id": "event-1",
      "timestamp": "2024-01-20T14:30:45Z",
      "type": "ai_edit",
      "author": "Claude Sonnet 4",
      "description": "Refactored main function",
      "files": ["src/main.go"]
    },
    {
      "id": "event-2", 
      "timestamp": "2024-01-20T12:15:30Z",
      "type": "human_edit",
      "author": "Developer",
      "description": "Fixed bug in error handling",
      "files": ["src/handler.go"]
    }
  ],
  "count": 2,
  "limit": 50,
  "timestamp": "2024-01-20T15:30:45Z"
}
```

### Blame情報

#### `GET /api/blame/{file_path}`

特定ファイルのblame情報を取得します。

**例:** `GET /api/blame/src/main.go`

**レスポンス:**
```json
{
  "blame": {
    "file_path": "src/main.go",
    "lines": [
      {
        "line_number": 1,
        "content": "package main",
        "author": "Developer",
        "author_type": "human",
        "timestamp": "2024-01-20T10:00:00Z"
      },
      {
        "line_number": 2,
        "content": "",
        "author": "Developer", 
        "author_type": "human",
        "timestamp": "2024-01-20T10:00:00Z"
      },
      {
        "line_number": 3,
        "content": "import \"fmt\"",
        "author": "Claude Sonnet 4",
        "author_type": "ai",
        "timestamp": "2024-01-20T14:30:00Z"
      }
    ],
    "summary": {
      "total_lines": 3,
      "ai_lines": 1,
      "human_lines": 2,
      "contributors": {
        "Developer": 2,
        "Claude Sonnet 4": 1
      }
    }
  },
  "file_path": "src/main.go",
  "timestamp": "2024-01-20T15:30:45Z"
}
```

## WebSocket

### エンドポイント

#### `WS /ws`

リアルタイム更新を受信するためのWebSocket接続。

### 接続例（JavaScript）
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = function() {
    console.log('WebSocket connected');
};

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Received:', data);
};
```

### メッセージ形式

#### 受信メッセージ
```json
{
  "type": "stats_updated",
  "timestamp": "2024-01-20T15:30:45Z",
  "data": {
    "total_lines": 1500,
    "ai_lines": 900,
    "human_lines": 600
  }
}
```

#### 送信メッセージ（Ping）
```json
{
  "type": "ping"
}
```

#### 受信メッセージ（Pong）
```json
{
  "type": "pong",
  "timestamp": "2024-01-20T15:30:45Z"
}
```

## データ形式

### 統計データ
```typescript
interface Stats {
  total_lines: number;
  ai_lines: number;
  human_lines: number;
  file_count: number;
  last_updated: string;
  file_stats: FileInfo[];
  contributors: Contributor[];
}

interface FileInfo {
  path: string;
  total_lines: number;
  ai_lines: number;
  human_lines: number;
}

interface Contributor {
  name: string;
  type: "ai" | "human";
  lines: number;
}
```

### タイムラインイベント
```typescript
interface TimelineEvent {
  id: string;
  timestamp: string;
  type: "ai_edit" | "human_edit" | "commit";
  author: string;
  description: string;
  files: string[];
}
```

### Blame情報
```typescript
interface BlameResult {
  file_path: string;
  lines: BlameLine[];
  summary: BlameSummary;
}

interface BlameLine {
  line_number: number;
  content: string;
  author: string;
  author_type: "ai" | "human";
  timestamp: string;
}

interface BlameSummary {
  total_lines: number;
  ai_lines: number;
  human_lines: number;
  contributors: Record<string, number>;
}
```

## エラーハンドリング

### エラーレスポンス形式
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "統計データの取得に失敗しました",
    "details": "具体的なエラー詳細"
  },
  "timestamp": "2024-01-20T15:30:45Z"
}
```

### ステータスコード
- `200 OK`: 成功
- `400 Bad Request`: リクエストが不正
- `404 Not Found`: リソースが見つからない
- `500 Internal Server Error`: サーバー内部エラー
- `503 Service Unavailable`: サービス利用不可

## 使用例

### cURLを使用した基本的なAPI呼び出し

```bash
# ヘルスチェック
curl http://localhost:8080/api/health

# 統計データを取得
curl http://localhost:8080/api/stats

# 貢献者リストを取得
curl http://localhost:8080/api/contributors

# タイムラインを取得（最新50件）
curl "http://localhost:8080/api/timeline?limit=50"

# 特定ファイルのblame情報を取得
curl http://localhost:8080/api/blame/src/main.go
```

### JavaScriptでのAPI使用例

```javascript
// 統計データを取得
async function getStats() {
    try {
        const response = await fetch('/api/stats');
        const data = await response.json();
        console.log('統計データ:', data.stats);
    } catch (error) {
        console.error('エラー:', error);
    }
}

// 貢献者データを取得してチャートに表示
async function loadContributorsChart() {
    const response = await fetch('/api/contributors');
    const data = await response.json();
    
    const chartData = {
        labels: data.contributors.map(c => c.name),
        datasets: [{
            data: data.contributors.map(c => c.lines),
            backgroundColor: data.contributors.map(c => 
                c.type === 'ai' ? '#ff6b6b' : '#4ecdc4'
            )
        }]
    };
    
    // Chart.jsで表示
    new Chart(ctx, {
        type: 'doughnut',
        data: chartData
    });
}
```

### PythonでのAPI使用例

```python
import requests
import json

# API基底URL
BASE_URL = "http://localhost:8080/api"

def get_stats():
    """統計データを取得"""
    response = requests.get(f"{BASE_URL}/stats")
    if response.status_code == 200:
        return response.json()
    else:
        raise Exception(f"API エラー: {response.status_code}")

def get_file_blame(file_path):
    """ファイルのblame情報を取得"""
    response = requests.get(f"{BASE_URL}/blame/{file_path}")
    if response.status_code == 200:
        return response.json()
    else:
        raise Exception(f"API エラー: {response.status_code}")

# 使用例
stats = get_stats()
print(f"総行数: {stats['stats']['total_lines']}")
print(f"AI生成: {stats['stats']['ai_lines']}")
print(f"人間作成: {stats['stats']['human_lines']}")
```