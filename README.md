# Go LiveView

Phoenix LiveViewのGo実装。サーバーサイドレンダリングとWebSocketを使用したリアルタイムWebアプリケーションを構築できます。

## 機能

- ✅ **LiveViewライフサイクル** - Mount, HandleEvent, HandleParams, Render
- ✅ **WebSocket通信** - gorilla/websocketによるリアルタイム通信
- ✅ **DOM差分計算** - 最小限のペイロードで高速更新
- ✅ **型安全なテンプレート** - a-h/templによる型安全なHTML生成
- ✅ **PubSub** - ローカル/Redis対応のブロードキャスト
- ✅ **セッション管理** - HMAC署名トークン
- ✅ **プロパティベーステスト** - gopterによる包括的テスト

## クイックスタート

### インストール

```bash
go get github.com/fu2hito/go-liveview
```

### 基本的なLiveViewの作成

```go
package main

import (
    "net/url"
    "github.com/a-h/templ"
    "github.com/fu2hito/go-liveview"
)

// Counter LiveView
type Counter struct {
    count int
}

func (c *Counter) Mount(ctx *liveview.Context, params url.Values) error {
    c.count = 0
    ctx.Assign("count", c.count)
    return nil
}

func (c *Counter) HandleEvent(ctx *liveview.Context, event string, payload map[string]interface{}) error {
    switch event {
    case "inc":
        c.count++
    case "dec":
        c.count--
    }
    ctx.Assign("count", c.count)
    return nil
}

func (c *Counter) Render(ctx *liveview.Context) templ.Component {
    count, _ := ctx.Get("count")
    return counterTemplate(count.(int))
}
```

### サーバーの設定

```go
func main() {
    // WebSocketサーバーの作成
    wsServer := socket.NewServer()
    
    // LiveViewマネージャーの作成
    manager := liveview.NewManager(wsServer)
    
    // LiveViewの登録
    manager.Register("counter", func() liveview.LiveView {
        return &Counter{}
    })
    
    // HTTPハンドラーの作成
    handler := liveview.NewHandler(manager, wsServer, liveview.HandlerOptions{
        Template: liveview.DefaultTemplate("Counter", "csrf-token"),
    })
    
    http.Handle("/", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 実行

```bash
# サンプルアプリケーションの実行
go run examples/counter/cmd/main.go

# ブラウザで開く
open http://localhost:8080
```

## アーキテクチャ

```
┌─────────────────────────────────────────────┐
│                  Browser                    │
│  ┌────────────────┐  ┌──────────────────┐  │
│  │   WebSocket    │  │  DOM Patching    │  │
│  │   Client       │  │  (morphdom)      │  │
│  └───────┬────────┘  └────────┬─────────┘  │
└──────────┼────────────────────┼────────────┘
           │                    │
           ▼                    ▼
┌─────────────────────────────────────────────┐
│                   Server                    │
│  ┌────────────────┐  ┌──────────────────┐  │
│  │  Socket Server │  │  LiveView        │  │
│  │  (WebSocket)   │──│  Manager         │  │
│  └───────┬────────┘  └────────┬─────────┘  │
│          │                    │            │
│  ┌───────▼────────┐  ┌───────▼─────────┐  │
│  │  Protocol      │  │  Render Engine  │  │
│  │  (Messages)    │  │  (templ + Diff) │  │
│  └────────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────┘
```

## 例

### Counter（カウンター）

増減ボタンを持つシンプルなカウンター。

```bash
go run examples/counter/cmd/main.go
```

### Chat（チャット）

PubSubを使用したリアルタイムチャット。

```bash
go run examples/chat/cmd/main.go
```

複数のブラウザで開くと、メッセージがリアルタイムで同期されます。

### Form（フォーム）

バリデーション付きフォーム。

```bash
go run examples/form/cmd/main.go
```

## テスト

### ユニットテスト

```bash
go test ./...
```

### プロパティベーステスト

```bash
go test ./tests/properties/... -v
```

### JavaScriptテスト

```bash
cd js && npm test
```

## パフォーマンス

- **Diff計算**: O(n) - 動的部分のみ差分を計算
- **メモリ**: セッションごとに状態を保持
- **通信**: 静的テンプレートは初回のみ、以降は動的部分のみ送信

## 設定

```go
opts := liveview.Options{
    Secret:            "your-secret-key",
    ReconnectStrategy: liveview.ReconnectReset,
    PubSub:            liveview.NewLocalPubSub(), // または RedisPubSub
}
```

## プロトコル

Phoenix LiveViewプロトコルに準拠:

- **Messages**: `topic`, `event`, `payload` を含むJSON
- **Diff Format**: `{"s": [...], "d": [...]}` (静的/動的部分)
- **Heartbeat**: 30秒間隔のping/pong
- **Reconnection**: 指数バックオフ付き自動再接続

## ライセンス

MIT License

## 貢献

貢献を歓迎します！詳細は CONTRIBUTING.md を参照してください。
