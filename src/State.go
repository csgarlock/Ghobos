package main

import (
	"fmt"
	"strconv"
	"strings"
)

// 0 = White King Side, 1 = Black King Side, 2 = White Queen Side, 3 = Black Queen Side
type CastleAvailability [4]bool
type State struct {
	board                  *Board
	turn                   uint8 // 0 for White, 1 for Black
	enPassantSquare        Square
	check                  bool
	captureHistory         *CaptureHistory
	canEnpassant           bool
	enPassantSquareHistory *EnPassantSquareHistory
	ply                    uint16
	castleAvailability     *CastleAvailability
	castleHistory          *CastleHistory
}

// Stores important information about the current move being generated to more easily pass to functions
type MoveGenInfo struct {
	includeQuiets bool
	enemyBoard    Bitboard
	occupied      Bitboard
	notOccupied   Bitboard
	popFunction   func(*Bitboard) Square
}

type SafetyCheckBoards struct {
	bishopsBoards Bitboard
	rooksBoard    Bitboard
	knightsBoard  Bitboard
	kingBoard     Bitboard
	pawnsBoard    Bitboard
	combinedBoard Bitboard
}

type PinSafety struct {
	key         Square
	safeSquares Bitboard
}

func (s *State) MakeMove(move Move, checkCounter *int64, enPassantCounter *int64, castleCounter *int64) {
	var friendIndex uint8 = s.turn * 6
	var enemyIndex uint8 = (1 - s.turn) * 6
	startSquare := move.OriginSquare()
	startBoard := Bitboard(1 << Bitboard(startSquare))
	var startBoardPtr *Bitboard = nil
	startBoardIndex := 0
	for i := friendIndex; i < friendIndex+6; i++ {
		board := &s.board[i]
		if *board&startBoard != 0 {
			startBoardPtr = board
			startBoardIndex = int(i)
		}
	}
	desSquare := move.DestinationSquare()
	desBoard := Bitboard(1 << Bitboard(desSquare))
	var desBoardPtr *Bitboard = nil
	desBoardIndex := 0
	for i := enemyIndex; i < enemyIndex+6; i++ {
		board := &s.board[i]
		if *board&desBoard != 0 {
			desBoardPtr = board
			desBoardIndex = int(i)
		}
	}
	*startBoardPtr ^= startBoard
	*startBoardPtr |= desBoard
	if desBoardPtr != nil {
		*desBoardPtr ^= desBoard
		s.captureHistory.Push(uint8(desBoardIndex), s.ply)
	}
	specialMove := move.SpecialMove()
	if specialMove == CastleSpecialMove {
		*castleCounter++
		rankIndex := Square(s.turn * 56)
		rookSquare := Square(5 + rankIndex)
		startingRookSquare := Square(7 + rankIndex)
		if desSquare == Square(2) || desSquare == Square(58) {
			s.castleAvailability[s.turn+2] = false
			s.castleHistory.Push(s.turn+2, s.ply)
			rookSquare = Square(3 + rankIndex)
			startingRookSquare = Square(0 + rankIndex)
		} else {
			s.castleAvailability[s.turn] = false
			s.castleHistory.Push(s.turn, s.ply)
		}
		s.board[friendIndex+Rook] ^= Bitboard(1 << Bitboard(startingRookSquare))
		s.board[friendIndex+Rook] |= Bitboard(1 << Bitboard(rookSquare))
	} else if specialMove == PromotionSpecialMove {
		promotionType := move.PromotionType()
		promotionBoard := &s.board[friendIndex+Queen]
		if promotionType == RookPromotion {
			promotionBoard = &s.board[friendIndex+Rook]
		} else if promotionType == KnightPromotion {
			promotionBoard = &s.board[friendIndex+Knight]
		} else if promotionType == BishopPromotion {
			promotionBoard = &s.board[friendIndex+Bishop]
		}
		*promotionBoard |= desBoard
		*startBoardPtr ^= desBoard
	} else if specialMove == EnPassantSpacialMove {
		*enPassantCounter++
		enemyPawnBoard := &s.board[enemyIndex+Pawn]
		relativeDownStep := DownStep
		if s.turn == Black {
			relativeDownStep = UpStep
		}
		enPassantCaptureSquare := desSquare.Step(relativeDownStep)
		enPassantCaptureBoard := Bitboard(1 << Bitboard(enPassantCaptureSquare))
		*enemyPawnBoard ^= enPassantCaptureBoard
		s.captureHistory.Push(enemyIndex+Pawn, s.ply)
	}
	enemyKingBoard := s.board[enemyIndex+King]
	enemyBoard := enemyKingBoard | s.board[enemyIndex+Queen] | s.board[enemyIndex+Rook] | s.board[enemyIndex+Bishop] | s.board[enemyIndex+Knight] | s.board[enemyIndex+Pawn]
	bishopBoard := s.board[friendIndex+Bishop] | s.board[friendIndex+Queen]
	rookBoard := s.board[friendIndex+Rook] | s.board[friendIndex+Queen]
	knightBoard := s.board[friendIndex+Knight]
	kingBoard := s.board[friendIndex+King]
	pawnBoard := s.board[friendIndex+Pawn]
	combinedBoard := bishopBoard | rookBoard | knightBoard | kingBoard | pawnBoard
	friendSafetyCheck := &SafetyCheckBoards{bishopBoard, rookBoard, knightBoard, kingBoard, pawnBoard, combinedBoard}
	s.canEnpassant = false
	s.enPassantSquare = Square(100)
	if startBoardIndex == int(friendIndex)+Pawn {
		rankDiff := startSquare.Rank() - desSquare.Rank()
		if rankDiff == 2 || rankDiff == -2 {
			if s.turn == 0 {
				s.enPassantSquare = startSquare.Step(UpStep)
			} else {
				s.enPassantSquare = startSquare.Step(DownStep)
			}
			s.enPassantSquareHistory.Push(s.enPassantSquare, s.ply)
			s.canEnpassant = true
		}
	} else if startBoardIndex == int(friendIndex)+King {
		if s.castleAvailability[s.turn] {
			s.castleAvailability[s.turn] = false
			s.castleHistory.Push(s.turn, s.ply)
		}
		if s.castleAvailability[s.turn+2] {
			s.castleAvailability[s.turn+2] = false
			s.castleHistory.Push(s.turn+2, s.ply)
		}
	} else if startBoardIndex == int(friendIndex)+Rook {
		if startSquare == Square(7+(8*s.turn)) && s.castleAvailability[s.turn] {
			s.castleAvailability[s.turn] = false
			s.castleHistory.Push(s.turn, s.ply)
		}
		if startSquare == Square(8*s.turn) && s.castleAvailability[s.turn+2] {
			s.castleAvailability[s.turn+2] = false
			s.castleHistory.Push(s.turn+2, s.ply)
		}
	}
	s.turn = 1 - s.turn
	s.check = !isSquareSafe(PopLSB(&enemyKingBoard), enemyBoard, friendSafetyCheck, s.turn)
	if s.check {
		*checkCounter++
		// fmt.Println(s)
	}
	s.ply++
}

