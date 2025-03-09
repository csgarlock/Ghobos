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

	mateValueCutoff int32 = CentiPawn * 30000

	NullMoveReduction = 2

	FutilityCutoff int32 = CentiPawn * 300
)

var nodesSearched uint64 = 0

var lastMoveScore int32 = startingEval

var historyTable HistoryTable = HistoryTable{}

func (s *State) IterativeDeepiningSearch(maxTime time.Duration, debugPrint bool) Move {
	totalNodes := uint64(0)
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
		if stateScore > mateValueCutoff {
			aspirationWindowLow = mateValueCutoff
			aspirationWindowHigh = max32
		} else if stateScore < -mateValueCutoff {
			aspirationWindowLow = min32
			aspirationWindowHigh = -mateValueCutoff
		}
		if debugPrint {
			fmt.Println("Search time left: ", maxTime-time.Since(startTime))
			fmt.Printf("Searching next depth with window [%f, %f]\n", NormalizeEval(aspirationWindowLow), NormalizeEval(aspirationWindowHigh))
		}
		nodesSearched = 0
		stateScore, contendingMove = s.NegaMax(currentDepth, aspirationWindowLow, aspirationWindowHigh, true, true, true)
		totalNodes += nodesSearched
		if debugPrint {
			effectiveBranchFactor := float64(nodesSearched) / float64(lastSearchNodes)
			fmt.Printf("Searched to Depth: %d, Best Move: %s, Score: %s, EBF: %.2f\n", currentDepth, contendingMove.ShortString(), prettyEval(stateScore, s.turn), effectiveBranchFactor)
		}
		// Check if returned score was at bounds of aspiration window
		if stateScore == aspirationWindowLow {
			if debugPrint {
				fmt.Println("Searched Failed Low")
			}
			aspirationWindowHigh -= aspirationDelta / 3
			if aspirationWindowLow <= -asperationMateSearchCutoff {
				aspirationWindowLow = min32
			} else {
				aspirationWindowLow -= aspirationDelta
				aspirationDelta *= 2
			}
		} else if stateScore == aspirationWindowHigh {
			if debugPrint {
				fmt.Println("Searched Failed High")
			}
			aspirationWindowLow += aspirationDelta / 3
			if aspirationWindowHigh >= asperationMateSearchCutoff {
				aspirationWindowHigh = max32
			} else {
				aspirationWindowHigh += aspirationDelta
				aspirationDelta *= 2
			}
		} else {
			if debugPrint {
				fmt.Println("Searched Succeeded")
			}
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
	fmt.Println("Best Move:", bestFoundMove.ShortString())
	fmt.Println("Move Evaluation:", prettyEval(stateScore, s.turn))
	fmt.Println("Expected Moves:", s.getPV())
	fmt.Println("Total Nodes Searched:", totalNodes)
	fmt.Println("Total Search Time:", time.Since(startTime))
	fmt.Printf("Million Nodes per Second: %.2f\n", float64(totalNodes)/time.Since(startTime).Seconds()/1_000_000.0)
	return bestFoundMove
}

func (s *State) NegaMax(depth int32, alpha int32, beta int32, skipIID bool, skipNull bool, forceSearch bool) (int32, Move) {
	s.searchParameters.trueDepth += 1
	nodesSearched++
	if s.lastCapOrPawn >= 100 || s.repetitionMap.get(s.hashcode) >= 3 {
		s.searchParameters.trueDepth--
		return 0, NilMove
	}
	result, found := transpositionTable.SearchState(s)
	projectedBestMove := NilMove
	if found {
		ttEval := EvalLowToHigh(result.eval)
		ttDepth, ttNodeType := result.depthAndNode.parseDepthandNode()
		if ttNodeType == TerminalNode {
			s.searchParameters.trueDepth--
			if ttEval == mateTranspositionValue {
				return clampInt32(LowestEval+int32(s.searchParameters.trueDepth), alpha, beta), NilMove
			} else {
				return clampInt32(0, alpha, beta), NilMove
			}
		}
		projectedBestMove = result.bestMove
		if !forceSearch && ttDepth >= uint16(depth) {
			if ttNodeType == AllNode && ttEval <= alpha {
				s.searchParameters.trueDepth--
				return alpha, projectedBestMove
			} else if ttNodeType == CutNode && ttEval >= beta {
				s.searchParameters.trueDepth--
				return beta, projectedBestMove
			} else if ttNodeType == pVNode && ttEval >= alpha && ttEval <= beta {
				s.searchParameters.trueDepth--
				return ttEval, projectedBestMove
			}
		}
	} else if depth > 5 && !skipIID {
		_, projectedBestMove = s.NegaMax(depth/2, alpha, beta, true, true, true)
	}
	if depth == 0 {
		s.searchParameters.trueDepth--
		qScore, qMove := s.QuiescenceSearch(alpha, beta)
		return qScore, qMove
	}
	futileNode := false
	if depth == 1 && !s.check && alpha > -asperationMateSearchCutoff && beta < asperationMateSearchCutoff {
		staticEval := s.EvalState(s.turn)
		if staticEval < alpha-FutilityCutoff {
			s.genAllMoves(false)
			futileNode = true
		} else {
			s.genAllMoves(true)
		}
	} else {
		s.genAllMoves(true)
	}
	if captureMoves.len() == 0 && quietMoves.len() == 0 && !futileNode {
		if s.check {
			eval := LowestEval + int32(s.searchParameters.trueDepth)
			transpositionTable.AddState(s, mateTranspositionValue, NilMove, uint16(depth), TerminalNode)
			s.searchParameters.trueDepth--
			return clampInt32(eval, alpha, beta), NilMove
		} else {
			transpositionTable.AddState(s, stalemateTranpositionValue, NilMove, uint16(depth), TerminalNode)
			s.searchParameters.trueDepth--
			return clampInt32(0, alpha, beta), NilMove
		}
	}
	var moves []Move
	if !futileNode {
		moves = s.orderMoves(projectedBestMove)
	} else {
		moves = s.orderCaptureMoves()
		if len(moves) == 0 {
			s.searchParameters.trueDepth--
			return alpha, NilMove
		}
	}
	if !s.check && depth > 2 && !skipNull {
		friendIndex := s.turn * 6
		hasBigPiece := false
		for i := uint8(1); i < 5; i++ {
			if s.board[friendIndex+i] != 0 {
				hasBigPiece = true
				break
			}
		}
		if hasBigPiece {
			s.MakeMove(PassingMove)
			score, _ := s.NegaMax(max(depth-NullMoveReduction-1, 1), -beta, -beta+1, false, false, false)
			score *= -1
			s.UnMakeMove(PassingMove)
			if score >= beta {
				s.searchParameters.trueDepth--
				return beta, PassingMove
			}
		}
	}
	allNode := true
	bestMove := moves[0]
	for i, move := range moves {
		// Reductions
		reduction := int32(0)
		// Late Move Reduction
		if depth > 2 {
			if i > 3 {
				reduction += 1
			}
		}
		s.MakeMove(move)
		score := int32(0)
		if i == 0 {
			score, _ = s.NegaMax(max(depth-reduction-1, 0), -beta, -alpha, false, false, false)
			score *= -1
		} else {
			score, _ = s.NegaMax(max(depth-reduction-1, 0), -alpha-1, -alpha, false, false, false)
			score *= -1
			if score > alpha && beta-alpha > 1 {
				score, _ = s.NegaMax(depth-1, -beta, -alpha, false, false, false)
				score *= -1
			}
		}
		s.UnMakeMove(move)
		if score >= beta {
			transpositionTable.AddState(s, beta, move, uint16(depth), CutNode)
			friendPiece := s.board.getColorPieceAt(move.OriginSquare(), s.turn)
			enemyPiece := s.board.getColorPieceAt(move.DestinationSquare(), 1-s.turn)
			if enemyPiece == NoPiece {
				historyTable[friendPiece][move.DestinationSquare()] += uint64(depth * depth)
				s.addKiller(move)
			}
			s.searchParameters.trueDepth--
			return beta, move
		}
		if score > alpha {
			allNode = false
			alpha = score
			bestMove = move
		}
	}
	if allNode {
		transpositionTable.AddState(s, alpha, bestMove, uint16(depth), AllNode)
	} else {
		transpositionTable.AddState(s, alpha, bestMove, uint16(depth), pVNode)
	}
	s.searchParameters.trueDepth--
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
	s.genAllMoves(false)
	captureMoves.sort()
	moves := make([]Move, captureMoves.len())
	for i, capture := range captureMoves.slice[0:captureMoves.len()] {
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

func (s *State) orderMoves(ttMove Move) []Move {
	sortedMoves := make([]Move, captureMoves.len()+quietMoves.len())
	totalIndex := 0
	captureMoves.sort()
	quietMoves.sort()
	if ttMove != NilMove && ttMove != PassingMove {
		sortedMoves[0] = ttMove
		totalIndex++
	}
	badCutoff := captureMoves.len()
	for i := uint16(0); i < captureMoves.len(); i++ {
		if captureMoves.slice[i].captureValue >= 0 && captureMoves.slice[i].move != ttMove {
			sortedMoves[totalIndex] = captureMoves.slice[i].move
			totalIndex++
		} else {
			badCutoff = i
			break
		}
	}
	skipIndex := [2]int{-1, -1}
	if int16(len(s.searchParameters.killerTable)) > s.searchParameters.trueDepth {
		for i := 0; i < int(quietMoves.len()); i++ {
			if quietMoves.slice[i].move == s.searchParameters.killerTable[s.searchParameters.trueDepth][0] && quietMoves.slice[i].move != ttMove {
				sortedMoves[totalIndex] = s.searchParameters.killerTable[s.searchParameters.trueDepth][0]
				skipIndex[0] = i
				totalIndex++
			} else if quietMoves.slice[i].move == s.searchParameters.killerTable[s.searchParameters.trueDepth][1] && quietMoves.slice[i].move != ttMove {
				sortedMoves[totalIndex] = s.searchParameters.killerTable[s.searchParameters.trueDepth][1]
				skipIndex[1] = i
				totalIndex++
			}
		}
	}
	for i := 0; i < int(quietMoves.len()); i++ {
		if i != skipIndex[0] && i != skipIndex[1] && quietMoves.slice[i].move != ttMove {
			sortedMoves[totalIndex] = quietMoves.slice[i].move
			totalIndex++
		}
	}
	for i := badCutoff; i < captureMoves.len(); i++ {
		if captureMoves.slice[i].move != ttMove {
			sortedMoves[totalIndex] = captureMoves.slice[i].move
			totalIndex++
		}
	}
	return sortedMoves
}

func (s *State) orderCaptureMoves() []Move {
	sortedMoves := make([]Move, captureMoves.len())
	captureMoves.sort()
	for i, capture := range captureMoves.slice[0:captureMoves.len()] {
		sortedMoves[i] = capture.move
	}
	return sortedMoves

}

func (s *State) addKiller(move Move) {
	if s.searchParameters.trueDepth > int16(len(s.searchParameters.killerTable)-1) {
		killerTable := make(KillerTable, len(s.searchParameters.killerTable)*2)
		for i := range len(s.searchParameters.killerTable) {
			killerTable[i][0] = s.searchParameters.killerTable[i][0]
			killerTable[i][1] = s.searchParameters.killerTable[i][1]
			killerTable[len(s.searchParameters.killerTable)+i][0] = NilMove
			killerTable[len(s.searchParameters.killerTable)+i][1] = NilMove
		}
		s.searchParameters.killerTable = killerTable
	}
	s.searchParameters.killerTable[s.searchParameters.trueDepth][1] = s.searchParameters.killerTable[s.searchParameters.trueDepth][0]
	s.searchParameters.killerTable[s.searchParameters.trueDepth][0] = move
}
