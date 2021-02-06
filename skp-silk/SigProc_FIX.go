package skpsilk

import "math"

const (
	MAX_ORDER_LPC          = 16
	MAX_CORRELATION_LENGTH = 640
	LSF_COS_TAB_SZ_FIX     = 128
)

func min_32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func RSHIFT_ROUND(a, shift int32) int32 {
	if shift == 1 {
		return (a >> 1) + (a & 1)
	} else {
		return ((a >> (shift - 1)) + 1) >> 1
	}
}

func ROR32(a32 int32, rot int) int32 {
	x := uint32(a32)
	r := uint32(rot)
	m := uint32(-rot)
	if rot <= 0 {
		return int32((x << m) | (x >> (32 - m)))
	} else {
		return int32((x << (32 - r)) | (x >> r))
	}
}

func SAT16(a int16) int16 {
	if a > math.MaxInt16 {
		return math.MaxInt16
	} else if a < math.MinInt16 {
		return math.MinInt16
	} else {
		return a
	}
}

func RAND(seed int32) int32 {
	return MLA_ovflw(907633515, (seed), 196314165)
}

func MLA_ovflw(a32, b32, c32 int32) int32 {
	return ADD32_ovflw(a32, uint32(b32)*uint32(c32))
}

func ADD32_ovflw(a int32, b uint32) int32 {
	return int32(uint32(a) + uint32(b))
}

func RSHIFT(a, shift int) int {
	return RSHIFT32(a, shift)
}

func RSHIFT32(a, shift int) int {
	return a >> shift
}
