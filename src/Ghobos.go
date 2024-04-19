package main

import (
	"fmt"
	"math/rand"
)

func main() {
	InitializeMoveBoards()
	checkMagics()
}

func checkMagics() {
	for i := 0; i < 100000; i++ {
		var occupied Bitboard = Bitboard(rand.Uint64() & rand.Uint64())
		square := Square(rand.Intn(64))
		bruteTestBishop := findBlockedSlidingAttack(square, &bishopSteps, occupied)
		magicTestBishop := getBishopMoves(square, occupied)
		bruteTestRook := findBlockedSlidingAttack(square, &rookSteps, occupied)
		magicTestRook := getRookMoves(square, occupied)
		if bruteTestBishop != magicTestBishop {
			fmt.Println("Failed Bishop Compare")
			fmt.Println(square)
			fmt.Println("Brute")
			fmt.Println(bruteTestBishop)
			fmt.Println("Magic")
			fmt.Println(magicTestBishop)
		}
		if bruteTestRook != magicTestRook {
			fmt.Println("Failed Rook Compare")
			fmt.Println(square)
			fmt.Println("Brute")
			fmt.Println(bruteTestRook)
			fmt.Println("Magic")
			fmt.Println(magicTestRook)
		}
	}
}
