package nrup

import (
	"log"

	"github.com/nyarime/gofec/ldpc"
)

// NewFECByType 根据类型创建FEC编解码器
func NewFECByType(fecType FECType, data, parity int) *FECCodec {
	switch fecType {
	case FECTypeLDPC:
		log.Printf("[FEC] 使用LDPC (GoFEC PEG, %d数据+%d校验)", data, parity)
		return newLDPCWrapped(data, parity)
	case FECTypeRaptorQ:
		log.Printf("[FEC] 使用LT喷泉码 (%d源块)", data)
		return NewFECCodec(data, parity) // LT暂用RS包装
	default: // RS
		log.Printf("[FEC] 使用Reed-Solomon (%d+%d)", data, parity)
		return NewFECCodec(data, parity)
	}
}

// newLDPCWrapped 包装GoFEC LDPC为FECCodec兼容接口
func newLDPCWrapped(data, parity int) *FECCodec {
	// LDPC编解码器
	codec := ldpc.New(data, parity, 0.3)
	
	// 包装成FECCodec(复用现有帧格式和序号)
	fec := NewFECCodec(data, parity)
	fec.ldpcCodec = codec
	fec.useLDPC = true
	return fec
}
