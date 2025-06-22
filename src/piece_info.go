package main

type Step int8

const (
	RightStep     Step = 1
	UpRightStep   Step = 9
	UpStep        Step = 8
	UpLeftStep    Step = 7
	LeftStep      Step = -1
	DownLeftStep  Step = -9
	DownStep      Step = -8
	DownRightStep Step = -7

	KnightStepRightUp   Step = 10
	KnightStepUpRight   Step = 17
	KnightStepUpLeft    Step = 15
	KnightStepLeftUp    Step = 6
	KnightStepLeftDown  Step = -10
	KnightStepDownLeft  Step = -17
	KnightStepDownRight Step = -15
	KnightStepRightDown Step = -6

	RightStepID     = 0
	UpRightStepID   = 1
	UpStepID        = 2
	UpLeftStepID    = 3
	LeftStepID      = 4
	DownLeftStepID  = 5
	DownStepID      = 6
	DownRightStepID = 7

	King   = 0
	Queen  = 1
	Rook   = 2
	Bishop = 3
	Knight = 4
	Pawn   = 5

	WhiteKing   = 0
	WhiteQueen  = 1
	WhiteRook   = 2
	WhiteBishop = 3
	WhiteKnight = 4
	WhitePawn   = 5

	BlackKing   = 6
	BlackQueen  = 7
	BlackRook   = 8
	BlackBishop = 9
	BlackKnight = 10
	BlackPawn   = 11

	NoPiece = 12
)

type Piece interface {
	// Params:
	// Square: square the piece is on
	// Bitboard: occupied squares
	getMoveBitboard(Square, Bitboard) Bitboard
}

type KingTemplate struct{}
type QueenTemplate struct{}
type RookTemplate struct{}
type BishopTemplate struct{}
type KnightTemplate struct{}
type PawnTemplate struct{}

var stepMap map[Step]int = map[Step]int{RightStep: 0, UpRightStep: 1, UpStep: 2, UpLeftStep: 3, LeftStep: 4, DownLeftStep: 5, DownStep: 6, DownRightStep: 7, KnightStepRightUp: 8, KnightStepUpRight: 9, KnightStepUpLeft: 10, KnightStepLeftUp: 11, KnightStepLeftDown: 12, KnightStepDownLeft: 13, KnightStepDownRight: 14, KnightStepRightDown: 15}
var stepIds [19]uint8 = [19]uint8{}

var allSteps [16]Step = [16]Step{RightStep, UpRightStep, UpStep, UpLeftStep, LeftStep, DownLeftStep, DownStep, DownRightStep, KnightStepRightUp, KnightStepUpRight, KnightStepUpLeft, KnightStepLeftUp, KnightStepLeftDown, KnightStepDownLeft, KnightStepDownRight, KnightStepRightDown}
var cardinalSteps [8]Step = [8]Step{RightStep, UpRightStep, UpStep, UpLeftStep, LeftStep, DownLeftStep, DownStep, DownRightStep}

var bishopSteps [4]Step = [4]Step{UpRightStep, UpLeftStep, DownRightStep, DownLeftStep}
var rookSteps [4]Step = [4]Step{RightStep, UpStep, LeftStep, DownStep}
var knightSteps [8]Step = [8]Step{KnightStepRightUp, KnightStepUpRight, KnightStepUpLeft, KnightStepLeftUp, KnightStepLeftDown, KnightStepDownLeft, KnightStepDownRight, KnightStepRightDown}

var squareToSquareStep [64][64]Step = [64][64]Step{}

var stepBoards [16][64]bool = [16][64]bool{}

var moveBoards [5][64]Bitboard = [5][64]Bitboard{}
var pawnAttackBoards [2][64]Bitboard = [2][64]Bitboard{}

// Can the piece do the slide (canSlide[pieceID][stepID])
var canSlide [12][8]bool = [12][8]bool{}

var kingInstance KingTemplate = KingTemplate{}
var queenInstance QueenTemplate = QueenTemplate{}
var rookInstance RookTemplate = RookTemplate{}
var bishopInstance BishopTemplate = BishopTemplate{}
var knightInstance KnightTemplate = KnightTemplate{}
var _ PawnTemplate = PawnTemplate{}

func (KingTemplate) getMoveBitboard(square Square, _ Bitboard) Bitboard {
	return moveBoards[King][square]
}

func (QueenTemplate) getMoveBitboard(square Square, occupied Bitboard) Bitboard {
	return getQueenMoves(square, occupied)
}

