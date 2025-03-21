package main

import (
	"fmt"
	"math"
)

const (
	CentiPawn   int32 = 65536
	PawnValue   int32 = CentiPawn * 100
	BishopValue int32 = CentiPawn * 325
	KnightValue int32 = CentiPawn * 300
	RookValue   int32 = CentiPawn * 500
	QueenValue  int32 = CentiPawn * 900
	KingValue   int32 = CentiPawn * 1000

	DoublePawnValue   int32 = CentiPawn * -25
	PassedPawnValue   int32 = CentiPawn * 50
	IsolatedPawnValue int32 = CentiPawn * -25

	OpenFileRookValue int32 = CentiPawn * 10

	MobilityValue int32 = CentiPawn * 1

	PawnKingAttackValue   int32 = 1
	KnightKingAttackValue int32 = 2
	BishopKingAttackValue int32 = 2
	RookKingAttackValue   int32 = 3
	QueenKingAttackValue  int32 = 5
	KingKingAttackValue   int32 = 1

	PawnPhaseValue   int8 = 0
	BishopPhaseValue int8 = 1
	KnightPhaseValue int8 = 1
	RookPhaseValue   int8 = 2
	QueenPhaseValue  int8 = 4
	KingPhaseValue   int8 = 0
	TotalPhaseValue  int8 = PawnPhaseValue*16 + BishopPhaseValue*4 + KnightPhaseValue*4 + RookPhaseValue*4 + QueenPhaseValue*2 + KingPhaseValue*2
)

var valueTable [6]int32 = [6]int32{KingValue, QueenValue, RookValue, BishopValue, KnightValue, PawnValue}

var midgamePieceSquareTable [12][64]int32
var endgamePieceSquareTable [12][64]int32

var kingSafetyTable [60]uint16
var gamePhaseValues [6]int8 = [6]int8{KingPhaseValue, QueenPhaseValue, RookPhaseValue, BishopPhaseValue, KingPhaseValue, PawnPhaseValue}

func InitializeEvalVariables() {
	setupEvalValues()
}

