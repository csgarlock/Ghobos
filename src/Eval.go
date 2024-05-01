package main

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
	return normalizeEval(rawEval)
}

func normalizeEval(rawEval int32) float64 {
	centiEval := rawEval / CentiPawn
	return float64(centiEval) / 100.0
}