func (s *State) UnMakeMove(move Move) {
	var friendIndex uint8 = s.turn * 6
	var enemyIndex uint8 = (1 - s.turn) * 6
	startSquare := move.OriginSquare()
	startBoard := Bitboard(1 << Bitboard(startSquare))
	desSquare := move.DestinationSquare()
	desBoard := Bitboard(1 << Bitboard(desSquare))
	var desBoardPtr *Bitboard = nil
	for i := enemyIndex; i < enemyIndex+6; i++ {
		board := &s.board[i]
		if *board&desBoard != 0 {
			desBoardPtr = board
		}
	}
	*desBoardPtr |= startBoard
	*desBoardPtr ^= desBoard
	if s.ply-1 == s.captureHistory.MostRecentCapturePly() {
		capture := s.captureHistory.Pop()
		capturedPiece := capture.piece
		captureBoardPtr := &s.board[capturedPiece]
		*captureBoardPtr |= desBoard
	}
	specialMove := move.SpecialMove()
	if specialMove == CastleSpecialMove {
		rankIndex := Square((1 - s.turn) * 56)
		startingRookSquare := Square(5 + rankIndex)
		endingRookSquare := Square(7 + rankIndex)
		if desSquare == Square(2) || desSquare == Square(58) {
			startingRookSquare = Square(3 + rankIndex)
			endingRookSquare = Square(0 + rankIndex)
		}
		s.board[enemyIndex+Rook] ^= Bitboard(1 << Bitboard(startingRookSquare))
		s.board[enemyIndex+Rook] |= Bitboard(1 << Bitboard(endingRookSquare))
	} else if specialMove == EnPassantSpacialMove {
		relativeUpStep := UpStep
		if s.turn == Black {
			relativeUpStep = DownStep
		}
		enPassantSquare := desSquare.Step(relativeUpStep)
		enPassantBoard := Bitboard(1 << Bitboard(enPassantSquare))
		s.board[friendIndex+Pawn] ^= desBoard
		s.board[friendIndex+Pawn] |= enPassantBoard
	} else if specialMove == PromotionSpecialMove {
		*desBoardPtr ^= startBoard
		enemyPawnBoard := &s.board[enemyIndex+Pawn]
		*enemyPawnBoard |= startBoard
	}
	enemyKingBoard := s.board[friendIndex+King]
	enemyBoard := enemyKingBoard | s.board[friendIndex+Queen] | s.board[friendIndex+Rook] | s.board[friendIndex+Bishop] | s.board[friendIndex+Knight] | s.board[friendIndex+Pawn]
	bishopBoard := s.board[enemyIndex+Bishop] | s.board[enemyIndex+Queen]
	rookBoard := s.board[enemyIndex+Rook] | s.board[enemyIndex+Queen]
	knightBoard := s.board[enemyIndex+Knight]
	kingBoard := s.board[enemyIndex+King]
	pawnBoard := s.board[enemyIndex+Pawn]
	combinedBoard := bishopBoard | rookBoard | knightBoard | kingBoard | pawnBoard
	friendSafetyCheck := &SafetyCheckBoards{bishopBoard, rookBoard, knightBoard, kingBoard, pawnBoard, combinedBoard}
	if s.enPassantSquareHistory.MostRecentCapturePly() == s.ply-1 {
		enPassantEntry := s.enPassantSquareHistory.Pop()
		s.enPassantSquare = enPassantEntry.square
		s.canEnpassant = true
	} else {
		s.enPassantSquare = Square(100)
		s.canEnpassant = false
	}
	if s.castleHistory.MostRecentCapturePly() == s.ply-1 {
		castleEntry := s.castleHistory.Pop()
		s.castleAvailability[castleEntry.castle] = true
		if s.castleHistory.MostRecentCapturePly() == s.ply-1 {
			castleEntry = s.castleHistory.Pop()
			s.castleAvailability[castleEntry.castle] = true
		}
	}
	s.turn = 1 - s.turn
	s.check = !isSquareSafe(PopLSB(&enemyKingBoard), enemyBoard, friendSafetyCheck, s.turn)
	s.ply--
}

