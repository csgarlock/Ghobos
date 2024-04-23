package main

import (
	"fmt"
	"strconv"
)

func main() {
	InitializeMoveBoards()
	fmt.Println("Setup Finished")
	state := FenState("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	var dC int64 = 0
	fmt.Println(state)
	state.MakeMove(BuildMove(12, 28, 0, 0), &dC)
	fmt.Println(state)
	state.MakeMove(BuildMove(48, 40, 0, 0), &dC)
	fmt.Println(state)
	state.MakeMove(BuildMove(3, 39, 0, 0), &dC)
	fmt.Println(state)
	moves := state.genAllMoves(true)
	for i, move := range *moves {
		fmt.Println("Move " + strconv.FormatInt(int64(i), 10) + ": " + move.String())
	}
	fmt.Println(len(*moves))
	// var perftCounter int64 = 0
	// var checkCounter int64 = 0
	// var mateCounter int64 = 0
	// start := time.Now()
	// Perft(4, &perftCounter, state, &checkCounter, &mateCounter)
	// fmt.Println(time.Since(start))
	// moves := state.genAllMoves(true)
	// for _, move := range *moves {
	// 	state.MakeMove(move, &checkCounter)
	// 	fmt.Println(move)
	// 	perftCounter = 0
	// 	Perft(3, &perftCounter, state, &checkCounter, &mateCounter)
	// 	state.UnMakeMove(move)
	// 	fmt.Println(perftCounter)
	// }
	// fmt.Println(perftCounter)
	// fmt.Println(checkCounter)
	// fmt.Println(mateCounter)
}

func Perft(depth int64, moveCounter *int64, s *State, checkCounter *int64, mateCounter *int64) {
	if depth == 0 {
		*moveCounter++
		if s.check {
			moves := s.genAllMoves(true)
			if len(*moves) == 0 {
				*mateCounter++
			}
		}
	} else {
		moves := s.genAllMoves(true)
		if len(*moves) == 0 {
			*mateCounter++
		}
		for _, move := range *moves {
			s.MakeMove(move, checkCounter)
			Perft(depth-1, moveCounter, s, checkCounter, mateCounter)
			s.UnMakeMove(move)
		}
	}
}
