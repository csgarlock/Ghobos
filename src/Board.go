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
	if color == White {
		return b.getPieceAtRanged(s, 0, 5)
	} else {
		return b.getPieceAtRanged(s, 6, 11)
	}
}

func (b *Board) getPieceAt(s Square) uint8 {
	return b.getPieceAtRanged(s, 0, 11)
}

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
	bottomline := "  -----------------"
	for i := 0; i < 8; i++ {
		lineS := strconv.FormatInt(int64(i+1), 10) + " "
		for j := 0; j < 8; j++ {
			spot := result[i*8+j]
			if spot == "" {
				spot = " "
			}
			lineS += "|" + spot
		}
		resultS = lineS + "|\n" + bottomline + "\n" + resultS
	}
	return bottomline + "\n" + resultS + "   a b c d e f g h "
}
