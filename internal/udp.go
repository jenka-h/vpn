// Berisi logika sederhana dari socket UDP
package internal

import "net"

type UDPTransport struct {
	conn *net.UDPConn
	peer *net.UDPAddr
}

func NewUDPTransport(
	listenAddr string,
	peerAddr string,
) (*UDPTransport, error)

func (u *UDPTransport) Send(data []byte) error

func (u *UDPTransport) Receive(
	buf []byte,
) (int, *net.UDPAddr, error)

func (u *UDPTransport) Close() error
