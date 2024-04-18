package main

func main() {
	InitializeMoveBoards()
	state := FenState("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1")
	state.genAllMoves()
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
