package arping

import (
	"bytes"
	"encoding/binary"
	"net"
)

const (
	requestOper  = 1
	responseOper = 2
)

type arpDatagram struct {
	htype uint16 // Hardware Type
	ptype uint16 // Protocol Type
	hlen  uint8  // Hardware address Length
	plen  uint8  // Protocol address length
	oper  uint16 // Operation 1->request, 2->response
	sha   []byte // Sender hardware address, length from Hlen
	spa   []byte // Sender protocol address, length from Plen
	tha   []byte // Target hardware address, length from Hlen
	tpa   []byte // Target protocol address, length from Plen
}

func newArpRequest(
	srcMac net.HardwareAddr,
	srcIP net.IP,
	dstMac net.HardwareAddr,
	dstIP net.IP) arpDatagram {
	return arpDatagram{
		htype: uint16(1),
		ptype: uint16(0x0800),
		hlen:  uint8(6),
		plen:  uint8(4),
		oper:  uint16(requestOper),
		sha:   srcMac,
		spa:   srcIP.To4(),
		tha:   dstMac,
		tpa:   dstIP.To4()}
}

func (datagram arpDatagram) Marshal() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, datagram.htype)
	binary.Write(buf, binary.BigEndian, datagram.ptype)
	binary.Write(buf, binary.BigEndian, datagram.hlen)
	binary.Write(buf, binary.BigEndian, datagram.plen)
	binary.Write(buf, binary.BigEndian, datagram.oper)
	buf.Write(datagram.sha)
	buf.Write(datagram.spa)
	buf.Write(datagram.tha)
	buf.Write(datagram.tpa)

	return buf.Bytes()
}

func (datagram arpDatagram) MarshalWithEthernetHeader() []byte {
	// ethernet frame header
	var ethernetHeader []byte
	ethernetHeader = append(ethernetHeader, datagram.tha...)
	ethernetHeader = append(ethernetHeader, datagram.sha...)
	ethernetHeader = append(ethernetHeader, []byte{0x08, 0x06}...) // arp

	return append(ethernetHeader, datagram.Marshal()...)
}

func (datagram arpDatagram) SenderIP() net.IP {
	return net.IP(datagram.spa)
}
func (datagram arpDatagram) SenderMac() net.HardwareAddr {
	return net.HardwareAddr(datagram.sha)
}

func (datagram arpDatagram) IsResponseOf(request arpDatagram) bool {
	return datagram.oper == responseOper && bytes.Equal(request.tpa, datagram.spa) &&
		bytes.Equal(request.spa, datagram.tpa)
}

func (datagram arpDatagram) IsResponseOfTarget(request arpDatagram) bool {
	return datagram.oper == responseOper && bytes.Equal(request.tpa, datagram.spa)
}

// IsDuplicateRequestOf case: Cisco switch will re-send an ARP request when it received a DAD-ARP request
func (datagram arpDatagram) IsDuplicateRequestOf(request arpDatagram, ignoreCheck bool) bool {
	return datagram.oper == requestOper && bytes.Equal(request.tpa, datagram.spa) &&
		(ignoreCheck || bytes.Equal(request.tpa, datagram.tpa))
}

// IsResponseOfDADRequest case: H3C switch will reply DAD ARP request when it received a DAD-ARP request
func (datagram arpDatagram) IsResponseOfDADRequest(request arpDatagram, ignoreCheck bool) bool {
	return datagram.oper == responseOper && bytes.Equal(request.tpa, datagram.spa) &&
		(ignoreCheck || bytes.Equal(request.tpa, datagram.tpa))
}

func parseArpDatagram(buffer []byte) arpDatagram {
	var datagram arpDatagram

	b := bytes.NewBuffer(buffer)
	binary.Read(b, binary.BigEndian, &datagram.htype)
	binary.Read(b, binary.BigEndian, &datagram.ptype)
	binary.Read(b, binary.BigEndian, &datagram.hlen)
	binary.Read(b, binary.BigEndian, &datagram.plen)
	binary.Read(b, binary.BigEndian, &datagram.oper)

	haLen := int(datagram.hlen)
	paLen := int(datagram.plen)
	datagram.sha = b.Next(haLen)
	datagram.spa = b.Next(paLen)
	datagram.tha = b.Next(haLen)
	datagram.tpa = b.Next(paLen)

	return datagram
}
