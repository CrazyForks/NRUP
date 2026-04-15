package nrup

import (
	"context"
	"net"
)

// Dialer 传输路径抽象接口
// NRUP自己实现UDP路径，外部(Lite)注入TCP路径(NRTP)
// 解决NRUP↔NRTP循环依赖
type Dialer interface {
	// DialUDP 使用NRUP的UDP路径
	DialUDP(ctx context.Context, addr string) (net.Conn, error)
	// DialTCP 使用外部注入的TCP路径
	DialTCP(ctx context.Context, addr string) (net.Conn, error)
}

// TransportStats 传输状态(TUI显示)
type TransportStats struct {
	Mode          string  // "UDP" / "TCP" / "UDP+TCP"
	UDPAvailable  bool
	Sessions      int     // 活跃session数
	RTT           int64   // ms
	LossRate      float64
	FECParity     int
	FECEfficiency float64
	Jitter        int64   // ms
	MTU           int
	BytesSent     int64
	BytesRecv     int64
}

// SmartDialer 智能传输选择器接口
// v1.5.0: 定义在NRUP层，由Lite实现并注入
type SmartDialer interface {
	Dialer
	// Stats 返回当前传输状态
	Stats() TransportStats
	// SetUDPAvailable 设置UDP可用性
	SetUDPAvailable(available bool)
	// Close 关闭所有连接
	Close() error
}
