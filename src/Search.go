package main

import "fmt"

const (
	LowestEval  int32 = -2147483646 + CentiPawn - 2 // The lowest 32 bit value such that the 16 least significant bits are all 0
	highestEval int32 = 2147483646 - CentiPawn + 2  // The largest 32 bit value such that the 16 least significant bits are all 0
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
		fmt.Printf("Move %d: %v Searched\nScore: %.2f\nRemaining: %d\n", i, move.ShortString(), NormalizeEval(moveEval), len(*moves)-i-1)
	}
	fmt.Println("Best Move:", bestMove.ShortString())
	return bestMove
}

func (s *State) NegaMax(depth int32, alpha int32, beta int32, nodesSearched *int32) int32 {
	*nodesSearched++
	result, found := transpositionTable.SearchState(s)
	if found {
		ttEval := EvalLowToHigh(result.eval)
		ttdepth, ttNodeType := result.depthAndNode.parseDepthandNode()
		if ttNodeType == TerminalNode {
			return ttEval
		}
		if ttdepth >= uint16(depth) {
			if ttNodeType == pVNode {
				if ttEval < alpha {
					return alpha
				} else if ttEval >= beta {
					return beta
				} else {
					return ttEval
				}
			} else if ttNodeType == CutNode {
				if ttEval >= beta {
					return beta
				}
			} else if ttNodeType == AllNode {
				if ttEval <= alpha {
					return alpha
				}
			}
		}
	}
	if depth == 0 {
		return s.QuiescenceSearch(alpha, beta, nodesSearched)
	}
	moves := s.genAllMoves(true)
	if len(*moves) == 0 {
		if s.check {
			eval := LowestEval + (((startingDepth - depth) / 2) * CentiPawn)
			transpositionTable.AddState(s, eval, 0, uint16(startingDepth)-uint16(depth), TerminalNode)
			return eval
		} else {
			transpositionTable.AddState(s, 0, 0, uint16(startingDepth)-uint16(depth), TerminalNode)
			return 0
		}
	}
	allNode := true
	bestScore := LowestEval
	bestMove := (*moves)[0]
	for _, move := range *moves {
		s.MakeMove(move)
		score := -s.NegaMax(depth-1, -beta, -alpha, nodesSearched)
		s.UnMakeMove(move)
		if score >= beta {
			transpositionTable.AddState(s, beta, move, uint16(startingDepth)-uint16(depth), CutNode)
			return beta
		}
		if score > alpha {
			allNode = false
			alpha = score
		}
		if score > bestScore {
			bestMove = move
			bestScore = score
		}
	}
	if allNode {
		transpositionTable.AddState(s, alpha, bestMove, uint16(startingDepth)-uint16(depth), AllNode)
	} else {
		transpositionTable.AddState(s, alpha, bestMove, uint16(startingDepth)-uint16(depth), pVNode)
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
