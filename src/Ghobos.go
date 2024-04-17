package main

import (
	"fmt"
	"math"
	"math/bits"
)

func main() {
	InitializeMoveBoards()
	total := 0.0
	for i, s := range moveBoards[Rook] {
		mask := s & (^(Rank0 | Rank7) | ranks[Square(i).Rank()]) & (^(File0 | File7) | files[Square(i).File()])
		sTotal := bits.OnesCount64(uint64(mask))
		fmt.Println(mask)
		total += math.Pow(2, float64(sTotal))
		// fmt.Println("Rank = ", Square(i).Rank(), ", File = ", Square(i).File())
		// fmt.Println(s)
	}
	fmt.Println(total)
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
