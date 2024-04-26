package main

import (
	"fmt"
	"strconv"
	"time"
)

func main() {
	InitializeMoveBoards()
	// fmt.Println(pawnAttackBoards[Black][23])
	// fmt.Println("Setup Finished")
	state := FenState("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	PerftRunner(6, state)
	// PerftChecker(3, state)
	// var dC int64 = 0
	// fmt.Println(state)
	// state.MakeMove(BuildMove(SFS("g2"), SFS("h1"), QueenPromotion, PromotionSpecialMove), &dC, &dC, &dC)
	// fmt.Println(state)
	// state.UnMakeMove(BuildMove(SFS("g2"), SFS("h1"), QueenPromotion, PromotionSpecialMove))
	// fmt.Println(state)
	// state.MakeMove(BuildMove(SFS("h3"), SFS("g2"), 0, 0), &dC, &dC, &dC)
	// fmt.Println(state)
	// fmt.Println(state.castleHistory)
	// state.UnMakeMove(BuildMove(4, 6, 0, CastleSpecialMove))
	// fmt.Println(state)
	// fmt.Println(state.castleHistory)
	// state.MakeMove(BuildMove(31, 39, 0, 0), &dC)
	// fmt.Println(state)
	// state.UnMakeMove(BuildMove(31, 39, 0, 0))
	// fmt.Println(state)
	// state.MakeMove(BuildMove(53, 37, 0, 0), &dC)
	// fmt.Println(state.enPassantSquare)
	// fmt.Println(state)
	// fmt.Println(state.enPassantSquareHistory)
	// moves := state.genAllMoves(true)
	// for i, move := range *moves {
	// 	fmt.Printf("Move %d\n", i)
	// 	fmt.Println(move)
	// }
	// fmt.Println(len(*moves))
	// var perftCounter int64 = 0
	// var checkCounter int64 = 0
	// var mateCounter int64 = 0
	// var enPassantCounter int64 = 0
	// var castleCounter int64 = 0
	// var timer time.Duration = 0
	// start := time.Now()
	// Perft(3, &perftCounter, state, &checkCounter, &mateCounter, &enPassantCounter, &castleCounter, &timer)
	// moves := state.genAllMoves(true)
	// for _, move := range *moves {
	// 	state.MakeMove(move, &checkCounter, &enPassantCounter, &castleCounter)
	// 	fmt.Println(move)
	// 	if move == BuildMove(14, 30, 0, 0) {
	// 		for _, m := range *state.genAllMoves(true) {
	// 			fmt.Println(m)
	// 		}
	// 	}
	// 	perftCounter = 0
	// 	Perft(1, &perftCounter, state, &checkCounter, &mateCounter, &enPassantCounter, &castleCounter, &timer)
	// 	state.UnMakeMove(move)
	// 	fmt.Println(perftCounter)
	// }
}

func PerftChecker(depth int64, s *State) {
	var d int64 = 0
	currentDepth := depth
	for {
		moves := s.genAllMoves(true)
		for i, move := range *moves {
			fmt.Printf("Move %d: ", i)
			fmt.Println(move.ShortString())
			s.MakeMove(move, &d, &d, &d)
			var counter int64 = 0
			Perft(currentDepth-1, &counter, s, &d, &d, &d, &d, (*time.Duration)(&d))
			fmt.Println(counter)
			s.UnMakeMove(move)
		}
		move_selection := GetUserNumber("Enter move number: ")
		s.MakeMove((*moves)[move_selection], &d, &d, &d)
		currentDepth--
	}
}

func PerftRunner(depth int64, s *State) {
	var perftCounter int64 = 0
	var checkCounter int64 = 0
	var mateCounter int64 = 0
	var enPassantCounter int64 = 0
	var castleCounter int64 = 0
	var timer time.Duration = 0
	start := time.Now()
	Perft(depth, &perftCounter, s, &checkCounter, &mateCounter, &enPassantCounter, &castleCounter, &timer)
	fmt.Printf("Highest Depth Node Count: %d\n", perftCounter)
	fmt.Printf("Check Counter: %d\n", checkCounter)
	fmt.Printf("Mate Counter: %d\n", mateCounter)
	fmt.Printf("En Passant Counter: %d\n", enPassantCounter)
	fmt.Printf("Castle Counter: %d\n", castleCounter)
	fmt.Println(time.Since(start))
	fmt.Println(timer)
}

func Perft(depth int64, moveCounter *int64, s *State, checkCounter *int64, mateCounter *int64, enPassantCounter *int64, castleCounter *int64, sectionTimer *time.Duration) {
	if depth == 0 {
		*moveCounter++
		if s.check {
			//start := time.Now()
			moves := s.genAllMoves(true)
			//*sectionTimer += time.Since(start)
			if len(*moves) == 0 {
				*mateCounter++
			}
		}
	} else {
		start := time.Now()
		moves := s.genAllMoves(true)
		*sectionTimer += time.Since(start)
		if len(*moves) == 0 {
			*mateCounter++
		}
		for _, move := range *moves {
			// fmt.Println(s)
			// fmt.Println("Trying to Make " + move.String())
			s.MakeMove(move, checkCounter, enPassantCounter, castleCounter)
			Perft(depth-1, moveCounter, s, checkCounter, mateCounter, enPassantCounter, castleCounter, sectionTimer)
			// fmt.Println(s)
			// fmt.Println("Trying to Unmake " + move.String())
			s.UnMakeMove(move)
		}
	}
}

func GetUserNumber(prompt string) int {
	for {
		var userInput string
		var num int

		fmt.Print(prompt)
		_, err := fmt.Scanln(&userInput)
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		num, err = strconv.Atoi(userInput)
		if err != nil {
			fmt.Println("Error converting to integer:", err)
			continue
		}
		return num
	}
}
