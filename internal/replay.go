// menolak paket duplikat
package internal

import "sync"

const replayWindowSize uint64 = 64

type ReplayGuard struct {
	mu      sync.Mutex
	highest uint64
	bitmap  uint64
	seenAny bool
}

// entitas pengurus paket dengan seq tertinggi
func NewReplayGuard() *ReplayGuard {
	return &ReplayGuard{}
}

// cek sequence
func (r *ReplayGuard) Accept(sequence uint64) bool {
	r.mu.Lock() // melakukan lock
	defer r.mu.Unlock()

	if !r.seenAny { // cek bit sequence
		r.highest = sequence
		r.bitmap = 1
		r.seenAny = true
		return true
	}

	if sequence > r.highest {
		shift := sequence - r.highest
		if shift >= replayWindowSize {
			r.bitmap = 1
		} else {
			r.bitmap = (r.bitmap << shift) | 1
		}
		r.highest = sequence
		return true
	}

	delta := r.highest - sequence // cek perbedaan dan buat windowsize
	if delta >= replayWindowSize {
		return false
	}
	mask := uint64(1) << delta
	if r.bitmap&mask != 0 {
		return false
	}
	r.bitmap |= mask
	return true
}