func (RookTemplate) getMoveBitboard(square Square, occupied Bitboard) Bitboard {
	return getRookMoves(square, occupied)
}

func (BishopTemplate) getMoveBitboard(square Square, occupied Bitboard) Bitboard {
	return getBishopMoves(square, occupied)
}

func (KnightTemplate) getMoveBitboard(square Square, _ Bitboard) Bitboard {
	return moveBoards[Knight][square]
}

// DO NOT USE
func (PawnTemplate) getMoveBitboard(square Square, _ Bitboard) Bitboard {
	return EmptyBitboard
}

func (s *State) genAllPawnMoves(captures bool, mask Bitboard, moveList *MoveList) {
	friendIndex := 6 * s.turn
	enemyIndex := 6 * (1 - s.turn)
	if captures {
		pawnBoard := s.board.uGet(friendIndex + Pawn)
		if s.canEnpassant && (s.checkInfo.checkers == 0 || s.checkInfo.enPassantBlockingBitboard != EmptyBitboard) {
			enPawnBoard := pawnAttackBoards[1-s.turn][s.enPassantSquare] & pawnBoard
			if enPawnBoard != EmptyBitboard {
				eRank := Rank3 << (Bitboard(1-s.turn) * 8)
				rookSliders := s.board.uGet(enemyIndex+Queen) | s.board.uGet(enemyIndex+Rook)
				kingBoard := s.board.uGet(friendIndex + King)
				shouldSafetyCheck := (kingBoard&eRank != EmptyBitboard) && (rookSliders&eRank != EmptyBitboard)
				for enPawnBoard != EmptyBitboard {
					pawnSquare := PopLSB(&enPawnBoard)
					if shouldSafetyCheck {
						modifiedOccupancy := s.occupied & ^(boardFromSquare(pawnSquare) | boardFromSquare(s.enPassantSquare.Step(-16*Step(1-s.turn)+8)))
						kingSquare := GetLSB(kingBoard)
						rookSliderCopy := rookSliders
						for rookSliderCopy != EmptyBitboard {
							rookSquare := PopLSB(&rookSliderCopy)
							if squareToSquareFillBoards[kingSquare][rookSquare]&modifiedOccupancy != EmptyBitboard {
								moveList.addMove(BuildMove(pawnSquare, s.enPassantSquare, 0, EnPassantSpacialMove))
							}
						}
					} else {
						moveList.addMove(BuildMove(pawnSquare, s.enPassantSquare, 0, EnPassantSpacialMove))
					}
				}
			}
		}
		for pawnBoard != EmptyBitboard {
			pawnSquare := PopLSB(&pawnBoard)
			attackBoard := pawnAttackBoards[s.turn][pawnSquare] & mask
			for attackBoard != EmptyBitboard {
				desSquare := PopLSB(&attackBoard)
				BuildPawnMoves(pawnSquare, desSquare, moveList)
			}
		}
	} else {
		pawnBoard := s.board.uGet(friendIndex + Pawn)
		if s.turn == White {
			if pawnBoard&Rank1 != EmptyBitboard {
				genPawnDoublePushes(pawnBoard&Rank1, s.notOccupied, mask, White, moveList)
			}
			pawnBoard &= ^Rank1
			for pawnBoard != EmptyBitboard {
				pawnSquare := PopLSB(&pawnBoard)
				genPawnSinglePushes(pawnSquare, s.notOccupied, mask, White, moveList)
			}
		} else {
			if pawnBoard&Rank6 != EmptyBitboard {
				genPawnDoublePushes(pawnBoard&Rank6, s.notOccupied, mask, Black, moveList)
			}
			pawnBoard &= ^Rank6
			for pawnBoard != EmptyBitboard {
				pawnSquare := PopLSB(&pawnBoard)
				genPawnSinglePushes(pawnSquare, s.notOccupied, mask, Black, moveList)
			}
		}
	}
}

func genPawnSinglePushes(pawnSquare Square, notOccupied Bitboard, mask Bitboard, turn uint8, moveList *MoveList) {
	if turn == White {
		desBoard := boardFromSquare(pawnSquare) << 8 & notOccupied & mask
		if desBoard != EmptyBitboard {
			BuildPawnMoves(pawnSquare, GetLSB(desBoard), moveList)
		}
	} else {
		desBoard := boardFromSquare(pawnSquare) >> 8 & notOccupied & mask
		if desBoard != EmptyBitboard {
			BuildPawnMoves(pawnSquare, GetLSB(desBoard), moveList)
		}
	}
}

