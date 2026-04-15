package nrup

import (
	"fmt"
	"github.com/nyarime/gofec/raptorq"
)

// RaptorQCodec GoFEC RaptorQ编解码器
// 纯Go + AVX2/NEON SIMD, 喷泉码无限修复符号
type RaptorQCodec struct {
	codec      *raptorq.Codec
	numSource  int
	symbolSize int
}

// NewRaptorQCodec 创建RaptorQ编解码器
func NewRaptorQCodec(numSource, symbolSize int) *RaptorQCodec {
	return &RaptorQCodec{
		codec:      raptorq.New(numSource, symbolSize),
		numSource:  numSource,
		symbolSize: symbolSize,
	}
}

// Encode RaptorQ编码: 源数据→源符号+修复符号
func (c *RaptorQCodec) Encode(data []byte, numRepair int) []raptorq.Symbol {
	return c.codec.Encode(data, numRepair)
}

// Decode RaptorQ解码: 收到足够符号即可恢复
func (c *RaptorQCodec) Decode(symbols []raptorq.Symbol, dataLen int) ([]byte, error) {
	return c.codec.Decode(symbols, dataLen)
}

// Type 返回编码类型
func (c *RaptorQCodec) Type() string {
	return string(FECTypeRaptorQ)
}

// String 格式化输出
func (c *RaptorQCodec) String() string {
	return fmt.Sprintf("RaptorQ(K=%d,T=%d)", c.numSource, c.symbolSize)
}
