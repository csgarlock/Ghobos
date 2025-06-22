package main

import (
	"fmt"
	"strconv"
)

const (
	White = 0
	Black = 1
)

type Board [12]Bitboard

// Inclusive on both sides
func (b *Board) getPieceAtRanged(s Square, min uint8, max uint8) uint8 {
	for i := min; i <= max; i++ {
		if b[i]&(1<<Bitboard(s)) != EmptyBitboard {
			return i
		}
	}
	return NoPiece
}

func (b *Board) getColorPieceAt(s Square, color uint8) uint8 {
	squareBoard := boardFromSquare(s)
	if color == White {
		if b.uGet(WhitePawn)&squareBoard != EmptyBitboard {
			return WhitePawn
		} else if b.uGet(WhiteBishop)&squareBoard != EmptyBitboard {
			return WhiteBishop
		} else if b.uGet(WhiteKnight)&squareBoard != EmptyBitboard {
			return WhiteKnight
		} else if b.uGet(WhiteRook)&squareBoard != EmptyBitboard {
			return WhiteRook
		} else if b.uGet(WhiteQueen)&squareBoard != EmptyBitboard {
			return WhiteQueen
		} else if b.uGet(WhiteKing)&squareBoard != EmptyBitboard {
			return WhiteKing
		} else {
			return NoPiece
		}
	} else {
		if b.uGet(BlackPawn)&squareBoard != EmptyBitboard {
			return BlackPawn
		} else if b.uGet(BlackBishop)&squareBoard != EmptyBitboard {
			return BlackBishop
		} else if b.uGet(BlackKnight)&squareBoard != EmptyBitboard {
			return BlackKnight
		} else if b.uGet(BlackRook)&squareBoard != EmptyBitboard {
			return BlackRook
		} else if b.uGet(BlackQueen)&squareBoard != EmptyBitboard {
			return BlackQueen
		} else if b.uGet(BlackKing)&squareBoard != EmptyBitboard {
			return BlackKing
		} else {
			return NoPiece
		}
	}
}

func (b *Board) getPieceAt(s Square) uint8 {
	return b.getPieceAtRanged(s, 0, 11)
}

// Does not bounds check
func (b *Board) uGet(index uint8) Bitboard { return uGetA(&b[0], uint(index)) }

func (b *Board) uGetPtr(index uint8) *Bitboard { return uGetAPtr(&b[0], uint(index)) }

// Does not bounds check
func (b *Board) uSet(bitboard Bitboard, index uint8) { uSetA(&b[0], bitboard, uint(index)) }

func (b *Board) String() string {
	pieceMap := [12]string{"K", "Q", "R", "B", "N", "P", "k", "q", "r", "b", "n", "p"}
	result := [64]string{}
	for i, c := range pieceMap {
		bitboard := b[i]
		var spot Square
		for bitboard != EmptyBitboard {
			spot = PopLSB(&bitboard)
			if result[spot] != "" {
				fmt.Println("Two pieces located in same place")
			}
			result[spot] = c
		}
	}
	resultS := ""
	bottomLine := "  -----------------"
	for i := 0; i < 8; i++ {
		lineS := strconv.FormatInt(int64(i+1), 10) + " "
		for j := 0; j < 8; j++ {
			spot := result[i*8+j]
			if spot == "" {
				spot = " "
			}
			lineS += "|" + spot
		}
		resultS = lineS + "|\n" + bottomLine + "\n" + resultS
	}
	return bottomLine + "\n" + resultS + "   a b c d e f g h "
}
