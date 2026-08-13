package main

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	vpn "sister2/vpn/internal"
)

func main() {
	cfg, err := vpn.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	key, err := vpn.LoadKey(cfg.KeyFile)
	if err != nil {
		log.Fatalf("load key: %v", err)
	}
	crypto, err := vpn.NewCrypto(key)
	if err != nil {
		log.Fatalf("init crypto: %v", err)
	}

	tun, err := vpn.NewTUN(cfg.TunName, cfg.TunIP, cfg.MTU)
	if err != nil {
		log.Fatalf("init tun: %v", err)
	}
	defer tun.Close()

	udp, err := vpn.NewUDPTransport(cfg.ListenAddr, cfg.PeerAddr)
	if err != nil {
		log.Fatalf("init udp: %v", err)
	}
	defer udp.Close()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	var sequence uint64
	replay := vpn.NewReplayGuard()
	errs := make(chan error, 2)

	go func() {
		errs <- tunToNetwork(tun, udp, crypto, &sequence, cfg.MTU)
	}()
	go func() {
		errs <- networkToTun(udp, crypto, replay, tun)
	}()

	log.Printf("vpn started: tun=%s tun-ip=%s listen=%s peer=%s", cfg.TunName, cfg.TunIP, cfg.ListenAddr, cfg.PeerAddr)

	select {
	case sig := <-signals:
		log.Printf("received %s, shutting down", sig)
	case err := <-errs:
		if err != nil && !errors.Is(err, os.ErrClosed) {
			log.Printf("tunnel stopped: %v", err)
		}
	}
}

func tunToNetwork(
	tun *vpn.TUN,
	udp *vpn.UDPTransport,
	crypto *vpn.Crypto,
	sequence *uint64,
	mtu int,
) error {
	buf := make([]byte, mtu)
	for {
		n, err := tun.ReadPacket(buf)
		if err != nil {
			return err
		}

		seq := atomic.AddUint64(sequence, 1)
		aad := vpn.EncodeAuthenticatedHeader(vpn.ProtocolVersion, seq)
		nonce, ciphertext, err := crypto.Encrypt(buf[:n], aad)
		if err != nil {
			return err
		}

		encoded, err := vpn.EncodePacket(&vpn.Packet{
			Version:    vpn.ProtocolVersion,
			Sequence:   seq,
			Nonce:      nonce,
			Ciphertext: ciphertext,
		})
		if err != nil {
			return err
		}
		if err := udp.Send(encoded); err != nil {
			return err
		}
	}
}

func networkToTun(
	udp *vpn.UDPTransport,
	crypto *vpn.Crypto,
	replay *vpn.ReplayGuard,
	tun *vpn.TUN,
) error {
	buf := make([]byte, 65535)
	for {
		n, addr, err := udp.Receive(buf)
		if err != nil {
			return err
		}
		if !udp.IsPeer(addr) {
			log.Printf("drop packet from unexpected peer: %s", addr)
			continue
		}

		packet, err := vpn.DecodePacket(buf[:n])
		if err != nil {
			log.Printf("drop malformed packet: %v", err)
			continue
		}
		plaintext, err := crypto.Decrypt(packet.Nonce, packet.Ciphertext, packet.AuthenticatedData())
		if err != nil {
			log.Printf("drop unauthenticated packet: %v", err)
			continue
		}
		if !replay.Accept(packet.Sequence) {
			log.Printf("drop replayed packet: sequence=%d", packet.Sequence)
			continue
		}
		if err := tun.WritePacket(plaintext); err != nil {
			return err
		}
	}
}