func (s *State) genAllMoves(includeQuiets bool) *[]Move {
	moves := make([]Move, 0, 50)
	// We want the pop function to pop the the bits at the top of the board relative to whos turn
	// it is. So when it's white's turn we pop the most significant bit first and with black
	// we pop the least significant bit first
	var popFunction func(*Bitboard) Square
	if s.turn == White {
		popFunction = PopMSB
	} else {
		popFunction = PopLSB
	}
	var friendIndex uint8 = s.turn * 6
	var enemyIndex uint8 = (1 - s.turn) * 6
	friendBoard := s.board[friendIndex]
	for i := friendIndex + 1; i < friendIndex+6; i++ {
		friendBoard |= s.board[i]
	}
	enemyBoard := s.board[enemyIndex]
	for i := enemyIndex + 1; i < enemyIndex+6; i++ {
		enemyBoard |= s.board[i]
	}
	occupied := friendBoard | enemyBoard
	notOccupied := ^occupied
	genInfo := &MoveGenInfo{includeQuiets, enemyBoard, occupied, notOccupied, popFunction}
	kingBoard := s.board[friendIndex+King]
	kingSquare := popFunction(&kingBoard)
	enemyBishopSliders := s.board[enemyIndex+Bishop] | s.board[enemyIndex+Queen]
	enemyRookSliders := s.board[enemyIndex+Rook] | s.board[enemyIndex+Queen]
	pinnedBoard, pinSafetys := getKingPins(kingSquare, friendBoard, enemyBishopSliders, enemyRookSliders, enemyBoard)
	// fmt.Println(pinnedBoard)
	// for _, pinSafe := range *pinSafetys {
	// 	fmt.Println(pinSafe.key)
	// 	fmt.Println(pinSafe.safeSquares)
	// }
	safetyCheckBoard := &SafetyCheckBoards{enemyBishopSliders, enemyRookSliders, s.board[enemyIndex+Knight], s.board[enemyIndex+King], s.board[enemyIndex+Pawn], enemyBoard}
	checkBlockerSquares := UniversalBitboard
	if s.check {
		checkBlockerSquares = GetCheckBlockerSquares(kingSquare, friendBoard, safetyCheckBoard, s.turn)
	}
	// Start Bishop
	if checkBlockerSquares != EmptyBitboard {
		bishopBoard := s.board[friendIndex+Bishop]
		genSliderMoves(bishopBoard, &moves, genInfo, getBishopMoves, pinnedBoard, pinSafetys, checkBlockerSquares)
	}
	// End Bishop
	// Start Knight
	if checkBlockerSquares != 0 {
		knightBoard := s.board[friendIndex+Knight]
		for knightBoard != 0 {
			knightSquare := popFunction(&knightBoard)
			if pinnedBoard&(1<<Bitboard(knightSquare)) == 0 {
				knightMoves := moveBoards[Knight][knightSquare]
				knightAttacks := knightMoves & enemyBoard & checkBlockerSquares
				for knightAttacks != 0 {
					attackSquare := popFunction(&knightAttacks)
					moves = append(moves, BuildMove(knightSquare, attackSquare, 0, 0))
				}
				if includeQuiets {
					knightQuiets := knightMoves & notOccupied & checkBlockerSquares
					for knightQuiets != 0 {
						quietSquare := popFunction(&knightQuiets)
						moves = append(moves, BuildMove(knightSquare, quietSquare, 0, 0))
					}
				}
			}
		}
	}
	// End Knight
	// Start Queen
	if checkBlockerSquares != 0 {
		queenBoard := s.board[friendIndex+Queen]
		genSliderMoves(queenBoard, &moves, genInfo, getQueenMoves, pinnedBoard, pinSafetys, checkBlockerSquares)
	}
	// End Queen
	// Start Rook
	if checkBlockerSquares != 0 {
		rookBoard := s.board[friendIndex+Rook]
		genSliderMoves(rookBoard, &moves, genInfo, getRookMoves, pinnedBoard, pinSafetys, checkBlockerSquares)
	}
	// End Rook
	// Start Pawn
	if checkBlockerSquares != 0 {
		enemyEnPassantBoard := enemyBoard | Bitboard(1<<s.enPassantSquare)
		pawnBoard := s.board[friendIndex+Pawn]
		var moveStep Step = UpStep
		var homeRank int8 = 1
		var promotionRank int8 = 7
		if s.turn == Black {
			moveStep = DownStep
			homeRank = 6
			promotionRank = 0
		}
		for pawnBoard != 0 {
			pawnSquare := popFunction(&pawnBoard)
			var safeBoard Bitboard = UniversalBitboard
			if pinnedBoard&(1<<Bitboard(pawnSquare)) != 0 {
				pinFound := false
				doublePinned := false
				var squarePinSafety PinSafety
				for _, pinSafety := range *pinSafetys {
					if pinSafety.key == pawnSquare {
						if pinFound {
							doublePinned = true
							break
						} else {
							pinFound = true
							squarePinSafety = pinSafety
						}
					}
					if doublePinned {
						safeBoard = EmptyBitboard
					} else {
						safeBoard = squarePinSafety.safeSquares
					}
				}
			}
			pawnAttacks := pawnAttackBoards[s.turn][pawnSquare] & enemyEnPassantBoard & safeBoard & checkBlockerSquares
			for pawnAttacks != 0 {
				attackSquare := popFunction(&pawnAttacks)
				if attackSquare == s.enPassantSquare {
					if s.canEnpassant {
						if s.EnPassantSafetyCheck(pawnSquare, attackSquare, friendIndex, enemyIndex, occupied) {
							moves = append(moves, BuildMove(pawnSquare, attackSquare, 0, EnPassantSpacialMove))
						}
					}
				} else {
					if attackSquare.Rank() == promotionRank {
						moves = append(moves, BuildMove(pawnSquare, attackSquare, 0, PromotionSpecialMove))
						moves = append(moves, BuildMove(pawnSquare, attackSquare, 1, PromotionSpecialMove))
						moves = append(moves, BuildMove(pawnSquare, attackSquare, 2, PromotionSpecialMove))
						moves = append(moves, BuildMove(pawnSquare, attackSquare, 3, PromotionSpecialMove))
					} else {
						moves = append(moves, BuildMove(pawnSquare, attackSquare, 0, 0))
					}
				}
			}
			if includeQuiets {
				pawnMoves := GetPawnMoves(pawnSquare, occupied, moveStep, homeRank) & safeBoard & checkBlockerSquares
				for pawnMoves != 0 {
					desSquare := popFunction(&pawnMoves)
					if desSquare.Rank() == promotionRank {
						moves = append(moves, BuildMove(pawnSquare, desSquare, 0, 2))
						moves = append(moves, BuildMove(pawnSquare, desSquare, 1, 2))
						moves = append(moves, BuildMove(pawnSquare, desSquare, 2, 2))
						moves = append(moves, BuildMove(pawnSquare, desSquare, 3, 2))
					} else {
						moves = append(moves, BuildMove(pawnSquare, desSquare, 0, 0))
					}
				}
			}

		}
	}
	// End Pawn
	// Start King
	kingBoard = s.board[friendIndex+King]
	noKingFriendBoard := friendBoard ^ kingBoard
	kingSquare = popFunction(&kingBoard)
	kingAttacks := moveBoards[King][kingSquare] & enemyBoard
	for kingAttacks != 0 {
		desSquare := popFunction(&kingAttacks)
		if isSquareSafe(desSquare, noKingFriendBoard, safetyCheckBoard, s.turn) {
			moves = append(moves, BuildMove(kingSquare, desSquare, 0, 0))
		}
	}
	if includeQuiets {
		kingQuiets := moveBoards[King][kingSquare] & notOccupied
		for kingQuiets != 0 {
			desSquare := popFunction(&kingQuiets)
			if isSquareSafe(desSquare, noKingFriendBoard, safetyCheckBoard, s.turn) {
				moves = append(moves, BuildMove(kingSquare, desSquare, 0, 0))
			}
		}
	}
	// Castle Start
	if !s.check {
		rankIndex := s.turn * 56
		// King Castle Start
		if s.castleAvailability[s.turn] {
			if occupied&Bitboard(0x60<<rankIndex) == 0 && s.board[friendIndex+Rook]&Bitboard(0x80<<rankIndex) != 0 {
				desSquare := kingSquare + 2
				if isSquareSafe(Square(5+rankIndex), noKingFriendBoard, safetyCheckBoard, s.turn) && isSquareSafe(desSquare, noKingFriendBoard, safetyCheckBoard, s.turn) {
					moves = append(moves, BuildMove(kingSquare, desSquare, 0, CastleSpecialMove))
				}
			}
		}
		// King Castle End
		// Queen Castle Start
		if s.castleAvailability[s.turn+2] {
			if occupied&Bitboard(0xE<<rankIndex) == 0 && s.board[friendIndex+Rook]&Bitboard(0x1<<rankIndex) != 0 {
				desSquare := kingSquare - 2
				if isSquareSafe(Square(3+rankIndex), noKingFriendBoard, safetyCheckBoard, s.turn) && isSquareSafe(desSquare, noKingFriendBoard, safetyCheckBoard, s.turn) {
					moves = append(moves, BuildMove(kingSquare, desSquare, 0, CastleSpecialMove))
				}
			}
		}
		// Queen Castle End
	}
	// Castle End
	// End King
	return &moves
}

