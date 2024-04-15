package main

import "fmt"

func main() {
	// InitializeMoveBoards()
	// for i, s := range moveBoards[Knight] {
	// 	fmt.Println("Rank = ", Square(i).Rank(), ", File = ", Square(i).File())
	// 	fmt.Println(s)
	// }
	move := BuildMove(9, 32, 1, 1)
	fmt.Printf("%b\n", move)
	fmt.Println(move)
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
