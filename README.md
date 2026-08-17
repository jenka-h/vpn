# My VPN

A minimal Layer 3 point-to-point VPN written in Go. It creates a Linux TUN interface, reads raw IP packets, encrypts them with AES-256-GCM, wraps them in a custom UDP packet format, and sends them to a peer endpoint.

![Remembrance](./asset/tomoko.jpg)

## Features

- Linux TUN Layer 3 interface.
- UDP transport between two endpoints.
- Bidirectional tunnel:

  ```text
  TUN -> AES-GCM encrypt -> UDP -> AES-GCM decrypt -> TUN
  ```

- AES-256-GCM encryption and integrity.
- Pre-shared key authentication.
- Custom VPN packet format:

  ```text
  Version | Sequence | Nonce | Ciphertext + Auth Tag
  ```

- `Version` and `Sequence` are used as AES-GCM AAD, so they are authenticated by the GCM tag.
- Replay protection using sequence numbers and a sliding replay window.
- Configurable TUN IP, listen address, peer address, key file, and MTU.

## Requirements

This project targets Linux.

Required tools:

- Go 1.22+
- `ip` from `iproute2`
- root privileges or equivalent capability for TUN setup
- optional: `tcpdump`, `curl`, `python3`, `openssl`

Install common tools on Debian/Ubuntu:

```sh
sudo apt update
sudo apt install -y golang iproute2 tcpdump curl python3 openssl
```

## Build

From this directory:

```sh
go build -o vpn ./cmd/vpn
```

## Create a pre-shared key

Both endpoints must use the same key.

```sh
openssl rand -hex 32 > vpn.key
```

The key file may contain one of these formats:

- 32 raw bytes
- 64 hex characters
- base64-encoded 32-byte key

## CLI options

```sh
sudo ./vpn \
  -tun tun0 \
  -tun-ip 10.10.0.1/24 \
  -listen :51820 \
  -peer 192.168.20.2:51820 \
  -key vpn.key \
  -mtu 1400
```

Options:

| Flag | Description | Example |
|---|---|---|
| `-tun` | TUN interface name | `tun0` |
| `-tun-ip` | TUN IP address with CIDR | `10.10.0.1/24` |
| `-listen` | UDP listen address | `:51820` or `192.168.100.1:51820` |
| `-peer` | UDP address of the other endpoint | `192.168.100.2:51820` |
| `-key` | Pre-shared key file | `vpn.key` |
| `-mtu` | TUN MTU | `1400` |

Important: `-peer` must be the peer's real/underlay IP address, not the VPN IP.

---

# Usage with two VMs or two machines

Use this when you want the most realistic setup.

Example topology:

| Endpoint | Underlay IP | VPN IP |
|---|---:|---:|
| VM A | `192.168.10.2` | `10.10.0.1/24` |
| VM B | `192.168.20.2` | `10.10.0.2/24` |

Copy the same `vpn` binary and `vpn.key` to both VMs.

## Run endpoint A

On VM A:

```sh
sudo ./vpn \
  -tun tun0 \
  -tun-ip 10.10.0.1/24 \
  -listen :51820 \
  -peer 192.168.20.2:51820 \
  -key vpn.key
```

## Run endpoint B

On VM B:

```sh
sudo ./vpn \
  -tun tun0 \
  -tun-ip 10.10.0.2/24 \
  -listen :51820 \
  -peer 192.168.10.2:51820 \
  -key vpn.key
```

## Test ping

From VM A:

```sh
ping 10.10.0.2
```

From VM B:

```sh
ping 10.10.0.1
```

## Test file transfer

On VM A:

```sh
dd if=/dev/urandom of=/tmp/test-3mb.bin bs=1M count=3
python3 -m http.server 8000 --bind 10.10.0.1 --directory /tmp
```

On VM B:

```sh
curl http://10.10.0.1:8000/test-3mb.bin -o /tmp/received.bin
sha256sum /tmp/received.bin
```

On VM A:

```sh
sha256sum /tmp/test-3mb.bin
```

The hashes should match.

## Verify encrypted underlay traffic

On the underlay interface, for example `eth0`:

```sh
sudo tcpdump -ni eth0 udp port 51820 -X
```

Then run a ping through the VPN:

```sh
ping 10.10.0.2
```

Expected result:

- On `tun0`, traffic appears as normal IP traffic, e.g. ICMP between `10.10.0.1` and `10.10.0.2`.
- On `eth0`, traffic appears as UDP packets between the real underlay IPs.
- The original plaintext packet is not visible on the underlay capture.

---

# Usage with Linux network namespaces

Use this if you only have one Linux machine and want to simulate two separate hosts.

Topology:

```text
nsA                                      nsB
VPN IP:      10.10.0.1/24               VPN IP:      10.10.0.2/24
Underlay IP: 192.168.100.1/24  <veth>   Underlay IP: 192.168.100.2/24
```

## 1. Clean old test state

From the repository root:

```sh
sudo ip netns delete nsA 2>/dev/null
sudo ip netns delete nsB 2>/dev/null
sudo ip link delete vethA 2>/dev/null
sudo ip link delete vethB 2>/dev/null
sudo rm -f /tmp/test-3mb.bin /tmp/received.bin
```