func getKingPins(kingSquare Square, friendBoard Bitboard, bishopBoard Bitboard, rookBoard Bitboard, enemyBoard Bitboard) (Bitboard, *[]PinSafety) {
	pinSafetys := make([]PinSafety, 0, 8)
	var pinnedBoard Bitboard = 0
	for _, step := range queenSteps {
		if kingSquare.tryStep(step) {
			possiblePinners := bishopBoard
			if step == UpStep || step == RightStep || step == DownStep || step == LeftStep {
				possiblePinners = rookBoard
			}
			nonPinners := enemyBoard & (^possiblePinners)
			stepSquare := kingSquare
			pinnedFound := false
			doubleBlocked := false
			enemyFound := false
			var pinnedSquare Square
			var safeSquares Bitboard = 0
			for stepSquare.tryStep(step) {
				stepSquare = stepSquare.Step(step)
				stepBoard := Bitboard(1 << stepSquare)
				safeSquares |= stepBoard
				if friendBoard&stepBoard != 0 {
					if pinnedFound {
						doubleBlocked = true
						break
					} else {
						pinnedFound = true
						pinnedSquare = stepSquare
					}
				}
				if possiblePinners&stepBoard != 0 {
					enemyFound = true
					break
				}
				if nonPinners&stepBoard != 0 {
					break
				}
			}
			if pinnedFound && enemyFound && (!doubleBlocked) {
				pin := Bitboard(1 << pinnedSquare)
				pinnedBoard |= pin
				pinSafetys = append(pinSafetys, PinSafety{pinnedSquare, safeSquares})
			}
		}
	}
	return pinnedBoard, &pinSafetys
}

