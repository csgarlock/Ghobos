package main

import (
	"fmt"
	"time"
)

type HistoryTable [12][64]uint64

const (
	min32 int32 = -2147483646
	max32 int32 = 2147483647

	LowestEval  int32 = -2147483646 + CentiPawn - 2 // The lowest 32 bit value such that the 16 least significant bits are all 0
	highestEval int32 = 2147483646 - CentiPawn + 2  // The largest 32 bit value such that the 16 least significant bits are all 0

	startingEval int32 = CentiPawn * 30

	startingAsperationWindowOffset int32 = CentiPawn * 10

	asperationMateSearchCutoff int32 = CentiPawn * 2000
)

var nodesSearched uint64 = 0

var startingDepth int32 = 0

var lastMoveScore int32 = startingEval

var historyTable HistoryTable = HistoryTable{}

func (s *State) IterativeDeepiningSearch(maxTime time.Duration) Move {
	startTime := time.Now()
	result, found := transpositionTable.SearchState(s)
	stateEvalGuess := lastMoveScore
	if found {
		stateEvalGuess = EvalLowToHigh(result.eval)
	}
	aspirationWindowLow := stateEvalGuess - startingAsperationWindowOffset
	aspirationWindowHigh := stateEvalGuess + startingAsperationWindowOffset
	aspirationDelta := startingAsperationWindowOffset
	bestFoundMove := Move(0)
	currentDepth := int32(1)
	contendingMove := NilMove
	stateScore := stateEvalGuess
	lastSearchNodes := uint64(1)
	for time.Since(startTime) < maxTime {
		fmt.Printf("Searching next depth with window [%f, %f]\n", NormalizeEval(aspirationWindowLow), NormalizeEval(aspirationWindowHigh))
		startingDepth = currentDepth
		nodesSearched = 0
		stateScore, contendingMove = s.NegaMax(currentDepth, aspirationWindowLow, aspirationWindowHigh)
		effectiveBranchFactor := float64(nodesSearched) / float64(lastSearchNodes)
		fmt.Printf("Searched to Depth: %d, Best Move: %s, Score: %.2f, EBF: %.2f\n", currentDepth, contendingMove.ShortString(), NormalizeEval(stateScore), effectiveBranchFactor)
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
			lastSearchNodes = nodesSearched
		}
	}
	fmt.Println("Excpected Moves: ", s.getPV())
	return bestFoundMove
}

func (s *State) NegaMax(depth int32, alpha int32, beta int32) (int32, Move) {
	nodesSearched++
	result, found := transpositionTable.SearchState(s)
	if found {
		ttEval := EvalLowToHigh(result.eval)
		_, ttNodeType := result.depthAndNode.parseDepthandNode()
		if ttNodeType == TerminalNode {
			return ttEval, NilMove
		}
	}
	if depth == 0 {
		return s.QuiescenceSearch(alpha, beta)
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
		score, _ := s.NegaMax(depth-1, -beta, -alpha)
		score *= -1
		s.UnMakeMove(move)
		if score >= beta {
			transpositionTable.AddState(s, beta, move, uint16(startingDepth)-uint16(depth), CutNode)
			friendPiece := s.board.getColorPieceAt(move.OriginSquare(), s.turn)
			enemyPiece := s.board.getColorPieceAt(move.DestinationSquare(), 1-s.turn)
			if enemyPiece == NoPiece {
				historyTable[friendPiece][move.DestinationSquare()] += uint64(depth * depth)
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

func (s *State) QuiescenceSearch(alpha int32, beta int32) (int32, Move) {
	nodesSearched++
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
		score, _ := s.QuiescenceSearch(-beta, -alpha)
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
