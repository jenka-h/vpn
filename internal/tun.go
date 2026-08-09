package internal

import "os"

type TUN struct {
	file *os.File
	Name string
}

func NewTUN(name string, cidr string, mtu int) (*TUN, error)

func createTUNDevice(name string) (*os.File, error)

func configureTUN(name string, cidr string, mtu int) error

func (t *TUN) ReadPacket(buf []byte) (int, error)

func (t *TUN) WritePacket(packet []byte) error

func (t *TUN) Close() error