func genSliderMoves(board Bitboard, moves *[]Move, genInfo *MoveGenInfo, magicRetriever func(Square, Bitboard) Bitboard, pinnedBoard Bitboard, pinSafetys *[]PinSafety, checkBlockerSquares Bitboard) {
	for board != 0 {
		safeSquares := UniversalBitboard
		sliderSquare := genInfo.popFunction(&board)
		if pinnedBoard&(1<<Bitboard(sliderSquare)) != 0 {
			pinFound := false
			doublePinned := false
			var squarePinSafety PinSafety
			for _, pinSafety := range *pinSafetys {
				if pinSafety.key == sliderSquare {
					if pinFound {
						doublePinned = true
						break
					} else {
						pinFound = true
						squarePinSafety = pinSafety
					}
				}
			}
			if pinFound && !doublePinned {
				safeSquares = squarePinSafety.safeSquares
			}
		}
		sliderMoves := magicRetriever(sliderSquare, genInfo.occupied)
		sliderAttacks := sliderMoves & genInfo.enemyBoard & safeSquares & checkBlockerSquares
		for sliderAttacks != 0 {
			attackSquare := genInfo.popFunction(&sliderAttacks)
			*moves = append(*moves, BuildMove(sliderSquare, attackSquare, 0, 0))
		}
		if genInfo.includeQuiets {
			sliderQuiets := sliderMoves & genInfo.notOccupied & safeSquares & checkBlockerSquares
			for sliderQuiets != 0 {
				quietSquare := genInfo.popFunction(&sliderQuiets)
				*moves = append(*moves, BuildMove(sliderSquare, quietSquare, 0, 0))
			}
		}
	}
}

