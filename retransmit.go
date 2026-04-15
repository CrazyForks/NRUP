package nrup

import (
	"sync"
	"time"
)

// RetransmitQueue 选择性重传队列
type RetransmitQueue struct {
	mu       sync.Mutex
	pending  map[uint32]*retransmitEntry
	maxRetry int
	// v1.4.3: 预测性重传
	jitter   time.Duration // RTT抖动
}

type retransmitEntry struct {
	frames   [][]byte  // 原始FEC帧
	sentAt   time.Time
	retries  int
	rto      time.Duration
}

func NewRetransmitQueue() *RetransmitQueue {
	return &RetransmitQueue{
		pending:  make(map[uint32]*retransmitEntry),
		maxRetry: 3,
	}
}

// Add 记录已发送的帧组（用于可能的重传）
func (rq *RetransmitQueue) Add(seq uint32, frames [][]byte, rto time.Duration) {
	rq.mu.Lock()
	rq.pending[seq] = &retransmitEntry{
		frames:  frames,
		sentAt:  time.Now(),
		rto:     rto,
	}
	rq.mu.Unlock()
}

// ACK 确认收到，从队列移除
func (rq *RetransmitQueue) ACK(seq uint32) {
	rq.mu.Lock()
	delete(rq.pending, seq)
	rq.mu.Unlock()
}

// GetExpired 获取需要重传的帧(含预测性重传)
func (rq *RetransmitQueue) GetExpired() []retransmitResult {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	var results []retransmitResult
	now := time.Now()

	for seq, entry := range rq.pending {
		elapsed := now.Sub(entry.sentAt)
		// v1.4.3: 预测性重传——当RTT抖动大时，提前0.8*RTO就重传
		threshold := entry.rto
		if rq.jitter > entry.rto/4 {
			threshold = entry.rto * 4 / 5 // 提前20%
		}
		if elapsed > threshold {
			if entry.retries >= rq.maxRetry {
				delete(rq.pending, seq)
				continue
			}
			entry.retries++
			entry.sentAt = now
			results = append(results, retransmitResult{
				Seq:    seq,
				Frames: entry.frames,
			})
		}
	}
	return results
}

// UpdateJitter 更新RTT抖动
func (rq *RetransmitQueue) UpdateJitter(jitter time.Duration) {
	rq.mu.Lock()
	rq.jitter = jitter
	rq.mu.Unlock()
}

type retransmitResult struct {
	Seq    uint32
	Frames [][]byte
}

// Size 队列大小
func (rq *RetransmitQueue) Size() int {
	rq.mu.Lock()
	defer rq.mu.Unlock()
	return len(rq.pending)
}
