package haruka

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// DefaultWebSocketUpgrader is used by UpgradeToWebSocket. Replace or
// adjust CheckSameOrigin for stricter origin policies.
var DefaultWebSocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// UpgradeToWebSocket upgrades the current request to a WebSocket
// connection using DefaultWebSocketUpgrader.
func (c *Context) UpgradeToWebSocket() (*websocket.Conn, error) {
	return DefaultWebSocketUpgrader.Upgrade(c.Writer, c.Request, nil)
}

// UpgradeToWebSocketWith upgrades the request with a custom upgrader.
func (c *Context) UpgradeToWebSocketWith(upgrader *websocket.Upgrader) (*websocket.Conn, error) {
	return upgrader.Upgrade(c.Writer, c.Request, nil)
}
