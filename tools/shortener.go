package tools

import (
	"errors"
	"fmt"
	"strings"
)

const (
	alphabet = "123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	length   = len(alphabet)
	offset   = 1000000
)

func Encode(n int) string {
	var s = ""
	n += offset
	for ; n > 0; n = n / length {
		s = string(alphabet[n%length]) + s
	}
	return s
}

func Decode(s string) (int, error) {
	var n int
	for _, c := range []byte(s) {
		i := strings.IndexByte(alphabet, c)
		if i < 0 {
			return 0, errors.New(fmt.Sprintf("invalid character '%v' in string '%s'", c, s))
		}
		n = n*length + i
	}
	return n - offset, nil
}
