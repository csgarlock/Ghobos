package main

import (
	"fmt"
	"time"
)

const (
	min32 int32 = -2147483646
	max32 int32 = 2147483647

	LowestEval  int32 = -2147483646 + CentiPawn - 2 // The lowest 32 bit value such that the 16 least significant bits are all 0
	highestEval int32 = 2147483646 - CentiPawn + 2  // The largest 32 bit value such that the 16 least significant bits are all 0

	startingEval int32 = CentiPawn * 30

	startingAsperationWindowOffset int32 = CentiPawn * 10

	asperationMateSearchCutoff int32 = CentiPawn * 2000
)

var startingDepth int32 = 0

var lastMoveScore int32 = startingEval

func (s *State) IterativeDeepiningSearch(maxTime time.Duration, nodesSearched *int32) Move {
	startTime := time.Now()
	result, found := transpositionTable.SearchState(s)
	var stateEvalGuess int32
	if found {
		stateEvalGuess = EvalLowToHigh(result.eval)
	} else {
		stateEvalGuess = lastMoveScore
	}
	var aspirationWindowLow int32 = stateEvalGuess - startingAsperationWindowOffset
	var aspirationWindowHigh int32 = stateEvalGuess + startingAsperationWindowOffset
	var aspirationDelta int32 = startingAsperationWindowOffset
	var bestFoundMove Move = 0
	var currentDepth int32 = 1
	var contendingMove Move
	stateScore := stateEvalGuess
	for time.Since(startTime) < maxTime {
		fmt.Printf("Searching next depth with window [%f, %f]\n", NormalizeEval(aspirationWindowLow), NormalizeEval(aspirationWindowHigh))
		startingDepth = currentDepth
		stateScore, contendingMove = s.NegaMax(currentDepth, aspirationWindowLow, aspirationWindowHigh, nodesSearched)
		fmt.Println("Searched to Depth:", currentDepth, ", Best Move:", contendingMove.ShortString(), " Value:", NormalizeEval(stateScore))
		// Check if returned score was at bounds of aspiration window
		if stateScore == aspirationWindowLow {
			fmt.Println("Searched Failed Low")
			aspirationWindowHigh -= aspirationDelta / 2
			if aspirationWindowLow <= -asperationMateSearchCutoff {
				aspirationWindowLow = min32
			} else {
				aspirationWindowLow -= aspirationDelta
				aspirationDelta *= 2
			}
		} else if stateScore == aspirationWindowHigh {
			fmt.Println("Searched Failed High")
			aspirationWindowLow += aspirationDelta / 2
			if aspirationWindowHigh >= asperationMateSearchCutoff {
				aspirationWindowHigh = max32
			} else {
				aspirationWindowHigh += aspirationDelta
				aspirationDelta *= 2
			}
		} else {
			fmt.Println("Searched Succeeded")
			bestFoundMove = contendingMove
			lastMoveScore = stateScore
			if currentDepth >= 5 {
				aspirationDelta = startingAsperationWindowOffset / 2
			} else {
				aspirationDelta = startingAsperationWindowOffset
			}
			aspirationWindowLow = stateScore - aspirationDelta
			aspirationWindowHigh = stateScore + aspirationDelta
			currentDepth += 1
		}
	}
	fmt.Println("Excpected Moves: ", s.getPV())
	return bestFoundMove
}

func (s *State) NegaMax(depth int32, alpha int32, beta int32, nodesSearched *int32) (int32, Move) {
	*nodesSearched++
	result, found := transpositionTable.SearchState(s)
	if found {
		ttEval := EvalLowToHigh(result.eval)
		_, ttNodeType := result.depthAndNode.parseDepthandNode()
		if ttNodeType == TerminalNode {
			return ttEval, NilMove
		}
	}
	if depth == 0 {
		return s.QuiescenceSearch(alpha, beta, nodesSearched)
	}
	moves := s.genAllMoves(true)
	if len(*moves) == 0 {
		if s.check {
			eval := LowestEval + (((startingDepth - depth) / 2) * CentiPawn)
			transpositionTable.AddState(s, eval, NilMove, uint16(startingDepth)-uint16(depth), TerminalNode)
			return eval, NilMove
		} else {
			transpositionTable.AddState(s, 0, NilMove, uint16(startingDepth)-uint16(depth), TerminalNode)
			return 0, NilMove
		}
	}
	// Swap the best move from the tt with the first move
	if found {
		ttBestMove := result.bestMove
		for i := range *moves {
			if (*moves)[i] == ttBestMove {
				firstMove := (*moves)[0]
				(*moves)[0] = (*moves)[i]
				(*moves)[i] = firstMove
			}
		}
	}
	allNode := true
	bestMove := (*moves)[0]
	for _, move := range *moves {
		s.MakeMove(move)
		score, _ := s.NegaMax(depth-1, -beta, -alpha, nodesSearched)
		score *= -1
		s.UnMakeMove(move)
		if score >= beta {
			transpositionTable.AddState(s, beta, move, uint16(startingDepth)-uint16(depth), CutNode)
			friendPiece := s.board.getColorPieceAt(move.OriginSquare(), s.turn)
			enemyPiece := s.board.getColorPieceAt(move.DestinationSquare(), 1-s.turn)
			if enemyPiece == NoPiece {
				s.historyTable[friendPiece][move.DestinationSquare()] += uint64(depth * depth)
			}
			return beta, move
		}
		if score > alpha {
			allNode = false
			alpha = score
			bestMove = move
		}
	}
	if allNode {
		transpositionTable.AddState(s, alpha, bestMove, uint16(startingDepth)-uint16(depth), AllNode)
	} else {
		transpositionTable.AddState(s, alpha, bestMove, uint16(startingDepth)-uint16(depth), pVNode)
	}
	return alpha, bestMove
}

func (s *State) QuiescenceSearch(alpha int32, beta int32, nodesSearched *int32) (int32, Move) {
	*nodesSearched++
	standingPat := s.EvalState(s.turn)
	if standingPat >= beta {
		return beta, NilMove
	}
	if alpha < standingPat {
		alpha = standingPat
	}
	moves := s.genAllMoves(false)
	bestMove := NilMove
	for _, move := range *moves {
		s.MakeMove(move)
		score, _ := s.QuiescenceSearch(-beta, -alpha, nodesSearched)
		score *= -1
		s.UnMakeMove(move)

		if score >= beta {
			return beta, move
		}
		if score > alpha {
			alpha = score
			bestMove = move
		}
	}
	return alpha, bestMove
}
