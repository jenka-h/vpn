package internal

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

// linux constant, pastikan kalau program dijalankan di wsl atau linux env
const (
	cloneDevice = "/dev/net/tun"
	iffTUN      = 0x0001
	iffNoPI     = 0x1000
	tunSetIFF   = 0x400454ca
	ifNameSize  = 16
)

type TUN struct {
	file *os.File
	Name string
}

func NewTUN(name string, cidr string, mtu int) (*TUN, error) {
	file, err := createTUNDevice(name)
	if err != nil {
		return nil, err
	}
	if err := configureTUN(name, cidr, mtu); err != nil {
		_ = file.Close()
		return nil, err
	}
	return &TUN{file: file, Name: name}, nil
}

func createTUNDevice(name string) (*os.File, error) {
	file, err := os.OpenFile(cloneDevice, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	var req [ifNameSize + 64]byte
	copy(req[:ifNameSize], name)
	flags := uint16(iffTUN | iffNoPI)
	req[ifNameSize] = byte(flags)
	req[ifNameSize+1] = byte(flags >> 8)

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), uintptr(tunSetIFF), uintptr(unsafe.Pointer(&req[0])))
	if errno != 0 {
		_ = file.Close()
		return nil, errno
	}
	return file, nil
}

func configureTUN(name string, cidr string, mtu int) error {
	commands := [][]string{
		{"addr", "add", cidr, "dev", name},
		{"link", "set", "dev", name, "mtu", fmt.Sprintf("%d", mtu)},
		{"link", "set", "dev", name, "up"},
	}

	for _, args := range commands {
		cmd := exec.Command("ip", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("ip %v failed: %w: %s", args, err, string(out))
		}
	}
	return nil
}

// file descriptor
func (t *TUN) ReadPacket(buf []byte) (int, error) {
	return t.file.Read(buf)
}

func (t *TUN) WritePacket(packet []byte) error {
	_, err := t.file.Write(packet)
	return err
}

func (t *TUN) Close() error {
	return t.file.Close()
}