Errors such as `No such file or directory` are safe to ignore.

## 2. Create namespaces and underlay link

```sh
sudo ip netns add nsA
sudo ip netns add nsB

sudo ip link add vethA type veth peer name vethB

sudo ip link set vethA netns nsA
sudo ip link set vethB netns nsB

sudo ip netns exec nsA ip addr add 192.168.100.1/24 dev vethA
sudo ip netns exec nsB ip addr add 192.168.100.2/24 dev vethB

sudo ip netns exec nsA ip link set lo up
sudo ip netns exec nsA ip link set vethA up

sudo ip netns exec nsB ip link set lo up
sudo ip netns exec nsB ip link set vethB up
```

Check underlay connectivity:

```sh
sudo ip netns exec nsA ping -c 3 192.168.100.2
```

This must work before starting the VPN.

## 3. Run endpoint A

Terminal 1, from the repository root:

```sh
sudo ip netns exec nsA ./vpn/vpn \
  -tun tun0 \
  -tun-ip 10.10.0.1/24 \
  -listen 192.168.100.1:51820 \
  -peer 192.168.100.2:51820 \
  -key vpn/vpn.key
```

## 4. Run endpoint B

Terminal 2, from the repository root:

```sh
sudo ip netns exec nsB ./vpn/vpn \
  -tun tun0 \
  -tun-ip 10.10.0.2/24 \
  -listen 192.168.100.2:51820 \
  -peer 192.168.100.1:51820 \
  -key vpn/vpn.key
```

## 5. Check TUN interfaces

Terminal 3:

```sh
sudo ip netns exec nsA ip addr show tun0
sudo ip netns exec nsB ip addr show tun0
```

Expected:

- `nsA` has `10.10.0.1/24`
- `nsB` has `10.10.0.2/24`

## 6. Test VPN ping

```sh
sudo ip netns exec nsA ping -c 4 10.10.0.2
sudo ip netns exec nsB ping -c 4 10.10.0.1
```

Expected:

```text
4 packets transmitted, 4 received
```

## 7. Test file transfer over VPN

Create a file:

```sh
dd if=/dev/urandom of=/tmp/test-3mb.bin bs=1M count=3
sudo rm -f /tmp/received.bin
```

Terminal 4, run an HTTP server in `nsA`:

```sh
sudo ip netns exec nsA python3 -m http.server 8000 --bind 10.10.0.1 --directory /tmp
```

Terminal 3, download from `nsB` through the VPN:

```sh
sudo ip netns exec nsB curl http://10.10.0.1:8000/test-3mb.bin -o /tmp/received.bin
```

Verify checksums:

```sh
sha256sum /tmp/test-3mb.bin
sha256sum /tmp/received.bin
```

The hashes should match.

## 8. Verify traffic encryption

Capture plaintext VPN-side traffic:

```sh
sudo ip netns exec nsA tcpdump -ni tun0
```

In another terminal:

```sh
sudo ip netns exec nsA ping -c 3 10.10.0.2
```

Expected on `tun0`:

```text
IP 10.10.0.1 > 10.10.0.2: ICMP echo request
IP 10.10.0.2 > 10.10.0.1: ICMP echo reply
```

Capture encrypted underlay traffic:

```sh
sudo ip netns exec nsA tcpdump -ni vethA udp port 51820 -X
```

Run ping again:

```sh
sudo ip netns exec nsA ping -c 3 10.10.0.2
```

Expected on `vethA`:

```text
IP 192.168.100.1.51820 > 192.168.100.2.51820: UDP
```

The underlay payload should appear as encrypted bytes, not as direct plaintext ICMP between `10.10.0.1` and `10.10.0.2`.

## 9. Cleanup

Stop the VPN processes and HTTP server with `Ctrl+C`, then run:

```sh
sudo ip netns delete nsA
sudo ip netns delete nsB
sudo rm -f /tmp/test-3mb.bin /tmp/received.bin
```

---

# Troubleshooting

## `/dev/net/tun`: permission denied

Run with `sudo`:

```sh
sudo ./vpn ...
```

Check that the TUN module/device exists:

```sh
sudo modprobe tun
ls -l /dev/net/tun
```

## `ip` command not found

Install `iproute2`:

```sh
sudo apt install iproute2
```

## TUN already exists

Delete it:

```sh
sudo ip link delete tun0
```

For namespaces:

```sh
sudo ip netns exec nsA ip link delete tun0
```

## `curl`: permission denied writing `/tmp/received.bin`

Remove the old file first:

```sh
sudo rm -f /tmp/received.bin
```

Then retry the download.

## Ping does not work

Check these in order:

```sh
sudo ip netns exec nsA ping -c 3 192.168.100.2
sudo ip netns exec nsA ip addr
sudo ip netns exec nsB ip addr
sudo ip netns exec nsA ip route
sudo ip netns exec nsB ip route
```

Also make sure both VPN processes are still running and both endpoints use the same `vpn.key`.

## Reference

- AES-GCM authenticated encryption: https://pkg.go.dev/crypto/cipher#AEAD
- Linux TUN/TAP documentation: https://docs.kernel.org/networking/tuntap.html
