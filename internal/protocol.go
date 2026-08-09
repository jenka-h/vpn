// Protokol VPN
package internal

const ProtocolVersion uint8 = 1

// HeaderSize is the authenticated header size:
// version:  1 byte
// sequence: 8 bytes
const HeaderSize = 9

type Packet struct {
	Version    uint8
	Sequence   uint64
	Nonce      []byte
	Ciphertext []byte
}

// EncodeAuthenticatedHeader returns the protocol header that must be passed
// as AEAD additional authenticated data when encrypting/decrypting a packet.
// This binds Version and Sequence to Ciphertext so they cannot be modified
// without authentication failure.
func EncodeAuthenticatedHeader(
	version uint8,
	sequence uint64,
) []byte

func (p *Packet) AuthenticatedData() []byte

func EncodePacket(packet *Packet) ([]byte, error)

func DecodePacket(data []byte) (*Packet, error)

func ValidatePacket(packet *Packet) error