func isSquareSafe(square Square, friendBoard Bitboard, enemyBoards *SafetyCheckBoards, turn uint8) bool {
	occupied := friendBoard | enemyBoards.combinedBoard
	bishopCast := getBishopMoves(square, occupied)
	if bishopCast&enemyBoards.bishopsBoards != 0 {
		return false
	}
	rookCast := getRookMoves(square, occupied)
	if rookCast&enemyBoards.rooksBoard != 0 {
		return false
	}
	knightCast := moveBoards[Knight][square]
	if knightCast&enemyBoards.knightsBoard != 0 {
		return false
	}
	kingCast := moveBoards[King][square]
	if kingCast&enemyBoards.kingBoard != 0 {
		return false
	}
	pawnCast := pawnAttackBoards[turn][square]
	return pawnCast&enemyBoards.pawnsBoard == 0
}

func (s *State) EnPassantSafetyCheck(startingSquare Square, desSquare Square, friendIndex uint8, enemyIndex uint8, occupied Bitboard) bool {
	startingBoard := Bitboard(1 << Bitboard(startingSquare))
	desBoard := Bitboard(1 << Bitboard(desSquare))
	friendPawnBoard := s.board[friendIndex+Pawn]
	friendPawnBoard ^= startingBoard
	friendPawnBoard ^= desBoard
	enemyPawnReverseStep := DownStep
	if s.turn == Black {
		enemyPawnReverseStep = UpStep
	}
	targetSquare := desSquare.Step(enemyPawnReverseStep)
	targetBoard := Bitboard(1 << Bitboard(targetSquare))
	enemyPawnBoard := s.board[enemyIndex+Pawn]
	enemyPawnBoard ^= targetBoard
	occupied ^= startingBoard | targetBoard
	occupied |= desBoard
	enemyBishopBoard := s.board[enemyIndex+Bishop] | s.board[enemyIndex+Queen]
	enemyRookBoard := s.board[enemyIndex+Rook] | s.board[enemyIndex+Queen]
	friendKingBoard := s.board[friendIndex+King]
	kingSquare := PopLSB(&friendKingBoard)
	bishopCast := getBishopMoves(kingSquare, occupied)
	if bishopCast&enemyBishopBoard != 0 {
		return false
	}
	rookCast := getRookMoves(kingSquare, occupied)
	return rookCast&enemyRookBoard == 0
}

