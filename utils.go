package db

import (
	"strconv"
)

func getBRCKWs(a, b uint64) []string {
	if a == b {
		return []string{makeTimeKW(a, 0)}
	}

	result := make([]string, 0)
	var t int
	for t = 64 - 1; t >= 0; t-- {
		if bit(a, t) != bit(b, t) {
			break
		}
	}
	if lastNBitsAllZero(a, t+1) {
		if lastNBitsAllOne(b, t+1) {
			result = append(result, makeTimeKW(a>>(t+1), t+1))
		} else {
			result = append(result, makeTimeKW(a>>t, t))
		}
	} else {
		var u int
		for u = 0; u < t; u++ {
			if bit(a, u) == 1 {
				break
			}
		}
		for i := t - 1; i >= u+1; i-- {
			if bit(a, i) == 0 {
				result = append(result, makeTimeKW((a>>(i+1))<<1+1, i))
			}
		}
		result = append(result, makeTimeKW(a>>u, u))
	}

	if lastNBitsAllOne(b, t+1) {
		result = append(result, makeTimeKW(b>>t, t))
	} else {
		var v int
		for v = 0; v < t; v++ {
			if bit(b, v) == 0 {
				break
			}
		}
		for i := t - 1; i >= v+1; i-- {
			if bit(b, i) == 1 {
				result = append(result, makeTimeKW((b>>(i+1))<<1, i))
			}
		}
		result = append(result, makeTimeKW(b>>v, v))
	}

	return result
}

func bit(x uint64, n int) uint {
	if x&(1<<n) != 0 {
		return 1
	} else {
		return 0
	}
}

func lastNBitsAllZero(x uint64, n int) bool {
	return x == ((x >> n) << n)
}

func lastNBitsAllOne(x uint64, n int) bool {
	return (x - ((x >> n) << n)) == ((1 << n) - 1)
}

func makeTimeKW(prefix uint64, suffixLen int) string {
	return strconv.FormatInt(int64(prefix), 10) + ":" + strconv.FormatInt(int64(64-suffixLen), 10)
}