func (s *State) EvalState(perspective uint8) int32 {
	var eval int32 = 0
	var phaseValue int8 = 0
	var mgPieceSquareTableEval int32 = 0
	var egPieceSquareTableEval int32 = 0
	for i := 0; i < 6; i++ {
		whiteCount := BitCount(s.board[i])
		blackCount := BitCount(s.board[6+i])
		eval += int32(whiteCount) * valueTable[i]
		eval -= int32(blackCount) * valueTable[i]
		phaseValue += int8(whiteCount+blackCount) * gamePhaseValues[i]
		whiteBoard := s.board[i]
		for whiteBoard != 0 {
			square := PopLSB(&whiteBoard)
			mgPieceSquareTableEval += midgamePieceSquareTable[i][square]
			egPieceSquareTableEval += endgamePieceSquareTable[i][square]
		}
		blackBoard := s.board[i+6]
		for blackBoard != 0 {
			square := PopLSB(&blackBoard)
			mgPieceSquareTableEval -= midgamePieceSquareTable[i+6][square]
			egPieceSquareTableEval -= endgamePieceSquareTable[i+6][square]
		}
	}
	// Adjust piece square values based off of phase of the game
	mgPhaseValue := min(phaseValue, TotalPhaseValue)
	egPhaseValue := TotalPhaseValue - mgPhaseValue
	eval += (mgPieceSquareTableEval*int32(mgPhaseValue) + egPieceSquareTableEval*int32(egPhaseValue)) / int32(TotalPhaseValue)

	// King Safety and Piece Mobility. Allows pieces to "move" to square occupied by friendly pieces because defending a friendly piece is still beneficial
	for side := uint8(0); side < 2; side++ {
		s.ensurePins(side)
		mobilityCount := int32(0)
		kingAttackerPoints := int32(0)
		friendIndex := side * 6
		enemyIndex := (1 - side) * 6
		friendKingSquare := GetLSB(s.board[friendIndex+King])
		enemyKingNeighbors := (moveBoards[King][GetLSB(s.board[enemyIndex+King])] & ^s.sideOccupied[1-side]) | s.board[enemyIndex+King]
		bishopBoard := s.board[friendIndex+Bishop]
		for bishopBoard != 0 {
			bishopSquare := PopLSB(&bishopBoard)
			safeSquares := s.getPinBoard(bishopSquare, friendKingSquare, side)
			bishopMoves := getBishopMoves(bishopSquare, s.occupied) & safeSquares
			mobilityCount += int32(BitCount(bishopMoves))
			kingAttackerPoints += int32(BitCount(bishopMoves&enemyKingNeighbors)) * BishopKingAttackValue
		}
		knightBoard := s.board[friendIndex+Knight]
		for knightBoard != 0 {
			knightSquare := PopLSB(&knightBoard)
			safeSquares := s.getPinBoard(knightSquare, friendKingSquare, side)
			knightMoves := moveBoards[Knight][knightSquare] & safeSquares
			mobilityCount += int32(BitCount(knightMoves))
			kingAttackerPoints += int32(BitCount(knightMoves&enemyKingNeighbors)) * KnightKingAttackValue
		}
		rookBoard := s.board[friendIndex+Rook]
		for rookBoard != 0 {
			rookSquare := PopLSB(&rookBoard)
			safeSquares := s.getPinBoard(rookSquare, friendKingSquare, side)
			rookMoves := getRookMoves(rookSquare, s.occupied) & safeSquares
			mobilityCount += int32(BitCount(rookMoves))
			kingAttackerPoints += int32(BitCount(rookMoves&enemyKingNeighbors)) * RookKingAttackValue
		}
		queenBoard := s.board[friendIndex+Queen]
		for queenBoard != 0 {
			queenSquare := PopLSB(&queenBoard)
			safeSquares := s.getPinBoard(queenSquare, friendKingSquare, side)
			queenMoves := getQueenMoves(queenSquare, s.occupied) & safeSquares
			mobilityCount += int32(BitCount(queenMoves))
			kingAttackerPoints += int32(BitCount(queenMoves&enemyKingNeighbors)) * QueenKingAttackValue
		}
		pawnBoard := s.board[friendIndex+Pawn]
		moveStep := -Step(16*side - 8) // Turns 0 into 8 for upstep and 1 int -8 for down step
		homeRank := int8(side*5 + 1)   // Turns 0 into 1, turns 1 into 6
		for pawnBoard != 0 {
			pawnSquare := PopLSB(&pawnBoard)
			safeSquare := s.getPinBoard(pawnSquare, friendKingSquare, side)
			pawnMoves := GetPawnMoves(pawnSquare, s.occupied, moveStep, homeRank) & safeSquare
			pawnAttacks := pawnAttackBoards[side][pawnSquare] & s.sideOccupied[1-side]
			mobilityCount += int32(BitCount(pawnMoves | pawnAttacks))
			kingAttackerPoints += int32(BitCount(pawnAttacks&enemyKingNeighbors)) * PawnKingAttackValue
		}
		kingBoard := s.board[friendIndex+King]
		for knightBoard != 0 {
			kingSquare := PopLSB(&kingBoard)
			kingMoves := moveBoards[King][kingSquare]
			mobilityCount += int32(BitCount(kingMoves))
			kingAttackerPoints += int32(BitCount(kingMoves&enemyKingNeighbors)) * KingKingAttackValue
		}
		if side == White {
			eval += mobilityCount * MobilityValue
			eval += getKingSafetyValue(kingAttackerPoints)
		} else {
			eval -= mobilityCount * MobilityValue
			eval -= getKingSafetyValue(kingAttackerPoints)
		}
	}

	// Pawn Eval
	whitePawns := s.board[WhitePawn]
	shallowestWhiteNeighbor := [8]int8{7, 7, 7, 7, 7, 7, 7, 7}
	neighboringWhitePawns := [8]int8{}
	whiteFileCount := [8]uint8{}
	blackPawns := s.board[BlackPawn]
	shallowestBlackNeighbor := [8]int8{0, 0, 0, 0, 0, 0, 0, 0}
	neighboringBlackPawns := [8]int8{}
	blackFileCount := [8]uint8{}
	openFiles := [8]bool{true, true, true, true, true, true, true, true}
	for whitePawns != 0 {
		square := PopLSB(&whitePawns)
		file := square.File()
		whiteFileCount[file]++
		openFiles[file] = false
		rank := square.Rank()
		if rank < shallowestWhiteNeighbor[file] {
			shallowestWhiteNeighbor[file] = rank
		}
		if file != 0 {
			neighboringWhitePawns[file-1]++
			if rank < shallowestWhiteNeighbor[file-1] {
				shallowestWhiteNeighbor[file-1] = rank
			}
		}
		if file != 7 {
			neighboringWhitePawns[file+1]++
			if rank < shallowestWhiteNeighbor[file+1] {
				shallowestWhiteNeighbor[file+1] = rank
			}
		}
	}
	for blackPawns != 0 {
		square := PopLSB(&blackPawns)
		file := square.File()
		blackFileCount[file]++
		openFiles[file] = false
		rank := square.Rank()
		if rank > shallowestBlackNeighbor[file] {
			shallowestBlackNeighbor[file] = rank
		}
		if file != 0 {
			neighboringBlackPawns[file-1]++
			if rank > shallowestBlackNeighbor[file-1] {
				shallowestBlackNeighbor[file-1] = rank
			}
		}
		if file != 7 {
			neighboringBlackPawns[file+1]++
			if rank > shallowestBlackNeighbor[file+1] {
				shallowestBlackNeighbor[file+1] = rank
			}
		}
	}
	whitePawns = s.board[WhitePawn]
	whiteDoubled := 0
	whiteIsolated := 0
	whitePassed := 0
	blackDoubled := 0
	blackIsolated := 0
	blackPassed := 0
	blackPawns = s.board[BlackPawn]
	for whitePawns != 0 {
		square := PopLSB(&whitePawns)
		file := square.File()
		rank := square.Rank()
		if whiteFileCount[file] > 1 {
			eval += DoublePawnValue
			whiteDoubled++
		}
		if neighboringWhitePawns[file] == 0 {
			eval += IsolatedPawnValue
			whiteIsolated++
		}
		if (neighboringBlackPawns[file] == 0 && blackFileCount[file] == 0) || rank >= shallowestBlackNeighbor[file] {
			eval += DoublePawnValue
			whitePassed++
		}
	}
	for blackPawns != 0 {
		square := PopLSB(&blackPawns)
		file := square.File()
		rank := square.Rank()
		if blackFileCount[file] > 1 {
			eval -= DoublePawnValue
			blackDoubled++
		}
		if neighboringBlackPawns[file] == 0 {
			eval -= IsolatedPawnValue
			blackIsolated++
		}
		if (neighboringWhitePawns[file] == 0 && whiteFileCount[file] == 0) || rank <= shallowestWhiteNeighbor[file] {
			eval -= DoublePawnValue
			blackPassed++
		}
	}

	// Open File Rook Bonus
	whiteRooks := s.board[WhiteRook]
	blackRooks := s.board[BlackRook]
	for whiteRooks != 0 {
		square := PopLSB(&whiteRooks)
		file := square.File()
		if openFiles[file] {
			eval += OpenFileRookValue
		}
	}
	for blackRooks != 0 {
		square := PopLSB(&blackRooks)
		file := square.File()
		if openFiles[file] {
			eval -= OpenFileRookValue
		}
	}
	if perspective == White {
		return eval
	}
	return -eval
}

