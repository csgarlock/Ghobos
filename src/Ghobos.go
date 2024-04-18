package main

import (
	"fmt"
)

func main() {
	InitializeMoveBoards()
	state := FenState("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1")
	state.genAllMoves()
	var occupied Bitboard = 0b10000010011001010100101000001011000001000001011010100
	fmt.Println(occupied)
	fmt.Println(getRookMoves(37, Bitboard(occupied)))
}
