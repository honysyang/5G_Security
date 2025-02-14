package packetparser

import (
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// ParsePacket 解析网络数据包
func ParsePacket(data []byte) (*gopacket.Packet, error) {
	// 使用 gopacket 解析数据包
	packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.DecodeOptions{NoCopy: true})
	// 检查是否有解码错误
	if packet.ErrorLayer() != nil {
		return nil, fmt.Errorf("failed to decode packet: %v", packet.ErrorLayer().Error())
	}
	return packet, nil
}