func (s *State) NormalizedEval(perspective uint8) float64 {
	rawEval := s.EvalState(perspective)
	return NormalizeEval(rawEval)
}

func NormalizeEval(rawEval int32) float64 {
	centiEval := rawEval / CentiPawn
	return float64(centiEval) / 100.0
}

func prettyEval(rawEval int32, perspective uint8) string {
	mateSign := int32(1)
	if rawEval < 0 {
		mateSign = -1
	}
	absRawEval := rawEval * mateSign
	if absRawEval > mateValueCutoff {
		mateDepth := highestEval - absRawEval
		if (mateSign == 1 && perspective == White) || (mateSign == -1 && perspective == Black) {
			return fmt.Sprintf("+M%d", mateDepth)
		} else {
			return fmt.Sprintf("-M%d", mateDepth)
		}
	}
	return fmt.Sprintf("%.2f", NormalizeEval(rawEval))
}

func EvalHighToLow(eval int32) int16 {
	return int16(eval / CentiPawn)
}

func EvalLowToHigh(eval int16) int32 {
	return int32(eval) * CentiPawn
}

func setupEvalValues() {
	for i := 0; i < len(kingSafetyTable); i++ {
		kingSafetyTable[i] = kingSafetyFunction(float64(i))
	}
	midgamePieceSquareTable[WhiteKing] = [64]int32{
		20, 30, 10, 0, 0, 10, 30, 20,
		20, 20, 0, 0, 0, 0, 20, 20,
		-10, -20, -20, -20, -20, -20, -20, -10,
		-20, -30, -30, -40, -40, -30, -30, -20,
		-30, -40, -40, -50, -50, -40, -40, -30,
		-30, -40, -40, -50, -50, -40, -40, -30,
		-30, -40, -40, -50, -50, -40, -40, -30,
		-30, -40, -40, -50, -50, -40, -40, -30}
	endgamePieceSquareTable[WhiteKing] = [64]int32{
		-50, -30, -30, -30, -30, -30, -30, -50,
		-30, -30, 0, 0, 0, 0, -30, -30,
		-30, -10, 20, 30, 30, 20, -10, -30,
		-30, -10, 30, 40, 40, 30, -10, -30,
		-30, -10, 30, 40, 40, 30, -10, -30,
		-30, -10, 20, 30, 30, 20, -10, -30,
		-30, -20, -10, 0, 0, -10, -20, -30,
		-50, -40, -30, -20, -20, -30, -40, -50}
	midgamePieceSquareTable[WhiteQueen] = [64]int32{
		-20, -10, -10, -5, -5, -10, -10, -20,
		-10, 0, 5, 0, 0, 0, 0, -10,
		-10, 5, 5, 5, 5, 5, 0, -10,
		0, 0, 5, 5, 5, 5, 0, -5,
		-5, 0, 5, 5, 5, 5, 0, -5,
		-10, 0, 5, 5, 5, 5, 0, -10,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-20, -10, -10, -5, -5, -10, -10, -20}
	endgamePieceSquareTable[WhiteQueen] = [64]int32{
		-20, -10, -10, -5, -5, -10, -10, -20,
		-10, 0, 5, 0, 0, 0, 0, -10,
		-10, 5, 5, 5, 5, 5, 0, -10,
		0, 0, 5, 5, 5, 5, 0, -5,
		-5, 0, 5, 5, 5, 5, 0, -5,
		-10, 0, 5, 5, 5, 5, 0, -10,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-20, -10, -10, -5, -5, -10, -10, -20}
	midgamePieceSquareTable[WhiteRook] = [64]int32{
		0, 0, 0, 5, 5, 0, 0, 0,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		5, 10, 10, 10, 10, 10, 10, 5,
		0, 0, 0, 0, 0, 0, 0, 0}
	endgamePieceSquareTable[WhiteRook] = [64]int32{
		0, 0, 0, 5, 5, 0, 0, 0,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		5, 10, 10, 10, 10, 10, 10, 5,
		0, 0, 0, 0, 0, 0, 0, 0}
	midgamePieceSquareTable[WhiteBishop] = [64]int32{
		-20, -10, -10, -10, -10, -10, -10, -20,
		-10, 5, 0, 0, 0, 0, 5, -10,
		-10, 10, 10, 10, 10, 10, 10, -10,
		-10, 0, 10, 10, 10, 10, 0, -10,
		-10, 5, 5, 10, 10, 5, 5, -10,
		-10, 0, 5, 10, 10, 5, 0, -10,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-20, -10, -10, -10, -10, -10, -10, -20}
	endgamePieceSquareTable[WhiteBishop] = [64]int32{
		-20, -10, -10, -10, -10, -10, -10, -20,
		-10, 5, 0, 0, 0, 0, 5, -10,
		-10, 10, 10, 10, 10, 10, 10, -10,
		-10, 0, 10, 10, 10, 10, 0, -10,
		-10, 5, 5, 10, 10, 5, 5, -10,
		-10, 0, 5, 10, 10, 5, 0, -10,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-20, -10, -10, -10, -10, -10, -10, -20}
	midgamePieceSquareTable[WhiteKnight] = [64]int32{
		-50, -40, -30, -30, -30, -30, -40, -50,
		-40, -20, 0, 5, 5, 0, -20, -40,
		-30, 5, 10, 15, 15, 10, 5, -30,
		-30, 0, 15, 20, 20, 15, 0, -30,
		-30, 5, 15, 20, 20, 15, 5, -30,
		-30, 0, 10, 15, 15, 10, 0, -30,
		-40, -20, 0, 0, 0, 0, -20, -40,
		-50, -40, -30, -30, -30, -30, -40, -50}
	endgamePieceSquareTable[WhiteKnight] = [64]int32{
		-50, -40, -30, -30, -30, -30, -40, -50,
		-40, -20, 0, 5, 5, 0, -20, -40,
		-30, 5, 10, 15, 15, 10, 5, -30,
		-30, 0, 15, 20, 20, 15, 0, -30,
		-30, 5, 15, 20, 20, 15, 5, -30,
		-30, 0, 10, 15, 15, 10, 0, -30,
		-40, -20, 0, 0, 0, 0, -20, -40,
		-50, -40, -30, -30, -30, -30, -40, -50}
	midgamePieceSquareTable[WhitePawn] = [64]int32{
		0, 0, 0, 0, 0, 0, 0, 0,
		5, 10, 10, -20, -20, 10, 10, 5,
		5, -5, -10, 0, 0, -10, -5, 5,
		0, 0, 0, 20, 20, 0, 0, 0,
		5, 5, 10, 25, 25, 10, 5, 5,
		10, 10, 20, 30, 30, 20, 10, 10,
		50, 50, 50, 50, 50, 50, 50, 50,
		0, 0, 0, 0, 0, 0, 0, 0}
	endgamePieceSquareTable[WhitePawn] = [64]int32{
		0, 0, 0, 0, 0, 0, 0, 0,
		-30, -30, -30, -30, -30, -30, -30, -30,
		-10, -10, -10, -10, -10, -10, -10, -10,
		0, 0, 0, 0, 0, 0, 0, 0,
		20, 20, 20, 20, 20, 20, 20, 20,
		40, 40, 40, 40, 40, 40, 40, 40,
		60, 60, 60, 60, 60, 60, 60, 60,
		0, 0, 0, 0, 0, 0, 0, 0}
	//Normalize tables to centipawn value
	for i := 0; i < 6; i++ {
		for s := 0; s < 64; s++ {
			midgamePieceSquareTable[i][s] *= CentiPawn
			endgamePieceSquareTable[i][s] *= CentiPawn
		}
	}
	//Reflect along x axis for black boards
	for b := 0; b < 6; b++ {
		for s := 0; s < 64; s++ {
			// xoring a square by 56 flips it over the x axis between the 4 and 5 rank
			midgamePieceSquareTable[b+6][s^56] = midgamePieceSquareTable[b][s]
			endgamePieceSquareTable[b+6][s^56] = endgamePieceSquareTable[b][s]
		}
	}
}

func kingSafetyFunction(x float64) uint16 {
	const a = float64(-4e-5)
	const b = float64(2.7e-3)
	const c = float64(0.088)
	const d = float64(0.70)
	const e = float64(-0.8)
	return uint16(math.Floor(max(0, a*math.Pow(x, 4)+b*math.Pow(x, 3)+c*math.Pow(x, 2)+d*x+e)))

}

func getKingSafetyValue(x int32) int32 {
	return CentiPawn * int32(kingSafetyTable[min(x, 59)])
}
