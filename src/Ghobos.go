package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	InitializeMoveBoards()
	state := FenState("rnbqkbnr/pppp1ppp/8/4p3/4P1Q1/8/PPPP1PPP/RNB1KBNR b KQkq - 1 2")
	fmt.Println(state)
	occupied := state.board[0]
	for i := 1; i < 12; i++ {
		occupied |= state.board[i]
	}
	fmt.Println(occupied)
	rookMoves := getRookMoves(25, occupied)
	bishopMoves := getBishopMoves(25, occupied)
	fmt.Println(rookMoves | bishopMoves)

	fmt.Println(time.Since(start))
	// for i, s := range moveBoards[Rook] {
	// 	mask := s & (^(Rank0 | Rank7) | ranks[Square(i).Rank()]) & (^(File0 | File7) | files[Square(i).File()])
	// 	sTotal := bits.OnesCount64(uint64(mask))
	// 	fmt.Println(mask)
	// 	total += math.Pow(2, float64(sTotal))
	// 	// fmt.Println("Rank = ", Square(i).Rank(), ", File = ", Square(i).File())
	// 	// fmt.Println(s)
	// }
	// fmt.Println(total)
}

func debugStepBoard(stepboard [64]bool) {
	result := ""
	for i := range 8 {
		row := ""
		for j := range 8 {
			if stepboard[i*8+j] {
				row += "1"
			} else {
				row += "0"
			}
		}
		result = row + "\n" + result
	}
	fmt.Println(result)
}
