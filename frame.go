package nrup

import (
	"encoding/binary"
)

// 帧类型
const (
	FrameData  = 0x01 // 数据帧
	FrameACK   = 0x02 // 确认帧
	FramePing  = 0x03 // 心跳帧
	FrameClose = 0x05
	FrameBatchACK = 0x06 // 批量确认帧 // 关闭帧
)

// DataFrame 数据帧
// [1B type=0x01][4B seq][1B fecIdx][1B fecTotal][2B dataLen][shard]



// ACKFrame 确认帧
// [1B type=0x02][4B ackSeq][4B bitmap]
// bitmap: 32个seq的确认状态
type ACKFrame struct {
	Type    byte
	AckSeq  uint32
	Bitmap  uint32 // 从ackSeq开始，每bit=一个seq的确认状态
}

func EncodeACKFrame(ackSeq uint32, bitmap uint32) []byte {
	frame := make([]byte, 9)
	frame[0] = FrameACK
	binary.BigEndian.PutUint32(frame[1:5], ackSeq)
	binary.BigEndian.PutUint32(frame[5:9], bitmap)
	return frame
}

func DecodeACKFrame(data []byte) *ACKFrame {
	if len(data) < 9 || data[0] != FrameACK {
		return nil
	}
	return &ACKFrame{
		Type:   data[0],
		AckSeq: binary.BigEndian.Uint32(data[1:5]),
		Bitmap: binary.BigEndian.Uint32(data[5:9]),
	}
}

// PingFrame 心跳帧
// [1B type=0x03][8B timestamp]

func EncodePingFrame(ts uint64) []byte {
	frame := make([]byte, 9)
	frame[0] = FramePing
	binary.BigEndian.PutUint64(frame[1:9], ts)
	return frame
}


// EncodeBatchACK 批量ACK帧: [1B type=0x06][1B count][N * (4B seq + 4B bitmap)]
func EncodeBatchACK(acks []ACKFrame) []byte {
	if len(acks) == 0 { return nil }
	if len(acks) == 1 { return EncodeACKFrame(acks[0].AckSeq, acks[0].Bitmap) }
	frame := make([]byte, 2+8*len(acks))
	frame[0] = FrameBatchACK
	frame[1] = byte(len(acks))
	for i, ack := range acks {
		binary.BigEndian.PutUint32(frame[2+8*i:], ack.AckSeq)
		binary.BigEndian.PutUint32(frame[6+8*i:], ack.Bitmap)
	}
	return frame
}

// DecodeBatchACK 解码批量ACK
func DecodeBatchACK(data []byte) []ACKFrame {
	if len(data) < 2 || data[0] != FrameBatchACK { return nil }
	count := int(data[1])
	if len(data) < 2+8*count { return nil }
	acks := make([]ACKFrame, count)
	for i := 0; i < count; i++ {
		acks[i] = ACKFrame{
			Type:   FrameACK,
			AckSeq: binary.BigEndian.Uint32(data[2+8*i:]),
			Bitmap: binary.BigEndian.Uint32(data[6+8*i:]),
		}
	}
	return acks
}
