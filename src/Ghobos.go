package main

import (
	"fmt"
	"strconv"
	"time"
)

func main() {
	InitializeMoveBoards()
	InitializeEvalVariables()
	SetupTable(1024)
	state := FenState("1Nk5/3R4/4R3/3P4/2p5/2P4p/5PPP/6K1 w - - 2 40")
	state.check = false
	var nodesSearched int32 = 0
	start := time.Now()
	bestMove := state.getBestMove(5, &nodesSearched)
	fmt.Println(time.Since(start))
	fmt.Println(bestMove)
	fmt.Println(nodesSearched)
}

func PerftChecker(depth int64, s *State) {
	var d int64 = 0
	currentDepth := depth
	for {
		moves := s.genAllMoves(true)
		for i, move := range *moves {
			fmt.Print(move.ShortString())
			fmt.Printf(", Move %d: ", i)
			s.MakeMove(move)
			var counter int64 = 0
			Perft(currentDepth-1, &counter, s, &d, &d, &d, &d, (*time.Duration)(&d))
			fmt.Println(counter)
			s.UnMakeMove(move)
		}
		move_selection := GetUserNumber("Enter move number: ")
		s.MakeMove((*moves)[move_selection])
		currentDepth--
	}
}

// Boards Complete to Depth 5: 1, 2, 3, 5, 6
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
			s.MakeMove(move)
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
