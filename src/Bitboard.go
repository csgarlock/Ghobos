package main

import (
	"fmt"
	"math/bits"
)

type Bitboard uint64
type Square uint8

const (
	EmptyBitboard     Bitboard = 0
	UniversalBitboard Bitboard = ^EmptyBitboard
)

func (s Square) tryStep(step Step) bool { return stepboards[stepMap[step]][s] }
func (s Square) Step(step Step) Square  { return (s + Square(step)) % 64 }

func (s Square) Rank() int8 { return int8(s / 8) }
func (s Square) File() int8 { return int8(s % 8) }
func LSB(b Bitboard) Square { return Square(bits.TrailingZeros64(uint64(b))) }
func PopLSB(b Bitboard) (Square, Bitboard) {
	lsb := LSB(b)
	return lsb, b ^ (1 << Bitboard(lsb))
}

func (b Bitboard) String() string {
	zeros := "00000000"
	outputS := ""
	for i := range 8 {
		lineS := fmt.Sprintf("%b", (b>>(8*i))&0xff)
		lineS = zeros[0:8-len(lineS)] + lineS
		lineArr := []rune(lineS)
		for i := 0; i < 4; i++ {
			lineArr[i], lineArr[7-i] = lineArr[7-i], lineArr[i]
		}
		lineS = string(lineArr)
		outputS = lineS + "\n" + outputS
	}
	return outputS
}