func GetCheckBlockerSquares(square Square, friendBoard Bitboard, enemyBoards *SafetyCheckBoards, turn uint8) Bitboard {
	occupied := friendBoard | enemyBoards.combinedBoard
	safeSquares := EmptyBitboard
	checkFound := false
	bishopCast := getBishopMoves(square, occupied)
	if bishopCast&enemyBoards.bishopsBoards != 0 {
		for _, step := range bishopSteps {
			if square.tryStep(step) {
				stepSafeSquares := EmptyBitboard
				enemyFound := false
				stepSquare := square
				for stepSquare.tryStep(step) {
					stepSquare = stepSquare.Step(step)
					stepBoard := Bitboard(1 << Bitboard(stepSquare))
					if enemyBoards.bishopsBoards&stepBoard != 0 {
						stepSafeSquares |= stepBoard
						enemyFound = true
						break
					} else if occupied&stepBoard != 0 {
						break
					} else {
						stepSafeSquares |= stepBoard
					}
				}
				if enemyFound {
					if !checkFound {
						safeSquares |= stepSafeSquares
						checkFound = true
					} else {
						return EmptyBitboard
					}
				}
			}
		}
	}
	rookCast := getRookMoves(square, occupied)
	if rookCast&enemyBoards.rooksBoard != 0 {
		for _, step := range rookSteps {
			if square.tryStep(step) {
				stepSafeSquares := EmptyBitboard
				enemyFound := false
				stepSquare := square
				for stepSquare.tryStep(step) {
					stepSquare = stepSquare.Step(step)
					stepBoard := Bitboard(1 << Bitboard(stepSquare))
					if enemyBoards.rooksBoard&stepBoard != 0 {
						stepSafeSquares |= stepBoard
						enemyFound = true
						break
					} else if occupied&stepBoard != 0 {
						break
					} else {
						stepSafeSquares |= stepBoard
					}
				}
				if enemyFound {
					if !checkFound {
						safeSquares |= stepSafeSquares
						checkFound = true
					} else {
						return EmptyBitboard
					}
				}
			}
		}
	}
	knightCast := moveBoards[Knight][square]
	if knightCast&enemyBoards.knightsBoard != 0 {
		for knightCast != 0 {
			knightSquare := PopLSB(&knightCast)
			knightBoard := Bitboard(1 << Bitboard(knightSquare))
			if enemyBoards.knightsBoard&knightBoard != 0 {
				if !checkFound {
					safeSquares |= knightBoard
					checkFound = true
				} else {
					return EmptyBitboard
				}
			}
		}
	}
	kingCast := moveBoards[King][square]
	if kingCast&enemyBoards.kingBoard != 0 {
		for kingCast != 0 {
			kingSquare := PopLSB(&kingCast)
			kingBoard := Bitboard(1 << Bitboard(kingSquare))
			if enemyBoards.kingBoard&kingBoard != 0 {
				if !checkFound {
					safeSquares |= kingBoard
					checkFound = true
				} else {
					return EmptyBitboard
				}
			}
		}
	}
	pawnCast := pawnAttackBoards[turn][square]
	if pawnCast&enemyBoards.pawnsBoard != 0 {
		for pawnCast != 0 {
			pawnSquare := PopLSB(&pawnCast)
			pawnBoard := Bitboard(1 << Bitboard(pawnSquare))
			if enemyBoards.pawnsBoard&pawnBoard != 0 {
				if !checkFound {
					safeSquares |= pawnBoard
					checkFound = true
				} else {
					return EmptyBitboard
				}
			}
		}
	}
	return safeSquares
}

