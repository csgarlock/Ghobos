package main

import (
	"fmt"
	"time"
)

var genTimer = NewRunningTimer()
var makeTimer = NewRunningTimer()
var unMakeTimer = NewRunningTimer()

func PerftChecker(depth int64, s *State) {
	currentDepth := depth
	for {
		moves := s.quickGenMoves()
		for i, move := range *moves {
			fmt.Print(move.ShortString())
			fmt.Printf(", Move %d: ", i)
			s.MakeMove(move)
			var counter int64 = 0
			Perft(currentDepth-1, &counter, s)
			fmt.Println(counter)
			s.UnMakeMove(move)
		}
		move_selection := GetUserNumber("Enter move number: ")
		s.MakeMove((*moves)[move_selection])
		currentDepth--
	}
}

func PerftTester() {
	PerftRunner(6, FenState("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"), 119060324)
	PerftRunner(5, FenState("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"), 193690690)
	PerftRunner(7, FenState("8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1"), 178633661)
	PerftRunner(5, FenState("rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8"), 89941194)
	PerftRunner(5, FenState("r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10"), 164075551)
	PerftRunner(5, FenState("n1n5/PPPk4/8/8/8/8/4Kppp/5N1N b - - 0 1"), 3605103)
}

func PerftRunner(depth int64, s *State, expectedCount int64) {
	genTimer.Reset()
	makeTimer.Reset()
	unMakeTimer.Reset()
	var perftCounter int64 = 0
	start := time.Now()
	Perft(depth, &perftCounter, s)
	duration := time.Since(start)
	rate := float64(perftCounter) / duration.Seconds() / 1_000_000.0
	fmt.Printf("Expected Node Count: %d. Highest Depth Node Count: %d. Found in %v. Rate: %.2f Million Nodes per Second\n", expectedCount, perftCounter, duration, rate)
	fmt.Printf("Move Gen Total: %v. Make Total: %v. Unmake Total: %v\n", genTimer.total, makeTimer.total, unMakeTimer.total)
	if perftCounter != expectedCount {
		fmt.Println("Error! expected count not equal to found count")
	}
}

func Perft(depth int64, moveCounter *int64, s *State) {
	if depth != 0 {
		genTimer.Start()
		moves := s.quickGenMoves()
		genTimer.Stop()
		for _, move := range *moves {
			makeTimer.Start()
			s.MakeMove(move)
			makeTimer.Stop()
			Perft(depth-1, moveCounter, s)
			unMakeTimer.Start()
			s.UnMakeMove(move)
			unMakeTimer.Stop()
		}
	} else {
		*moveCounter++
	}
}
