# Go LiveView - 使い方ガイド

## 基本的な使い方

### 1. シンプルなカウンター

```go
package main

import (
    "context"
    "io"
    "log"
    "net/http"
    "net/url"
    "strconv"

    "github.com/a-h/templ"
    "github.com/fu2hito/go-liveview"
    "github.com/fu2hito/go-liveview/internal/socket"
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

func (c *Counter) HandleParams(ctx *liveview.Context, params url.Values) error {
    return nil
}

func (c *Counter) Render(ctx *liveview.Context) templ.Component {
    return counterTemplate(c.count)
}

func counterTemplate(count int) templ.Component {
    return &simpleComponent{
        html: `<div>
            <h1>Count: ` + strconv.Itoa(count) + `</h1>
            <button phx-click="dec">-</button>
            <button phx-click="inc">+</button>
        </div>`,
    }
}

type simpleComponent struct {
    html string
}

func (s *simpleComponent) Render(ctx context.Context, w io.Writer) error {
    _, err := w.Write([]byte(s.html))
    return err
}

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
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 2. PubSubを使ったリアルタイム機能

```go
// Chat LiveView with PubSub
type Chat struct {
    messages []Message
    username string
}

func (c *Chat) Mount(ctx *liveview.Context, params url.Values) error {
    c.username = params.Get("username")
    if c.username == "" {
        c.username = "Anonymous"
    }
    
    ctx.Assign("messages", c.messages)
    ctx.Assign("username", c.username)
    
    // PubSubでブロードキャストを購読
    if broadcaster := ctx.GetBroadcaster(); broadcaster != nil {
        broadcaster.SubscribeContext(ctx, "chat:room", ctx.ID, 
            func(msg liveview.BroadcastMessage) {
                if newMsg, ok := msg.Payload.(Message); ok {
                    c.messages = append(c.messages, newMsg)
                    ctx.Assign("messages", c.messages)
                    ctx.Socket.PushEvent("phx:update", map[string]interface{}{})
                }
            })
    }
    
    return nil
}

