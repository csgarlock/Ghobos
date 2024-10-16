package main

import (
	"fmt"
	"math/bits"
	"strconv"
)

type Bitboard uint64
type Square uint8

type SubsetIterator struct {
	n Bitboard
	d Bitboard
}

const (
	EmptyBitboard     Bitboard = 0
	UniversalBitboard Bitboard = ^EmptyBitboard

	Rank0 Bitboard = 0xff
	Rank1 Bitboard = Rank0 << (8 * 1)
	Rank2 Bitboard = Rank0 << (8 * 2)
	Rank3 Bitboard = Rank0 << (8 * 3)
	Rank4 Bitboard = Rank0 << (8 * 4)
	Rank5 Bitboard = Rank0 << (8 * 5)
	Rank6 Bitboard = Rank0 << (8 * 6)
	Rank7 Bitboard = Rank0 << (8 * 7)

	File0 Bitboard = 0x0101010101010101
	File1 Bitboard = File0 << 1
	File2 Bitboard = File0 << 2
	File3 Bitboard = File0 << 3
	File4 Bitboard = File0 << 4
	File5 Bitboard = File0 << 5
	File6 Bitboard = File0 << 6
	File7 Bitboard = File0 << 7
)

var ranks [8]Bitboard = [8]Bitboard{Rank0, Rank1, Rank2, Rank3, Rank4, Rank5, Rank6, Rank7}
var files [8]Bitboard = [8]Bitboard{File0, File1, File2, File3, File4, File5, File6, File7}

func SFS(square string) Square {
	rank, _ := strconv.Atoi(string(square[1]))
	rank--
	fileMap := [8]byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}
	file := 0
	for i, r := range fileMap {
		if square[0] == r {
			file = i
		}
	}
	return Square(rank*8 + file)
}

func sFromRankFile(file int, rank int) Square {
	return Square(rank*8 + file)
}

func (s Square) tryStep(step Step) bool { return stepboards[stepMap[step]][s] }
func (s Square) Step(step Step) Square  { return (s + Square(step)) % 64 }

func (s Square) Rank() int8 { return int8(s / 8) }
func (s Square) File() int8 { return int8(s % 8) }

func PopLSB(b *Bitboard) Square {
	lsb := Square(bits.TrailingZeros64(uint64(*b)))
	*b ^= (1 << Bitboard(lsb))
	return lsb
}

func PopMSB(b *Bitboard) Square {
	msb := 63 - Square(bits.LeadingZeros64(uint64(*b)))
	*b ^= (1 << Bitboard(msb))
	return msb
}

func BitCount(b Bitboard) int {
	return int(bits.OnesCount64(uint64(b)))
}

func NewSubsetIterator(d Bitboard) *SubsetIterator {
	return &SubsetIterator{n: 0, d: d}
}

func (sI *SubsetIterator) nextSubset() Bitboard {
	sI.n = (sI.n - sI.d) & sI.d
	return sI.n
}

func (sI *SubsetIterator) getSubset() Bitboard {
	return sI.n
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

func (s Square) String() string {
	fileMap := [8]string{"a", "b", "c", "d", "e", "f", "g", "h"}
	rank := s.Rank()
	file := s.File()
	return fileMap[file] + strconv.FormatInt(int64(rank+1), 10)
}
