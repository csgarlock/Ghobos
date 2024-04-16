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

var stepboards [16][64]bool = [16][64]bool{}

var moveBoards [5][64]Bitboard = [5][64]Bitboard{}
var pawnMoveBoards [2][64]Bitboard = [2][64]Bitboard{}
var pawnAttackBoards [2][64]Bitboard = [2][64]Bitboard{}

func InitializeMoveBoards() {
	InitializeStepBoard()
	FillSlidingAttacks(&bishopSteps, &moveBoards[Bishop])
	FillSlidingAttacks(&rookSteps, &moveBoards[Rook])
	var square Square
	for square = 0; square < 64; square++ {
		var bitboard Bitboard = EmptyBitboard
		for _, step := range kingSteps {
			if square.tryStep(step) {
				bitboard |= 1 << square.Step(step)
			}
		}
		moveBoards[King][square] = bitboard
		bitboard = EmptyBitboard
		for _, step := range knightSteps {
			if square.tryStep(step) {
				bitboard |= 1 << square.Step(step)
			}
		}
		moveBoards[Knight][square] = bitboard
		moveBoards[Queen][square] = moveBoards[Bishop][square] | moveBoards[Rook][square]
	}
}

func InitializeStepBoard() {
	for i, step := range allSteps {
		center := Square(35)
		centerStep := center.Step(step)
		rankDiff := centerStep.Rank() - center.Rank()
		fmt.Println(rankDiff)
		fileDiff := centerStep.File() - center.File()
		fmt.Println(fileDiff)
		var square Square
		for square = 0; square < 64; square++ {
			squareStep := square.Step(step)
			if squareStep.Rank()-square.Rank() == rankDiff && squareStep.File()-square.File() == fileDiff {
				stepboards[i][square] = true
			} else {
				stepboards[i][square] = false
			}
		}
	}
}

func FillSlidingAttacks(steps *[4]Step, resultBitboards *[64]Bitboard) {
	var square Square
	for _, step := range steps {
		for square = 0; square < 64; square++ {
			var stepSquare Square = square
			for stepSquare.tryStep(step) {
				stepSquare = stepSquare.Step(step)
				resultBitboards[square] |= 1 << stepSquare
			}
		}
	}
}

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
