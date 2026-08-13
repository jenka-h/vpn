// Berisi logika sederhana dari socket UDP.
package internal

import "net"

type UDPTransport struct {
	conn *net.UDPConn
	peer *net.UDPAddr
}

func NewUDPTransport(
	listenAddr string,
	peerAddr string,
) (*UDPTransport, error) {
	listen, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return nil, err
	}
	peer, err := net.ResolveUDPAddr("udp", peerAddr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", listen)
	if err != nil {
		return nil, err
	}

	return &UDPTransport{conn: conn, peer: peer}, nil
}

func (u *UDPTransport) Send(data []byte) error {
	_, err := u.conn.WriteToUDP(data, u.peer)
	return err
}

func (u *UDPTransport) Receive(
	buf []byte,
) (int, *net.UDPAddr, error) {
	return u.conn.ReadFromUDP(buf)
}

func (u *UDPTransport) IsPeer(addr *net.UDPAddr) bool {
	if addr == nil {
		return false
	}
	return addr.IP.Equal(u.peer.IP) && addr.Port == u.peer.Port
}

func (u *UDPTransport) Close() error {
	return u.conn.Close()
}
