package main

import "fmt"

const (
	LowestEval  int32 = -2147483647
	highestEval int32 = 2147483647
)

var startingDepth int32 = 0

func (s *State) getBestMove(depth int32, nodesSearched *int32) Move {
	startingDepth = depth
	*nodesSearched++
	moves := s.genAllMoves(true)
	bestMove := (*moves)[0]
	alpha := LowestEval
	for i, move := range *moves {
		s.MakeMove(move)
		moveEval := -s.NegaMax(depth-1, LowestEval, highestEval, nodesSearched)
		s.UnMakeMove(move)
		if moveEval > alpha {
			bestMove = move
			alpha = moveEval
		}
		fmt.Printf("Move %d: %v Searched\nScore: %.2f\nRemaining: %d\n", i, move.ShortString(), normalizeEval(moveEval), len(*moves)-i-1)
	}
	return bestMove
}

func (s *State) NegaMax(depth int32, alpha int32, beta int32, nodesSearched *int32) int32 {
	*nodesSearched++
	if depth == 0 {
		return s.QuiescenceSearch(alpha, beta, nodesSearched)
	}
	moves := s.genAllMoves(true)
	if len(*moves) == 0 && s.check {
		return LowestEval + ((startingDepth - depth) / 2)
	}
	for _, move := range *moves {
		s.MakeMove(move)
		score := -s.NegaMax(depth-1, -beta, -alpha, nodesSearched)
		s.UnMakeMove(move)
		if score >= beta {
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}
	return alpha
}

func (s *State) QuiescenceSearch(alpha int32, beta int32, nodesSearched *int32) int32 {
	*nodesSearched++
	standingPat := s.EvalState(s.turn)
	if standingPat >= beta {
		return beta
	}
	if alpha < standingPat {
		alpha = standingPat
	}
	moves := s.genAllMoves(false)
	for _, move := range *moves {
		s.MakeMove(move)
		score := -s.QuiescenceSearch(-beta, -alpha, nodesSearched)
		s.UnMakeMove(move)

		if score >= beta {
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}
	return alpha
}
