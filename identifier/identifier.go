package identifier

import (
	"encoding/base32"
	"hash/fnv"
)

func New(tokens ...string) string {
	h := fnv.New64()
	for i, t := range tokens {
		if i > 0 {
			h.Write([]byte{0x31})
		}
		h.Write([]byte(t))
	}
	sum := h.Sum(nil)

	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(sum)
}