func (c *Chat) HandleEvent(ctx *liveview.Context, event string, payload map[string]interface{}) error {
    if event == "send_message" {
        if text, ok := payload["message"].(string); ok && text != "" {
            msg := Message{
                User: c.username,
                Text: text,
                Time: time.Now(),
            }
            
            // 全クライアントにブロードキャスト
            if broadcaster := ctx.GetBroadcaster(); broadcaster != nil {
                broadcaster.Broadcast("chat:room", "new_message", msg)
            }
        }
    }
    return nil
}
```

### 3. サーバーの設定

```go
func main() {
    // WebSocketサーバー
    wsServer := socket.NewServer()
    
    // PubSub（ローカルまたはRedis）
    pubsub := liveview.NewLocalPubSub()
    // または: pubsub := NewRedisPubSub("localhost:6379")
    
    // ブロードキャスター
    broadcaster := liveview.NewBroadcaster(pubsub)
    
    // マネージャー
    manager := liveview.NewManager(wsServer)
    manager.SetBroadcaster(broadcaster)
    
    // LiveViewの登録
    manager.Register("chat", func() liveview.LiveView {
        return &Chat{}
    })
    
    // ハンドラー
    handler := liveview.NewHandler(manager, wsServer, liveview.HandlerOptions{
        Template: customTemplate(),
    })
    
    // 静的ファイル
    http.Handle("/liveview.js", http.FileServer(http.Dir("./js/dist")))
    http.Handle("/", handler)
    
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## HTMLテンプレート

### JavaScriptクライアントの読み込み

```html
<!DOCTYPE html>
<html>
<head>
    <title>My LiveView App</title>
    <meta name="csrf-token" content="{{.CSRFToken}}">
</head>
<body>
    <div id="live-view-root"></div>
    
    <script src="/liveview.js"></script>
    <script>
        const liveSocket = new LiveSocket('/live', {
            params: { _csrf_token: document.querySelector('meta[name="csrf-token"]').content }
        });
        liveSocket.connect();
    </script>
</body>
</html>
```

## イベントハンドリング

### クライアント側（HTML）

```html
<!-- クリックイベント -->
<button phx-click="inc">Increment</button>

<!-- 値を渡す -->
<button phx-click="set" phx-value="10">Set to 10</button>

<!-- フォーム入力 -->
<input type="text" phx-change="validate" name="email" />

<!-- フォーム送信 -->
<form phx-submit="save">
    <input type="text" name="name" />
    <button type="submit">Save</button>
</form>
```

### サーバー側（Go）

```go
func (c *MyLiveView) HandleEvent(ctx *liveview.Context, event string, payload map[string]interface{}) error {
    switch event {
    case "inc":
        // ボタンクリック
        c.count++
        
    case "set":
        // 値付きイベント
        if val, ok := payload["value"].(string); ok {
            if n, err := strconv.Atoi(val); err == nil {
                c.count = n
            }
        }
        
    case "validate":
        // フォーム入力変更
        if email, ok := payload["email"].(string); ok {
            // バリデーション
        }
        
    case "save":
        // フォーム送信
        if name, ok := payload["name"].(string); ok {
            // 保存処理
        }
    }
    
    ctx.Assign("count", c.count)
    return nil
}
```

## テスト

### LiveViewのテスト

```go
func TestMyLiveView(t *testing.T) {
    // セットアップ
    wsServer := socket.NewServer()
    manager := liveview.NewManager(wsServer)
    
    manager.Register("test", func() liveview.LiveView {
        return &MyLiveView{}
    })
    
    // HTTPテスト
    handler := liveview.NewHandler(manager, wsServer, liveview.HandlerOptions{
        Template: `<html><body></body></html>`,
    })
    
    req := httptest.NewRequest("GET", "/", nil)
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)
    
    if rec.Code != http.StatusOK {
        t.Errorf("Expected 200, got %d", rec.Code)
    }
}
```

### PubSubのテスト

```go
func TestPubSub(t *testing.T) {
    pubsub := liveview.NewLocalPubSub()
    
    received := make(chan interface{}, 1)
    
    // 購読
    pubsub.Subscribe("test:topic", func(msg interface{}) {
        received <- msg
    })
    
    // 公開
    pubsub.Publish("test:topic", "hello")
    
    // 確認
    select {
    case msg := <-received:
        if msg != "hello" {
            t.Errorf("Expected hello, got %v", msg)
        }
    case <-time.After(time.Second):
        t.Error("Timeout")
    }
}
```

## 高度な機能

### ファイルアップロード

```go
func (c *MyLiveView) Mount(ctx *liveview.Context, params url.Values) error {
    ctx.Socket.AllowUpload("avatar", liveview.UploadConfig{
        Accept:      []string{".jpg", ".png"},
        MaxEntries:  1,
        MaxFileSize: 1024 * 1024, // 1MB
    })
    return nil
}
```

### フラッシュメッセージ

```go
func (c *MyLiveView) HandleEvent(ctx *liveview.Context, event string, payload map[string]interface{}) error {
    if event == "save" {
        // 保存処理
        ctx.Socket.PutFlash("info", "Saved successfully!")
    }
    return nil
}
```

### JavaScriptイベントの送信

```go
func (c *MyLiveView) HandleEvent(ctx *liveview.Context, event string, payload map[string]interface{}) error {
    // JavaScriptクライアントにイベントを送信
    ctx.Socket.PushEvent("focus", map[string]interface{}{
        "selector": "#search-input",
    })
    return nil
}
```

## デプロイ

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/js/dist ./js/dist
CMD ["./server"]
```

### ビルド

```bash
# Goアプリケーションのビルド
go build -o server ./cmd/main.go

# JavaScriptクライアントのビルド
cd js && npm install && npm run build

# 実行
./server
```

## トラブルシューティング

### WebSocketが接続できない

```bash
# ファイアウォールの確認
sudo lsof -i :8080

# 正しいポートでリッスンしているか確認
netstat -tlnp | grep 8080
```

### メモリリーク

```go
// 必ずUnsubscribeを呼ぶ
defer broadcaster.Unsubscribe("topic", subscriberID)
```

### パフォーマンス

```go
// 大きなリストはページネーションを使用
ctx.Assign("items", items[:100])
```

## 参考リンク

- [Phoenix LiveView ドキュメント](https://hexdocs.pm/phoenix_live_view/Phoenix.LiveView.html)
- [gorilla/websocket](https://github.com/gorilla/websocket)
- [a-h/templ](https://github.com/a-h/templ)
- [morphdom](https://github.com/patrick-steele-idem/morphdom)
