package generator

import (
	crand "crypto/rand"
	"fmt"
	"hash/fnv"
	"strconv"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	ulidMu      sync.Mutex
	ulidEntropy = ulid.Monotonic(crand.Reader, 0)
)

func newULID() string {
	ulidMu.Lock()
	id := ulid.MustNew(ulid.Timestamp(time.Now()), ulidEntropy)
	ulidMu.Unlock()
	return id.String()
}

func newOrderID() string       { return "ord_" + newULID() }
func newTransactionID() string { return "txn_" + newULID() }
func newRequestID() string     { return "req_" + newULID() }

// (опционально) симпатичный track number с чексуммой (сохраняем формат WB...)
func newTrackNumber() string {
	ts36 := strconv.FormatInt(time.Now().UTC().Unix(), 36) // компактное время
	r := make([]byte, 4)
	if _, err := crand.Read(r); err != nil {
		// в маловероятном случае fallback к pseudo-rand
		for i := range r {
			r[i] = byte(rng.Intn(256))
		}
	}
	r36 := strconv.FormatUint(uint64(uint32(r[0])<<24|uint32(r[1])<<16|uint32(r[2])<<8|uint32(r[3])), 36)
	core := (ts36 + r36)
	if len(core) > 10 {
		core = core[:10]
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(core))
	check := int(h.Sum32() % 97) // 00..96
	return fmt.Sprintf("WB%s%02d", core, check)
}