func FenState(fenString string) *State {
	pieceMap := map[rune]int{'K': 0, 'Q': 1, 'R': 2, 'B': 3, 'N': 4, 'P': 5, 'k': 6, 'q': 7, 'r': 8, 'b': 9, 'n': 10, 'p': 11}
	splitFenString := strings.Split(fenString, " ")
	boardString := strings.Split(splitFenString[0], "/")
	board := Board{}
	for i := 0; i < 8; i++ {
		r := boardString[7-i]
		column := 0
		for _, c := range r {
			_, ok := pieceMap[c]
			if ok {
				board[pieceMap[c]] |= 1 << (i*8 + column)
			} else {
				num, err := strconv.Atoi(string(c))
				if err != nil {
					panic("Invalid Fen String (Invalid Piece Structure)")
				}
				column += num - 1
			}
			column++
		}
	}
	turnString := splitFenString[1]
	turn := uint8(0)
	if turnString == "b" {
		turn = 1
	} else if turnString != "w" {
		panic("Invalid Fen String (Invalid Turn)")
	}
	castleAvailability := CastleAvailability{}
	castleString := splitFenString[2]
	if castleString != "-" {
		castleOptions := [4]rune{'K', 'k', 'Q', 'q'}
		for i, r := range castleOptions {
			for _, c := range castleString {
				if r == c {
					castleAvailability[i] = true
				}
			}
		}
	}
	var enPassantSquare Square
	enpassantString := splitFenString[3]
	canEnpassant := false
	if enpassantString != "-" {
		canEnpassant = true
		rankMap := map[rune]int{'a': 0, 'b': 1, 'c': 2, 'd': 3, 'e': 4, 'f': 5, 'g': 6, 'h': 7}
		fileRune := enpassantString[0]
		file, ok := rankMap[rune(fileRune)]
		if !ok {
			panic("Invalid Fen String (Invalid En Passant Square)")
		}
		rankRune := enpassantString[1]
		rank, err := strconv.Atoi(string(rankRune))
		if err != nil {
			panic("Invalid Fen String (Invalid En Passant Square)")
		}
		enPassantSquare = Square(rank*8 + file)

	}
	return &State{board: &board, turn: turn, enPassantSquare: enPassantSquare, check: false, captureHistory: NewCaptureHistory(32), canEnpassant: canEnpassant, enPassantSquareHistory: NewEnpassantHistory(16), ply: 0, castleAvailability: &castleAvailability, castleHistory: NewCastleHistory(4)}
}

func StartingFen() *State {
	return FenState("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
}

func (s *State) String() string {
	result := ""
	result += s.board.String() + "\n"
	if s.turn == 0 {
		result += "Turn: White\n"
	} else if s.turn == 1 {
		result += "Turn: Black\n"
	} else {
		result += "Invalid Turn"
	}
	if s.check {
		result += "In Check\n"
	}
	if s.canEnpassant {
		result += "Can En Passant: " + s.enPassantSquare.String() + "\n"
	}
	result += fmt.Sprintf("Castle Availability: %v", *s.castleAvailability)
	return result + "\n"
}