func genPawnDoublePushes(pawns Bitboard, notOccupied Bitboard, mask Bitboard, turn uint8, moveList *MoveList) {
	validBoard := EmptyBitboard
	if turn == White {
		validBoard = Rank2 & notOccupied
		validBoard |= (validBoard << 8) & notOccupied
	} else {
		validBoard = Rank5 & notOccupied
		validBoard |= (validBoard >> 8) & notOccupied
	}
	validBoard &= mask
	for pawns != EmptyBitboard {
		pawnSquare := PopLSB(&pawns)
		pawnMoves := files[pawnSquare.File()] & validBoard
		for pawnMoves != EmptyBitboard {
			desSquare := PopLSB(&pawnMoves)
			moveList.addMove(BuildSimpleMove(pawnSquare, desSquare))
		}
	}
}

func (s *State) genKingMoves(captures bool, mask Bitboard, moveList *MoveList) {
	friendIndex := 6 * s.turn
	kingBoard := s.board.uGet(friendIndex + King)
	kingSquare := GetLSB(kingBoard)
	// Temporarily remove king from occupied to check move safety
	s.occupied &= ^kingBoard
	moveBoard := kingInstance.getMoveBitboard(kingSquare, EmptyBitboard) & mask
	for moveBoard != EmptyBitboard {
		desSquare := PopLSB(&moveBoard)
		if s.isSquareSafeEasy(desSquare) {
			moveList.addMove(BuildSimpleMove(kingSquare, desSquare))
		}
	}
	s.occupied |= kingBoard
	if !captures && !s.check {
		rankIndex := s.turn * 56
		if s.castleAvailability[s.turn] {
			if s.occupied&Bitboard(0x60<<rankIndex) == EmptyBitboard && s.board.uGet(friendIndex+Rook)&Bitboard(0x80<<rankIndex) != EmptyBitboard {
				desSquare := kingSquare + 2
				if s.isSquareSafeEasy(5+Square(rankIndex)) && s.isSquareSafeEasy(6+Square(rankIndex)) {
					moveList.addMove(BuildMove(kingSquare, desSquare, 0, CastleSpecialMove))
				}
			}
		}
		if s.castleAvailability[s.turn+2] {
			if s.occupied&Bitboard(0xE<<rankIndex) == EmptyBitboard && s.board.uGet(friendIndex+Rook)&Bitboard(0x1<<rankIndex) != EmptyBitboard {
				desSquare := kingSquare - 2
				if s.isSquareSafeEasy(3+Square(rankIndex)) && s.isSquareSafeEasy(2+Square(rankIndex)) {
					moveList.addMove(BuildMove(kingSquare, desSquare, 0, CastleSpecialMove))
				}
			}
		}
	}
}

func InitializeMoveBoards() {
	InitializeStepBoard()
	FillSlidingAttacks(&bishopSteps, &moveBoards[Bishop])
	FillSlidingAttacks(&rookSteps, &moveBoards[Rook])
	InitializeMagics()
	var square Square
	for square = 0; square < 64; square++ {
		var bitboard Bitboard = EmptyBitboard
		for _, step := range cardinalSteps {
			if square.tryStep(step) {
				bitboard |= 1 << square.Step(step)
			}
		}
		moveBoards[King][square] = bitboard
		bitboard = EmptyBitboard
		for _, step := range knightSteps {
			if square.tryStep(step) {
				bitboard |= 1 << square.Step(step)
			}
		}
		moveBoards[Knight][square] = bitboard
		moveBoards[Queen][square] = moveBoards[Bishop][square] | moveBoards[Rook][square]
		var whitePawnAttacks Bitboard = 0
		var blackPawnAttacks Bitboard = 0
		if square.tryStep(UpLeftStep) {
			whitePawnAttacks |= (1 << square.Step(UpLeftStep))
		}
		if square.tryStep(UpRightStep) {
			whitePawnAttacks |= (1 << square.Step(UpRightStep))
		}
		if square.tryStep(DownLeftStep) {
			blackPawnAttacks |= (1 << square.Step(DownLeftStep))
		}
		if square.tryStep(DownRightStep) {
			blackPawnAttacks |= (1 << square.Step(DownRightStep))
		}
		pawnAttackBoards[White][square] = whitePawnAttacks
		pawnAttackBoards[Black][square] = blackPawnAttacks
	}
	for i := uint8(0); i < 12; i++ {
		if isSlider(i) {
			if i%6 == Queen {
				canSlide[i] = [8]bool{true, true, true, true, true, true, true, true}
			} else if i%6 == Rook {
				canSlide[i] = [8]bool{true, false, true, false, true, false, true, false}
			} else if i%6 == Bishop {
				canSlide[i] = [8]bool{false, true, false, true, false, true, false, true}
			}
		}
	}
}

