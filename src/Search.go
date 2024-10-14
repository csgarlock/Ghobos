package main

import (
	"fmt"
	"sort"
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
		stateScore, contendingMove = s.NegaMax(currentDepth, aspirationWindowLow, aspirationWindowHigh, true)
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

func (s *State) NegaMax(depth int32, alpha int32, beta int32, skipIID bool) (int32, Move) {
	s.searchParameters.trueDepth += 1
	nodesSearched++
	result, found := transpositionTable.SearchState(s)
	projectedBestMove := NilMove
	if found {
		ttEval := EvalLowToHigh(result.eval)
		_, ttNodeType := result.depthAndNode.parseDepthandNode()
		if ttNodeType == TerminalNode {
			s.searchParameters.trueDepth -= 1
			return ttEval, NilMove
		}
		projectedBestMove = result.bestMove
	} else if depth > 5 && !skipIID {
		_, projectedBestMove = s.NegaMax(depth/2, alpha, beta, true)
	}
	if depth == 0 {
		s.searchParameters.trueDepth -= 1
		return s.QuiescenceSearch(alpha, beta)
	}
	captures, quiets := s.genAllMoves(true)
	if len(*captures) == 0 && len(*quiets) == 0 {
		if s.check {
			eval := LowestEval + (((startingDepth - depth) / 2) * CentiPawn)
			transpositionTable.AddState(s, eval, NilMove, uint16(startingDepth)-uint16(depth), TerminalNode)
			s.searchParameters.trueDepth -= 1
			return eval, NilMove
		} else {
			transpositionTable.AddState(s, 0, NilMove, uint16(startingDepth)-uint16(depth), TerminalNode)
			s.searchParameters.trueDepth -= 1
			return 0, NilMove
		}
	}
	moves := s.orderMoves(captures, quiets, projectedBestMove)
	allNode := true
	bestMove := (*moves)[0]
	for _, move := range *moves {
		s.MakeMove(move)
		score, _ := s.NegaMax(depth-1, -beta, -alpha, false)
		score *= -1
		s.UnMakeMove(move)
		if score >= beta {
			transpositionTable.AddState(s, beta, move, uint16(startingDepth)-uint16(depth), CutNode)
			friendPiece := s.board.getColorPieceAt(move.OriginSquare(), s.turn)
			enemyPiece := s.board.getColorPieceAt(move.DestinationSquare(), 1-s.turn)
			if enemyPiece == NoPiece {
				historyTable[friendPiece][move.DestinationSquare()] += uint64(depth * depth)
				s.addKiller(move)
			}
			s.searchParameters.trueDepth -= 1
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
	s.searchParameters.trueDepth -= 1
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
	captures, _ := s.genAllMoves(false)
	sort.Slice(*captures, func(i, j int) bool {
		return (*captures)[i].captureValue > (*captures)[j].captureValue
	})
	moves := make([]Move, len(*captures))
	for i, capture := range *captures {
		moves[i] = capture.move
	}
	bestMove := NilMove
	for _, move := range moves {
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

func (s *State) orderMoves(captures *[]CaptureMove, quiets *[]QuietMove, ttMove Move) *[]Move {
	sortedMoves := make([]Move, len(*captures)+len(*quiets))
	totalIndex := 0
	sort.Slice(*captures, func(i, j int) bool {
		return (*captures)[i].captureValue > (*captures)[j].captureValue
	})
	sort.Slice(*quiets, func(i, j int) bool {
		return (*quiets)[i].historyValue > (*quiets)[j].historyValue
	})
	if ttMove != NilMove {
		sortedMoves[0] = ttMove
		totalIndex++
	}
	badCutoff := len(*captures)
	for i := 0; i < len(*captures); i++ {
		if (*captures)[i].captureValue >= 0 && (*captures)[i].move != ttMove {
			sortedMoves[totalIndex] = (*captures)[i].move
			totalIndex++
		} else {
			badCutoff = i
			break
		}
	}
	skipIndex := [2]int{-1, -1}
	if int16(len(*s.searchParameters.killerTable)) > s.searchParameters.trueDepth {
		for i := 0; i < len(*quiets); i++ {
			if (*quiets)[i].move == (*s.searchParameters.killerTable)[s.searchParameters.trueDepth][0] && (*quiets)[i].move != ttMove {
				sortedMoves[totalIndex] = (*s.searchParameters.killerTable)[s.searchParameters.trueDepth][0]
				skipIndex[0] = i
				totalIndex++
			} else if (*quiets)[i].move == (*s.searchParameters.killerTable)[s.searchParameters.trueDepth][1] && (*quiets)[i].move != ttMove {
				sortedMoves[totalIndex] = (*s.searchParameters.killerTable)[s.searchParameters.trueDepth][1]
				skipIndex[1] = i
				totalIndex++
			}
		}
	}
	for i := 0; i < len(*quiets); i++ {
		if i != skipIndex[0] && i != skipIndex[1] && (*quiets)[i].move != ttMove {
			sortedMoves[totalIndex] = (*quiets)[i].move
			totalIndex++
		}
	}
	for i := badCutoff; i < len(*captures); i++ {
		if (*captures)[i].move != ttMove {
			sortedMoves[totalIndex] = (*captures)[i].move
			totalIndex++
		}
	}
	return &sortedMoves
}

func (s *State) addKiller(move Move) {
	if s.searchParameters.trueDepth > int16(len(*s.searchParameters.killerTable)-1) {
		killerTable := make(KillerTable, len(*s.searchParameters.killerTable)*2)
		for i := range len(*s.searchParameters.killerTable) {
			killerTable[i][0] = (*s.searchParameters.killerTable)[i][0]
			killerTable[i][1] = (*s.searchParameters.killerTable)[i][1]
			killerTable[len(*s.searchParameters.killerTable)+i][0] = NilMove
			killerTable[len(*s.searchParameters.killerTable)+i][1] = NilMove
		}
		s.searchParameters.killerTable = &killerTable
	}
	(*s.searchParameters.killerTable)[s.searchParameters.trueDepth][1] = (*s.searchParameters.killerTable)[s.searchParameters.trueDepth][0]
	(*s.searchParameters.killerTable)[s.searchParameters.trueDepth][0] = move
}
