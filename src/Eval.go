package main

const (
	CentiPawn   int32 = 65536
	PawnValue   int32 = CentiPawn * 100
	BishopValue int32 = CentiPawn * 310
	KnightValue int32 = CentiPawn * 300
	RookValue   int32 = CentiPawn * 500
	QueenValue  int32 = CentiPawn * 900
	KingValue   int32 = 0
)

var valueTable [6]int32 = [6]int32{KingValue, QueenValue, RookValue, BishopValue, KnightValue, PawnValue}
var pieceSquareTable [12][64]int32

func InitializeEvalVariables() {
	setupPieceSquareTable()
}

func (s *State) EvalState(perspective uint8) int32 {
	var eval int32 = 0
	for i := 0; i < 6; i++ {
		eval += int32(BitCount(s.board[i])) * valueTable[i]
		eval -= int32(BitCount(s.board[6+i])) * valueTable[i]
		whiteBoard := s.board[i]
		for whiteBoard != 0 {
			eval += pieceSquareTable[i][PopLSB(&whiteBoard)]
		}
		blackBoard := s.board[i+6]
		for blackBoard != 0 {
			eval -= pieceSquareTable[i+6][PopLSB(&blackBoard)]
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

func EvalHighToLow(eval int32) int16 {
	return int16(eval / CentiPawn)
}

func EvalLowToHigh(eval int16) int32 {
	return int32(eval) * CentiPawn
}

func setupPieceSquareTable() {
	pieceSquareTable[WhiteKing] = [64]int32{
		20, 30, 10, 0, 0, 10, 30, 20,
		20, 20, 0, 0, 0, 0, 20, 20,
		-10, -20, -20, -20, -20, -20, -20, -10,
		-20, -30, -30, -40, -40, -30, -30, -20,
		-30, -40, -40, -50, -50, -40, -40, -30,
		-30, -40, -40, -50, -50, -40, -40, -30,
		-30, -40, -40, -50, -50, -40, -40, -30,
		-30, -40, -40, -50, -50, -40, -40, -30}
	pieceSquareTable[WhiteQueen] = [64]int32{
		-20, -10, -10, -5, -5, -10, -10, -20,
		-10, 0, 5, 0, 0, 0, 0, -10,
		-10, 5, 5, 5, 5, 5, 0, -10,
		0, 0, 5, 5, 5, 5, 0, -5,
		-5, 0, 5, 5, 5, 5, 0, -5,
		-10, 0, 5, 5, 5, 5, 0, -10,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-20, -10, -10, -5, -5, -10, -10, -20}
	pieceSquareTable[WhiteRook] = [64]int32{
		0, 0, 0, 5, 5, 0, 0, 0,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		5, 10, 10, 10, 10, 10, 10, 5,
		0, 0, 0, 0, 0, 0, 0, 0}
	pieceSquareTable[WhiteBishop] = [64]int32{
		-20, -10, -10, -10, -10, -10, -10, -20,
		-10, 5, 0, 0, 0, 0, 5, -10,
		-10, 10, 10, 10, 10, 10, 10, -10,
		-10, 0, 10, 10, 10, 10, 0, -10,
		-10, 5, 5, 10, 10, 5, 5, -10,
		-10, 0, 5, 10, 10, 5, 0, -10,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-20, -10, -10, -10, -10, -10, -10, -20}
	pieceSquareTable[WhiteKnight] = [64]int32{
		-50, -40, -30, -30, -30, -30, -40, -50,
		-40, -20, 0, 5, 5, 0, -20, -40,
		-30, 5, 10, 15, 15, 10, 5, -30,
		-30, 0, 15, 20, 20, 15, 0, -30,
		-30, 5, 15, 20, 20, 15, 5, -30,
		-30, 0, 10, 15, 15, 10, 0, -30,
		-40, -20, 0, 0, 0, 0, -20, -40,
		-50, -40, -30, -30, -30, -30, -40, -50}
	pieceSquareTable[WhitePawn] = [64]int32{
		0, 0, 0, 0, 0, 0, 0, 0,
		5, 10, 10, -20, -20, 10, 10, 5,
		5, -5, -10, 0, 0, -10, -5, 5,
		0, 0, 0, 20, 20, 0, 0, 0,
		5, 5, 10, 25, 25, 10, 5, 5,
		10, 10, 20, 30, 30, 20, 10, 10,
		50, 50, 50, 50, 50, 50, 50, 50,
		0, 0, 0, 0, 0, 0, 0, 0}
	//Normalize tables to centipawn value
	for i := 0; i < 6; i++ {
		for s := 0; s < 64; s++ {
			pieceSquareTable[i][s] *= CentiPawn
		}
	}
	//Reflect along x axis for black boards
	for b := 0; b < 6; b++ {
		for s := 0; s < 64; s++ {
			// xoring a square by 56 flips it over the x axis between the 4 and 5 rank
			pieceSquareTable[b+6][s^56] = pieceSquareTable[b][s]
		}
	}
}
