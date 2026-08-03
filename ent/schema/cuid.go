package schema

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var cuidCounter atomic.Uint32

// NewCUID provides the client-side default used by the legacy Prisma models.
func NewCUID() string {
	timestamp := padBase36(uint64(time.Now().UnixMilli()), 8)
	var random [10]byte
	if _, err := rand.Read(random[:]); err != nil {
		binary.BigEndian.PutUint32(random[:4], cuidCounter.Add(1))
	}
	counter := padBase36(uint64(cuidCounter.Add(1)), 4)
	return "c" + timestamp + base64.RawURLEncoding.EncodeToString(random[:]) + counter
}

func padBase36(value uint64, width int) string {
	encoded := strconv.FormatUint(value, 36)
	if len(encoded) >= width {
		return encoded[len(encoded)-width:]
	}
	return strings.Repeat("0", width-len(encoded)) + encoded
}
