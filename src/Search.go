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
		contendingMove, stateScore = s.getBestMove(currentDepth, nodesSearched, aspirationWindowLow, aspirationWindowHigh)
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

func (s *State) getBestMove(depth int32, nodesSearched *int32, aspirationLow int32, aspirationHigh int32) (Move, int32) {
	startingDepth = depth
	*nodesSearched++
	moves := s.genAllMoves(true)
	result, found := transpositionTable.SearchState(s)
	// Swap the best best move from the tt with the first move
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
	bestMove := (*moves)[0]
	alpha := aspirationLow
	startCheck := s.check
	for _, move := range *moves {
		s.MakeMove(move)
		moveEval := -s.NegaMax(depth-1, -aspirationHigh, -alpha, nodesSearched)
		s.UnMakeMove(move)
		if moveEval > alpha {
			bestMove = move
			alpha = moveEval
		}
		//fmt.Printf("Move %d: %v Searched\nScore: %.2f\nRemaining: %d\n", i, move.ShortString(), NormalizeEval(moveEval), len(*moves)-i-1)
	}
	s.check = startCheck
	fmt.Println("Searched to Depth:", depth, ", Best Move:", bestMove.ShortString(), " Value:", NormalizeEval(alpha))
	return bestMove, alpha
}

func (s *State) NegaMax(depth int32, alpha int32, beta int32, nodesSearched *int32) int32 {
	*nodesSearched++
	result, found := transpositionTable.SearchState(s)
	if found {
		ttEval := EvalLowToHigh(result.eval)
		_, ttNodeType := result.depthAndNode.parseDepthandNode()
		if ttNodeType == TerminalNode {
			return ttEval
		}
		/*
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
		*/
	}
	if depth == 0 {
		return s.QuiescenceSearch(alpha, beta, nodesSearched)
	}
	moves := s.genAllMoves(true)
	if len(*moves) == 0 {
		if s.check {
			eval := LowestEval + (((startingDepth - depth) / 2) * CentiPawn)
			transpositionTable.AddState(s, eval, NilMove, uint16(startingDepth)-uint16(depth), TerminalNode)
			return eval
		} else {
			transpositionTable.AddState(s, 0, NilMove, uint16(startingDepth)-uint16(depth), TerminalNode)
			return 0
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
	bestScore := LowestEval
	bestMove := (*moves)[0]
	for _, move := range *moves {
		s.MakeMove(move)
		score := -s.NegaMax(depth-1, -beta, -alpha, nodesSearched)
		s.UnMakeMove(move)
		if score >= beta {
			transpositionTable.AddState(s, beta, move, uint16(startingDepth)-uint16(depth), CutNode)
			friendPiece := s.board.getColorPieceAt(move.OriginSquare(), s.turn)
			enemyPiece := s.board.getColorPieceAt(move.DestinationSquare(), 1-s.turn)
			if enemyPiece == NoPiece {
				s.historyTable[friendPiece][move.DestinationSquare()] += uint64(depth * depth)
			}
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
