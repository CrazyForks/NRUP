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
		log.Printf("[FEC] 使用RaptorQ (GoFEC, %d源块+%d修复)", data, parity)
		return newRaptorQWrapped(data, parity)
	default:
		log.Printf("[FEC] 使用Reed-Solomon (%d+%d)", data, parity)
		return NewFECCodec(data, parity)
	}
}

// newLDPCWrapped 包装GoFEC LDPC
func newLDPCWrapped(data, parity int) *FECCodec {
	codec := ldpc.New(data, parity, 0.3)
	fec := NewFECCodec(data, parity)
	fec.ldpcCodec = codec
	fec.useLDPC = true
	return fec
}

// newRaptorQWrapped 包装GoFEC RaptorQ
func newRaptorQWrapped(data, parity int) *FECCodec {
	fec := NewFECCodec(data, parity)
	fec.raptorqCodec = NewRaptorQCodec(data, 64) // 64字节/符号(UDP包大小友好)
	fec.useRaptorQ = true
	return fec
}
