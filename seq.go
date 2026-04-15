package nrup

import (
	"sync"
	"time"
)

// SeqTracker 序列号跟踪+丢包检测
type SeqTracker struct {
	mu        sync.Mutex
	nextSeq   uint32
	ackMap    map[uint32]time.Time
	received  map[uint32]bool
	rttSum    time.Duration
	rttCount  int
	lostCount int
	sentCount int
	// v1.4.3: jitter跟踪
	lastRTT   time.Duration
	jitter    time.Duration // EWMA RTT抖动
}

func NewSeqTracker() *SeqTracker {
	return &SeqTracker{
		ackMap:   make(map[uint32]time.Time),
		received: make(map[uint32]bool),
	}
}

// OnSend 记录发送
func (s *SeqTracker) OnSend(seq uint32) {
	s.mu.Lock()
	s.ackMap[seq] = time.Now()
	s.sentCount++
	s.mu.Unlock()
}

// OnRecvACK 收到确认
func (s *SeqTracker) OnRecvACK(seq uint32) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	sendTime, ok := s.ackMap[seq]
	if !ok {
		return 0
	}
	rtt := time.Since(sendTime)
	// v1.4.3: 计算jitter (EWMA)
	if s.lastRTT > 0 {
		diff := rtt - s.lastRTT
		if diff < 0 { diff = -diff }
		s.jitter = s.jitter*7/8 + diff/8 // EWMA α=0.125
	}
	s.lastRTT = rtt
	s.rttSum += rtt
	s.rttCount++
	delete(s.ackMap, seq)
	return rtt
}

// CheckLoss 检测丢包（超过2*RTT未确认=丢了）
func (s *SeqTracker) CheckLoss() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	avgRTT := s.AvgRTT()
	if avgRTT == 0 {
		avgRTT = 200 * time.Millisecond
	}
	timeout := avgRTT * 3

	lost := 0
	now := time.Now()
	for seq, sendTime := range s.ackMap {
		if now.Sub(sendTime) > timeout {
			delete(s.ackMap, seq)
			lost++
		}
	}
	s.lostCount += lost
	return lost
}

// AvgRTT 平均RTT
func (s *SeqTracker) AvgRTT() time.Duration {
	if s.rttCount == 0 {
		return 0
	}
	return s.rttSum / time.Duration(s.rttCount)
}

// LossRate 丢包率
func (s *SeqTracker) LossRate() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sentCount == 0 {
		return 0
	}
	return float64(s.lostCount) / float64(s.sentCount)
}

// Stats 统计
func (s *SeqTracker) Stats() (sent, lost int, rtt time.Duration, lossRate float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sent = s.sentCount
	lost = s.lostCount
	rtt = s.AvgRTT()
	if sent > 0 {
		lossRate = float64(lost) / float64(sent)
	}
	return
}

// Jitter 返回当前RTT抖动
func (s *SeqTracker) Jitter() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.jitter
}
