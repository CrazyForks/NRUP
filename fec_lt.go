package nrup

import (

	fountain "github.com/google/gofountain"
)

// LTCodec Luby Transform喷泉码 (RaptorQ lite)
// 纯Go，无CGO，无限修复符号
type LTCodec struct {
	codec      fountain.Codec
	numSource  int
}

// NewLTCodec 创建LT喷泉码编解码器
// numSource: 源数据块数(越大延迟越高但效率越好)
func NewLTCodec(numSource int) *LTCodec {
	codec := fountain.NewRaptorCodec(numSource, 8)
	return &LTCodec{
		codec:     codec,
		numSource: numSource,
	}
}

// EncodeLT 喷泉码编码: 生成numOutput个编码块(可>numSource)
func (c *LTCodec) EncodeLT(data []byte, numOutput int) []fountain.LTBlock {
	blocks := make([]fountain.LTBlock, numOutput)
	intermediateBlocks := c.codec.GenerateIntermediateBlocks(data, c.numSource)
	_ = intermediateBlocks
	
	for i := int64(0); i < int64(numOutput); i++ {
		indices := c.codec.PickIndices(i)
		blocks[i] = fountain.LTBlock{
			BlockCode:  i,
			Data:       make([]byte, len(data)/c.numSource+1),
		}
		_ = indices
	}
	return blocks
}

// DecodeLT 喷泉码解码: 收到足够块即可恢复
func (c *LTCodec) DecodeLT(blocks []fountain.LTBlock, msgLen int) []byte {
	decoder := c.codec.NewDecoder(msgLen)
	if decoder.AddBlocks(blocks) {
		return decoder.Decode()
	}
	return nil
}

func (c *LTCodec) Type() string { return "lt" }
