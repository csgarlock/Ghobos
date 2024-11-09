package main

import (
	"fmt"
	"time"
)

func PerftChecker(depth int64, s *State) {
	var d int64 = 0
	currentDepth := depth
	for {
		moves := s.quickGenMoves()
		for i, move := range *moves {
			fmt.Print(move.ShortString())
			fmt.Printf(", Move %d: ", i)
			s.MakeMove(move)
			var counter int64 = 0
			Perft(currentDepth-1, &counter, s, (*time.Duration)(&d))
			fmt.Println(counter)
			s.UnMakeMove(move)
		}
		move_selection := GetUserNumber("Enter move number: ")
		s.MakeMove((*moves)[move_selection])
		currentDepth--
	}
}

func PerftTester() {
	PerftRunner(5, FenState("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"), 4865609)
	PerftRunner(5, FenState("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"), 193690690)
	PerftRunner(5, FenState("8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1"), 674624)
	PerftRunner(5, FenState("rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8"), 89941194)
	PerftRunner(5, FenState("r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10"), 164075551)
	PerftRunner(5, FenState("n1n5/PPPk4/8/8/8/8/4Kppp/5N1N b - - 0 1"), 3605103)
}

func PerftRunner(depth int64, s *State, expectedCount int64) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(s)
			fmt.Println(s.fenString())
			fmt.Println(s.pinnedBoards[0])
			fmt.Println(s.pinnedBoards[1])
			for i := range 2 {
				fmt.Println()
				for j := range 8 {
					fmt.Println(s.pinners[i][j])
				}
			}
			panic("AHHHH")
		}
	}()
	var perftCounter int64 = 0
	var timer time.Duration = 0
	start := time.Now()
	Perft(depth, &perftCounter, s, &timer)
	fmt.Printf("Expected Node Count: %d. Highest Depth Node Count: %d. Found in %v\n", expectedCount, perftCounter, time.Since(start))
	if perftCounter != expectedCount {
		fmt.Println("Error! expected count not equal to found count")
	}
}

func Perft(depth int64, moveCounter *int64, s *State, sectionTimer *time.Duration) {
	if depth != 0 {
		start := time.Now()
		moves := s.quickGenMoves()
		*sectionTimer += time.Since(start)
		for _, move := range *moves {
			s.MakeMove(move)
			Perft(depth-1, moveCounter, s, sectionTimer)
			s.UnMakeMove(move)
		}
	} else {
		*moveCounter++
	}
}