func InitializeStepBoard() {
	for i, step := range allSteps {
		center := Square(35)
		centerStep := center.Step(step)
		rankDiff := centerStep.Rank() - center.Rank()
		fileDiff := centerStep.File() - center.File()
		var square Square
		for square = 0; square < 64; square++ {
			squareStep := square.Step(step)
			if squareStep.Rank()-square.Rank() == rankDiff && squareStep.File()-square.File() == fileDiff {
				stepBoards[i][square] = true
			} else {
				stepBoards[i][square] = false
			}
		}
	}
	for i := Square(0); i < 64; i++ {
		for j := Square(0); j < 64; j++ {
			if i != j {
				rankDiff := j.Rank() - i.Rank()
				fileDiff := j.File() - i.File()
				if rankDiff == 0 {
					if fileDiff > 0 {
						squareToSquareStep[i][j] = RightStep
					} else {
						squareToSquareStep[i][j] = LeftStep
					}
				} else if fileDiff == 0 {
					if rankDiff > 0 {
						squareToSquareStep[i][j] = UpStep
					} else {
						squareToSquareStep[i][j] = DownStep
					}
				} else if rankDiff == fileDiff {
					if rankDiff > 0 {
						squareToSquareStep[i][j] = UpRightStep
					} else {
						squareToSquareStep[i][j] = DownLeftStep
					}
				} else if rankDiff == -fileDiff {
					if rankDiff < 0 {
						squareToSquareStep[i][j] = DownRightStep
					} else {
						squareToSquareStep[i][j] = UpLeftStep
					}
				}
			}
		}
	}
	stepIds[RightStep+9] = RightStepID
	stepIds[UpRightStep+9] = UpRightStepID
	stepIds[UpStep+9] = UpStepID
	stepIds[UpLeftStep+9] = UpLeftStepID
	stepIds[LeftStep+9] = LeftStepID
	stepIds[DownLeftStep+9] = DownLeftStepID
	stepIds[DownStep+9] = DownStepID
	stepIds[DownRightStep+9] = DownRightStepID
}

func FillSlidingAttacks(steps *[4]Step, resultBitboards *[64]Bitboard) {
	var square Square
	for _, step := range steps {
		for square = 0; square < 64; square++ {
			var stepSquare Square = square
			for stepSquare.tryStep(step) {
				stepSquare = stepSquare.Step(step)
				resultBitboards[square] |= 1 << stepSquare
			}
		}
	}
}

func FindBlockedSlidingAttack(square Square, steps *[4]Step, occupied Bitboard) Bitboard {
	var result Bitboard = 0
	if (1<<square)&occupied != 0 {
		occupied = occupied ^ (1 << square)
	}
	for _, step := range steps {
		var stepSquare Square = square
		for stepSquare.tryStep(step) && ((1<<stepSquare)&occupied == 0) {
			stepSquare = stepSquare.Step(step)
			result |= 1 << stepSquare
		}
	}
	return result
}

func GetPawnMoves(square Square, occupied Bitboard, moveStep Step, homeRank int8) Bitboard {
	var resultBoard Bitboard = EmptyBitboard
	if square.Rank() == homeRank {
		square = square.Step(moveStep)
		if occupied&(1<<Bitboard(square)) == 0 {
			resultBoard |= (1 << Bitboard(square))
		} else {
			return resultBoard
		}
		square = square.Step(moveStep)
		if occupied&(1<<Bitboard(square)) == 0 {
			resultBoard |= (1 << Bitboard(square))
		}
		return resultBoard
	}
	square = square.Step(moveStep)
	if occupied&(1<<Bitboard(square)) == 0 {
		resultBoard |= (1 << Bitboard(square))
	}
	return resultBoard
}

func isSlider(id uint8) bool {
	pieceId := id % 6
	if pieceId > King && pieceId < Knight {
		return true
	}
	return false
}

func getStepId(step Step) uint8 {
	return stepIds[step+9]
}
