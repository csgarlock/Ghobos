package main

const (
	CentiPawn   int32 = 32768 // 2 ^ 15
	PawnValue   int32 = CentiPawn * 100
	BishopValue int32 = PawnValue * 3
	KnightValue int32 = PawnValue * 3
	RookValue   int32 = PawnValue * 5
	QueenValue  int32 = PawnValue * 9
	KingValue   int32 = 0
)

var valueTable [6]int32 = [6]int32{KingValue, QueenValue, RookValue, BishopValue, KnightValue, PawnValue}

func (s *State) EvalState(perspective uint8) int32 {
	var eval int32 = 0
	for i := 0; i < 6; i++ {
		eval += int32(BitCount(s.board[i])) * valueTable[i]
		eval -= int32(BitCount(s.board[6+i])) * valueTable[i]
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
