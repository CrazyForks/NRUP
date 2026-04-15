package nrup

import (
	"context"
	"net"
	"sync"
)

// Client NRUP客户端(v1.5.0独立传输库入口)
type Client struct {
	cfg    *Config
	dialer SmartDialer
	mu     sync.Mutex
	conn   *Conn // 当前UDP连接
}

// NewClient 创建NRUP客户端
// dialer由Lite注入(含NRTP TCP路径)
func NewClient(cfg *Config, dialer SmartDialer) *Client {
	return &Client{
		cfg:    cfg,
		dialer: dialer,
	}
}

// DialUDP 通过NRUP建立UDP连接
func (c *Client) DialUDP(addr string) (*Conn, error) {
	return Dial(addr, c.cfg)
}

// DialSmart 智能选择UDP/TCP
func (c *Client) DialSmart(ctx context.Context, addr string) (net.Conn, error) {
	if c.dialer == nil {
		// 没有注入dialer，只用UDP
		return Dial(addr, c.cfg)
	}
	return c.dialer.DialUDP(ctx, addr)
}

// Stats 返回传输状态
func (c *Client) Stats() TransportStats {
	if c.dialer != nil {
		return c.dialer.Stats()
	}
	return TransportStats{Mode: "UDP", UDPAvailable: true}
}
