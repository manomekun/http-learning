package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

type TCPHeader []byte

func (h TCPHeader) SourcePort() uint16 {
	return binary.BigEndian.Uint16(h[0:2])
}

func (h TCPHeader) DestinationPort() uint16 {
	return binary.BigEndian.Uint16(h[2:4])
}

func (h TCPHeader) SequenceNumber() uint32 {
	return binary.BigEndian.Uint32(h[4:8])
}

func (h TCPHeader) AcknowledgementNumber() uint32 {
	return binary.BigEndian.Uint32(h[8:12])
}

func (h TCPHeader) DataOffset() uint8 {
	return (h[12] >> 4) & 0x0F
}

func (h TCPHeader) URG() bool {
	return (h[13] & 0x20) != 0
}

func (h TCPHeader) ACK() bool {
	return (h[13] & 0x10) != 0
}

func (h TCPHeader) PSH() bool {
	return (h[13] & 0x08) != 0
}

func (h TCPHeader) RST() bool {
	return (h[13] & 0x04) != 0
}

func (h TCPHeader) SYN() bool {
	return (h[13] & 0x02) != 0
}

func (h TCPHeader) FIN() bool {
	return (h[13] & 0x01) != 0
}

func (h TCPHeader) Window() uint16 {
	return binary.BigEndian.Uint16(h[14:16])
}

func (h TCPHeader) Checksum() uint16 {
	return binary.BigEndian.Uint16(h[16:18])
}

func (h TCPHeader) UrgentPointer() uint16 {
	return binary.BigEndian.Uint16(h[18:20])
}

func NewTCPHeader(srcPort, dstPort uint16, seqNum uint32, flags byte) TCPHeader {
	h := make(TCPHeader, 20)

	binary.BigEndian.PutUint16(h[0:2], srcPort)
	binary.BigEndian.PutUint16(h[2:4], dstPort)
	binary.BigEndian.PutUint32(h[4:8], seqNum)
	binary.BigEndian.PutUint32(h[8:12], 0)
	h[12] = 0x50
	h[13] = flags
	binary.BigEndian.PutUint16(h[14:16], 65535)

	return h
}

func (h TCPHeader) CalcCheckSum(srcIP, dstIP net.IP) uint16 {
	buf := make([]byte, 32)

	// 疑似ヘッダー
	copy(buf[0:4], srcIP.To4())
	copy(buf[4:8], dstIP.To4())
	buf[8] = 0x00
	buf[9] = 0x06
	binary.BigEndian.PutUint16(buf[10:12], uint16(len(h)))

	// TCP ヘッダーをコピー
	copy(buf[12:], h)

	// CheckSum フィールドを0にして計算
	buf[12+16] = 0x00
	buf[12+17] = 0x00

	return calcCheckSum(buf)
}

func calcCheckSum(data []byte) uint16 {
	var sum uint32

	// 16ビット単位で加算
	for i := 0; i < len(data)-1; i += 2 {
		sum += uint32(binary.BigEndian.Uint16(data[i:]))
	}

	// 奇数バイトの場合、最後の1バイトを処理
	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1]) << 8
	}

	// キャリーを足し戻す
	for sum > 0xFFFF {
		sum = (sum & 0xFFFF) + (sum >> 16)
	}

	return ^uint16(sum)
}

func main() {
	dstIP := net.ParseIP("142.250.189.206").To4()
	srcIP := net.ParseIP("192.168.215.3").To4()

	srcPort := uint16(54321)
	dstPort := uint16(80)

	h := NewTCPHeader(srcPort, dstPort, 1000, 0x02)

	checksum := h.CalcCheckSum(srcIP, dstIP)
	binary.BigEndian.PutUint16(h[16:18], checksum)

	conn, err := net.Dial("ip4:tcp", dstIP.String())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	n, err := conn.Write(h)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Sent %d bytes SYN packet to %s:%d\n", n, dstIP, dstPort)

	response := TCPHeader{}
	for {
		// ここから SYN-ACK を待つ
		// recieve
		buf := make([]byte, 1024)
		n, err = conn.Read(buf)
		if err != nil {
			panic(err)
		}

		// デバッグ: 受信した生データを表示
		fmt.Printf("Received %d bytes\n", n)
		fmt.Printf("Raw data: %x\n", buf[:n])
		fmt.Printf("Byte 13 (flags): %08b\n", buf[13])

		ipHeaderLen := int((buf[0] & 0x0F) * 4)
		response = TCPHeader(buf[ipHeaderLen:n])

		// 相手からの SYN-ACK かチェック
		if response.SourcePort() == dstPort && response.DestinationPort() == srcPort {
			fmt.Printf("Received SYN-ACK!\n")
			fmt.Printf("  Seq: %d\n", response.SequenceNumber())
			fmt.Printf("  Ack: %d\n", response.AcknowledgementNumber())
			fmt.Printf("  SYN: %v, ACK: %v\n", response.SYN(), response.ACK())
			break
		}

		fmt.Printf("Skipping packet (SrcPort=%d, DstPort=%d)\n",
			response.SourcePort(), response.DestinationPort())

	}

	// ここから ACK
	ackPacket := NewTCPHeader(srcPort, dstPort, 1001, 0x10)
	// ACK 番号をセット
	binary.BigEndian.PutUint32(ackPacket[8:12], response.SequenceNumber()+1)
	// チェックサム計算
	checkSum := ackPacket.CalcCheckSum(srcIP, dstIP)
	binary.BigEndian.PutUint16(ackPacket[16:18], checkSum)

	conn.Write(ackPacket)
	fmt.Println("Sent ACK! 3-way handshake complete!")
}
