// Protokol packet VPN.
package internal

import (
	"encoding/binary"
	"errors"
)

const ProtocolVersion uint8 = 1

// HeaderSize is authenticated as AEAD additional data:
// version:  1 byte
// sequence: 8 bytes
const HeaderSize = 9
const NonceSize = 12
const MinimumPacketSize = HeaderSize + NonceSize + 16

// struktur sederhana
type Packet struct {
	Version    uint8
	Sequence   uint64
	Nonce      []byte
	Ciphertext []byte
}

func EncodeAuthenticatedHeader(
	version uint8,
	sequence uint64,
) []byte {
	header := make([]byte, HeaderSize)
	header[0] = version
	binary.BigEndian.PutUint64(header[1:], sequence)
	return header
}

func (p *Packet) AuthenticatedData() []byte {
	return EncodeAuthenticatedHeader(p.Version, p.Sequence)
}

func EncodePacket(packet *Packet) ([]byte, error) {
	if err := ValidatePacket(packet); err != nil {
		return nil, err
	}

	data := make([]byte, 0, HeaderSize+len(packet.Nonce)+len(packet.Ciphertext))
	data = append(data, packet.AuthenticatedData()...)
	data = append(data, packet.Nonce...)
	data = append(data, packet.Ciphertext...)
	return data, nil
}

func DecodePacket(data []byte) (*Packet, error) {
	if len(data) < MinimumPacketSize {
		return nil, errors.New("vpn packet is too short")
	}

	packet := &Packet{
		Version:    data[0],
		Sequence:   binary.BigEndian.Uint64(data[1:HeaderSize]),
		Nonce:      append([]byte(nil), data[HeaderSize:HeaderSize+NonceSize]...),
		Ciphertext: append([]byte(nil), data[HeaderSize+NonceSize:]...),
	}
	if err := ValidatePacket(packet); err != nil {
		return nil, err
	}
	return packet, nil
}

func ValidatePacket(packet *Packet) error {
	if packet == nil {
		return errors.New("packet is nil")
	}
	if packet.Version != ProtocolVersion {
		return errors.New("unsupported protocol version")
	}
	if len(packet.Nonce) != NonceSize {
		return errors.New("invalid nonce size")
	}
	if len(packet.Ciphertext) < 16 {
		return errors.New("ciphertext is too short")
	}
	return nil
}
