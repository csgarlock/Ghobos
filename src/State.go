package main

import (
	"fmt"
	"strconv"
	"strings"
)

type KillerTable [][2]Move

type SearchParameters struct {
	killerTable *KillerTable
	trueDepth   int16 // The true depth from the root node
}

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
	lastCapOrPawn          uint16
	ply                    uint16
	castleAvailability     *CastleAvailability
	castleHistory          *CastleHistory
	fiftyMoveHistory       *FiftyMoveHistory
	repetitionMap          *RepetitionMap
	hashcode               uint64
	hashHistory            *HashHistory
	searchParameters       SearchParameters
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

type QuietMove struct {
	move         Move
	historyValue uint64
}

type CaptureMove struct {
	move         Move
	captureValue int32
}

func (s *State) MakeMove(move Move) {
	var friendIndex uint8 = s.turn * 6
	var enemyIndex uint8 = (1 - s.turn) * 6
	startSquare := move.OriginSquare()
	startBoard := Bitboard(1 << Bitboard(startSquare))
	var startBoardPtr *Bitboard = nil
	s.lastCapOrPawn += 1
	startBoardIndex := 0
	for i := friendIndex; i < friendIndex+6; i++ {
		board := &s.board[i]
		if *board&startBoard != 0 {
			startBoardPtr = board
			startBoardIndex = int(i)
		}
	}
	s.hashcode ^= squareHashes[startBoardIndex][startSquare]
	desSquare := move.DestinationSquare()
	desBoard := Bitboard(1 << Bitboard(desSquare))
	s.hashcode ^= squareHashes[startBoardIndex][desSquare]
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
	isCapture := false
	if desBoardPtr != nil {
		*desBoardPtr ^= desBoard
		s.captureHistory.Push(uint8(desBoardIndex), s.ply)
		s.fiftyMoveHistory.Push(s.lastCapOrPawn-1, s.ply)
		s.lastCapOrPawn = 0
		isCapture = true
		s.hashcode ^= squareHashes[desBoardIndex][desSquare]
	}
	if s.canEnpassant {
		s.hashcode ^= enPassantHashes[s.enPassantSquare%8]
	}
	specialMove := move.SpecialMove()
	if specialMove == CastleSpecialMove {
		rankIndex := Square(s.turn * 56)
		rookSquare := Square(5 + rankIndex)
		startingRookSquare := Square(7 + rankIndex)
		if desSquare == Square(2) || desSquare == Square(58) {
			s.castleAvailability[s.turn+2] = false
			s.hashcode ^= castleHashes[s.turn+2]
			s.castleHistory.Push(s.turn+2, s.ply)
			rookSquare = Square(3 + rankIndex)
			startingRookSquare = Square(0 + rankIndex)
		} else {
			s.castleAvailability[s.turn] = false
			s.hashcode ^= castleHashes[s.turn]
			s.castleHistory.Push(s.turn, s.ply)
		}
		s.board[friendIndex+Rook] ^= Bitboard(1 << Bitboard(startingRookSquare))
		s.board[friendIndex+Rook] |= Bitboard(1 << Bitboard(rookSquare))
		s.hashcode ^= squareHashes[friendIndex+Rook][startingRookSquare] ^ squareHashes[friendIndex+Rook][rookSquare]
	} else if specialMove == PromotionSpecialMove {
		s.hashcode ^= squareHashes[startBoardIndex][desSquare]
		promotionType := move.PromotionType()
		promotionBoard := &s.board[friendIndex+Queen]
		if promotionType == RookPromotion {
			promotionBoard = &s.board[friendIndex+Rook]
			s.hashcode ^= squareHashes[friendIndex+Rook][desSquare]
		} else if promotionType == KnightPromotion {
			promotionBoard = &s.board[friendIndex+Knight]
			s.hashcode ^= squareHashes[friendIndex+Knight][desSquare]
		} else if promotionType == BishopPromotion {
			promotionBoard = &s.board[friendIndex+Bishop]
			s.hashcode ^= squareHashes[friendIndex+Bishop][desSquare]
		} else {
			s.hashcode ^= squareHashes[friendIndex+Queen][desSquare]
		}
		*promotionBoard |= desBoard
		*startBoardPtr ^= desBoard
	} else if specialMove == EnPassantSpacialMove {
		enemyPawnBoard := &s.board[enemyIndex+Pawn]
		relativeDownStep := DownStep
		if s.turn == Black {
			relativeDownStep = UpStep
		}
		enPassantCaptureSquare := desSquare.Step(relativeDownStep)
		enPassantCaptureBoard := Bitboard(1 << Bitboard(enPassantCaptureSquare))
		*enemyPawnBoard ^= enPassantCaptureBoard
		s.hashcode ^= squareHashes[enemyIndex+Pawn][enPassantCaptureSquare]
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
			s.hashcode ^= enPassantHashes[s.enPassantSquare%8]
			s.canEnpassant = true
		}
		if !isCapture {
			s.fiftyMoveHistory.Push(s.lastCapOrPawn-1, s.ply)
			s.lastCapOrPawn = 0
		}
	} else if startBoardIndex == int(friendIndex)+King {
		if s.castleAvailability[s.turn] {
			s.castleAvailability[s.turn] = false
			s.hashcode ^= castleHashes[s.turn]
			s.castleHistory.Push(s.turn, s.ply)
		}
		if s.castleAvailability[s.turn+2] {
			s.castleAvailability[s.turn+2] = false
			s.hashcode ^= castleHashes[s.turn+2]
			s.castleHistory.Push(s.turn+2, s.ply)
		}
	} else if startBoardIndex == int(friendIndex)+Rook {
		if startSquare == Square(7+(8*s.turn)) && s.castleAvailability[s.turn] {
			s.castleAvailability[s.turn] = false
			s.hashcode ^= castleHashes[s.turn]
			s.castleHistory.Push(s.turn, s.ply)
		}
		if startSquare == Square(8*s.turn) && s.castleAvailability[s.turn+2] {
			s.castleAvailability[s.turn+2] = false
			s.hashcode ^= castleHashes[s.turn+2]
			s.castleHistory.Push(s.turn+2, s.ply)
		}
	}
	s.turn = 1 - s.turn
	s.hashcode ^= blackHash
	s.hashHistory.Push(s.hashcode)
	s.check = !isSquareSafe(PopLSB(&enemyKingBoard), enemyBoard, friendSafetyCheck, s.turn)
	s.ply++
	s.repetitionMap.add(s.hashcode)
}

func (s *State) UnMakeMove(move Move) {
	s.repetitionMap.remove(s.hashcode)
	var friendIndex uint8 = s.turn * 6
	var enemyIndex uint8 = (1 - s.turn) * 6
	startSquare := move.OriginSquare()
	startBoard := Bitboard(1 << Bitboard(startSquare))
	desSquare := move.DestinationSquare()
	desBoard := Bitboard(1 << Bitboard(desSquare))
	s.lastCapOrPawn -= 1
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
	if s.enPassantSquareHistory.MostRecentCapturePly() == s.ply-1 {
		s.enPassantSquareHistory.Pop()
	}
	if s.enPassantSquareHistory.MostRecentCapturePly() == s.ply-2 {
		enPassantEntry := s.enPassantSquareHistory.Peek()
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
	if s.fiftyMoveHistory.lastReset() == s.ply-1 {
		s.lastCapOrPawn = s.fiftyMoveHistory.Pop().lastCount
	}
	s.turn = 1 - s.turn
	enemyKingBoard := s.board[enemyIndex+King]
	enemyBoard := enemyKingBoard | s.board[enemyIndex+Queen] | s.board[enemyIndex+Rook] | s.board[enemyIndex+Bishop] | s.board[enemyIndex+Knight] | s.board[enemyIndex+Pawn]
	bishopBoard := s.board[friendIndex+Bishop] | s.board[friendIndex+Queen]
	rookBoard := s.board[friendIndex+Rook] | s.board[friendIndex+Queen]
	knightBoard := s.board[friendIndex+Knight]
	kingBoard := s.board[friendIndex+King]
	pawnBoard := s.board[friendIndex+Pawn]
	combinedBoard := bishopBoard | rookBoard | knightBoard | kingBoard | pawnBoard
	friendSafetyCheck := &SafetyCheckBoards{bishopBoard, rookBoard, knightBoard, kingBoard, pawnBoard, combinedBoard}
	s.check = !isSquareSafe(PopLSB(&enemyKingBoard), enemyBoard, friendSafetyCheck, s.turn)
	s.ply--
	s.hashHistory.Pop()
	s.hashcode = s.hashHistory.Peek()
}

func (s *State) quickGenMoves() *[]Move {
	captures, quiets := s.genAllMoves(true)
	moves := make([]Move, len(*captures)+len(*quiets))
	totalIndex := 0
	badCutoff := len(*captures)
	for i := 0; i < len(*captures); i++ {
		if (*captures)[i].captureValue >= 0 {
			moves[totalIndex] = (*captures)[i].move
			totalIndex++
		} else {
			badCutoff = i
			break
		}
	}
	for i := 0; i < len(*quiets); i++ {
		moves[totalIndex] = (*quiets)[i].move
		totalIndex++
	}
	for i := badCutoff; i < len(*captures); i++ {
		moves[totalIndex] = (*captures)[i].move
		totalIndex++
	}
	return &moves
}

func (s *State) genAllMoves(includeQuiets bool) (*[]CaptureMove, *[]QuietMove) {
	// We want the pop function to pop the the bits at the top of the board relative to whos turn
	// it is. So when it's white's turn we pop the most significant bit first and with black
	// we pop the least significant bit first
	var popFunction func(*Bitboard) Square
	if s.turn == White {
		popFunction = PopMSB
	} else {
		popFunction = PopLSB
	}
	captures := make([]CaptureMove, 0, 20)
	quiets := make([]QuietMove, 0, 20)
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
	safetyCheckBoard := &SafetyCheckBoards{enemyBishopSliders, enemyRookSliders, s.board[enemyIndex+Knight], s.board[enemyIndex+King], s.board[enemyIndex+Pawn], enemyBoard}
	checkBlockerSquares := UniversalBitboard
	enPassantCheckBlockerSquares := UniversalBitboard
	if s.check {
		checkBlockerSquares, enPassantCheckBlockerSquares = s.GetCheckBlockerSquares(kingSquare, friendBoard, safetyCheckBoard, s.turn)
	}
	// Start Bishop
	if checkBlockerSquares != EmptyBitboard {
		bishopBoard := s.board[friendIndex+Bishop]
		genSliderMoves(s, Bishop, bishopBoard, &captures, &quiets, genInfo, getBishopMoves, pinnedBoard, pinSafetys, checkBlockerSquares)
	}
	// End Bishop
	// Start Knight
	if checkBlockerSquares != 0 {
		pieceIndex := Knight + friendIndex
		knightBoard := s.board[pieceIndex]
		for knightBoard != 0 {
			knightSquare := popFunction(&knightBoard)
			if pinnedBoard&(1<<Bitboard(knightSquare)) == 0 {
				knightMoves := moveBoards[Knight][knightSquare]
				knightAttacks := knightMoves & enemyBoard & checkBlockerSquares
				for knightAttacks != 0 {
					attackSquare := popFunction(&knightAttacks)
					attackedPiece := s.board.getColorPieceAt(attackSquare, 1-s.turn)
					captures = append(captures, CaptureMove{BuildMove(knightSquare, attackSquare, 0, 0), valueTable[attackedPiece%6] - valueTable[Knight]})
				}
				if includeQuiets {
					knightQuiets := knightMoves & notOccupied & checkBlockerSquares
					for knightQuiets != 0 {
						quietSquare := popFunction(&knightQuiets)
						quiets = append(quiets, QuietMove{BuildMove(knightSquare, quietSquare, 0, 0), historyTable[pieceIndex][quietSquare]})
					}
				}
			}
		}
	}
	// End Knight
	// Start Queen
	if checkBlockerSquares != 0 {
		queenBoard := s.board[friendIndex+Queen]
		genSliderMoves(s, Queen, queenBoard, &captures, &quiets, genInfo, getQueenMoves, pinnedBoard, pinSafetys, checkBlockerSquares)
	}
	// End Queen
	// Start Rook
	if checkBlockerSquares != 0 {
		rookBoard := s.board[friendIndex+Rook]
		genSliderMoves(s, Rook, rookBoard, &captures, &quiets, genInfo, getRookMoves, pinnedBoard, pinSafetys, checkBlockerSquares)
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
			pawnAttacks := pawnAttackBoards[s.turn][pawnSquare] & enemyEnPassantBoard & safeBoard & (checkBlockerSquares | enPassantCheckBlockerSquares)
			for pawnAttacks != 0 {
				attackSquare := popFunction(&pawnAttacks)
				attackedPiece := s.board.getColorPieceAt(attackSquare, 1-s.turn)
				if attackSquare == s.enPassantSquare {
					if s.canEnpassant {
						if s.EnPassantSafetyCheck(pawnSquare, attackSquare, friendIndex, enemyIndex, occupied) {
							captures = append(captures, CaptureMove{BuildMove(pawnSquare, attackSquare, 0, EnPassantSpacialMove), 0})
						}
					}
				} else {
					attackValue := valueTable[attackedPiece%6] - valueTable[Pawn]
					if attackSquare.Rank() == promotionRank {
						captures = append(captures, CaptureMove{BuildMove(pawnSquare, attackSquare, 0, PromotionSpecialMove), attackValue})
						captures = append(captures, CaptureMove{BuildMove(pawnSquare, attackSquare, 1, PromotionSpecialMove), attackValue})
						captures = append(captures, CaptureMove{BuildMove(pawnSquare, attackSquare, 2, PromotionSpecialMove), attackValue})
						captures = append(captures, CaptureMove{BuildMove(pawnSquare, attackSquare, 3, PromotionSpecialMove), attackValue})
					} else {
						captures = append(captures, CaptureMove{BuildMove(pawnSquare, attackSquare, 0, 0), attackValue})
					}
				}
			}
			if includeQuiets {
				pawnMoves := GetPawnMoves(pawnSquare, occupied, moveStep, homeRank) & safeBoard & checkBlockerSquares
				for pawnMoves != 0 {
					desSquare := popFunction(&pawnMoves)
					historyValue := historyTable[Pawn+friendIndex][desSquare]
					if desSquare.Rank() == promotionRank {
						quiets = append(quiets, QuietMove{BuildMove(pawnSquare, desSquare, 0, PromotionSpecialMove), historyValue})
						quiets = append(quiets, QuietMove{BuildMove(pawnSquare, desSquare, 1, PromotionSpecialMove), historyValue})
						quiets = append(quiets, QuietMove{BuildMove(pawnSquare, desSquare, 2, PromotionSpecialMove), historyValue})
						quiets = append(quiets, QuietMove{BuildMove(pawnSquare, desSquare, 3, PromotionSpecialMove), historyValue})
					} else {
						quiets = append(quiets, QuietMove{BuildMove(pawnSquare, desSquare, 0, 0), historyValue})
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
		attackedPiece := s.board.getColorPieceAt(desSquare, 1-s.turn)
		if isSquareSafe(desSquare, noKingFriendBoard, safetyCheckBoard, s.turn) {
			captures = append(captures, CaptureMove{BuildMove(kingSquare, desSquare, 0, 0), valueTable[attackedPiece%6] - valueTable[King]})
		}
	}
	if includeQuiets {
		kingQuiets := moveBoards[King][kingSquare] & notOccupied
		for kingQuiets != 0 {
			desSquare := popFunction(&kingQuiets)
			if isSquareSafe(desSquare, noKingFriendBoard, safetyCheckBoard, s.turn) {
				quiets = append(quiets, QuietMove{BuildMove(kingSquare, desSquare, 0, 0), historyTable[King+friendIndex][desSquare]})
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
					quiets = append(quiets, QuietMove{BuildMove(kingSquare, desSquare, 0, CastleSpecialMove), historyTable[King+friendIndex][desSquare]})
				}
			}
		}
		// King Castle End
		// Queen Castle Start
		if s.castleAvailability[s.turn+2] {
			if occupied&Bitboard(0xE<<rankIndex) == 0 && s.board[friendIndex+Rook]&Bitboard(0x1<<rankIndex) != 0 {
				desSquare := kingSquare - 2
				if isSquareSafe(Square(3+rankIndex), noKingFriendBoard, safetyCheckBoard, s.turn) && isSquareSafe(desSquare, noKingFriendBoard, safetyCheckBoard, s.turn) {
					quiets = append(quiets, QuietMove{BuildMove(kingSquare, desSquare, 0, CastleSpecialMove), historyTable[King+friendIndex][desSquare]})
				}
			}
		}
		// Queen Castle End
	}
	// Castle End
	// End King
	return &captures, &quiets
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

func genSliderMoves(s *State, piece uint8, board Bitboard, captures *[]CaptureMove, quiets *[]QuietMove, genInfo *MoveGenInfo, magicRetriever func(Square, Bitboard) Bitboard, pinnedBoard Bitboard, pinSafetys *[]PinSafety, checkBlockerSquares Bitboard) {
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
			attackedPiece := s.board.getPieceAt(attackSquare)
			*captures = append(*captures, CaptureMove{BuildMove(sliderSquare, attackSquare, 0, 0), valueTable[attackedPiece%6] - valueTable[piece]})
		}
		if genInfo.includeQuiets {
			sliderQuiets := sliderMoves & genInfo.notOccupied & safeSquares & checkBlockerSquares
			for sliderQuiets != 0 {
				quietSquare := genInfo.popFunction(&sliderQuiets)
				*quiets = append(*quiets, QuietMove{BuildMove(sliderSquare, quietSquare, 0, 0), historyTable[piece][quietSquare]})
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

func (s *State) GetCheckBlockerSquares(square Square, friendBoard Bitboard, enemyBoards *SafetyCheckBoards, turn uint8) (Bitboard, Bitboard) {
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
						return EmptyBitboard, EmptyBitboard
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
						return EmptyBitboard, EmptyBitboard
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
					return EmptyBitboard, EmptyBitboard
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
					return EmptyBitboard, EmptyBitboard
				}
			}
		}
	}
	enPassantCheckBlockerSquares := EmptyBitboard
	pawnCast := pawnAttackBoards[turn][square]
	if pawnCast&enemyBoards.pawnsBoard != 0 {
		for pawnCast != 0 {
			pawnSquare := PopLSB(&pawnCast)
			pawnBoard := Bitboard(1 << Bitboard(pawnSquare))
			if enemyBoards.pawnsBoard&pawnBoard != 0 {
				if !checkFound {
					if s.canEnpassant {
						enemyBackStep := UpStep
						if s.turn == Black {
							enemyBackStep = DownStep
						}
						if pawnSquare.Step(enemyBackStep) == s.enPassantSquare {
							enPassantCheckBlockerSquares |= Bitboard(1 << Bitboard(pawnSquare.Step(enemyBackStep)))
						}
					}
					safeSquares |= pawnBoard
					checkFound = true
				} else {
					return EmptyBitboard, EmptyBitboard
				}
			}
		}
	}
	return safeSquares, enPassantCheckBlockerSquares
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
	halfMoveClock, err := strconv.Atoi(splitFenString[4])
	if err != nil {
		panic("Invalid Fen String (Invalid half move clock)")
	}
	fullMoveNumber, err := strconv.Atoi(splitFenString[5])
	if err != nil {
		panic("Invalid Fen String (Invalid full move number)")
	}
	ply := uint16((fullMoveNumber - 1) * 2)
	if turn == Black {
		ply += 1
	}
	killerTable := make(KillerTable, 5)
	for i := range len(killerTable) {
		killerTable[i][0] = NilMove
		killerTable[i][1] = NilMove
	}
	searchParameters := SearchParameters{killerTable: &killerTable, trueDepth: -1}
	fiftyMoveRule := newFiftyMoveRuleHistory(104)
	repetitionMap := make(RepetitionMap, 50)
	hashHistory := NewHashHistory(5)
	s := &State{board: &board, turn: turn, enPassantSquare: enPassantSquare, check: false, captureHistory: NewCaptureHistory(32), canEnpassant: canEnpassant, enPassantSquareHistory: NewEnpassantHistory(16), lastCapOrPawn: uint16(halfMoveClock), ply: ply, castleAvailability: &castleAvailability, castleHistory: NewCastleHistory(4), fiftyMoveHistory: fiftyMoveRule, repetitionMap: &repetitionMap, hashHistory: hashHistory, searchParameters: searchParameters}
	s.hashcode = s.hash()
	s.hashHistory.Push(s.hashcode)
	return s
}

func StartingFen() *State {
	return FenState("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
}

func (s *State) fenString() string {
	output := ""
	pieceMap := map[uint8]rune{0: 'K', 1: 'Q', 2: 'R', 3: 'B', 4: 'N', 5: 'P', 6: 'k', 7: 'q', 8: 'r', 9: 'b', 10: 'n', 11: 'p'}
	for r := 7; r >= 0; r-- {
		rankString := ""
		emptySquares := 0
		for f := 0; f < 8; f++ {
			square := Square(r*8 + f)
			piece := s.board.getPieceAt(square)
			if piece != NoPiece {
				if emptySquares > 0 {
					rankString += strconv.FormatInt(int64(emptySquares), 10)
				}
				rankString += string(pieceMap[piece])
				emptySquares = 0
			} else {
				emptySquares += 1
			}
		}
		if emptySquares != 0 {
			rankString += strconv.FormatInt(int64(emptySquares), 10)
		}
		output += rankString
		if r != 0 {
			output += "/"
		}
	}
	if s.turn == White {
		output += " w "
	} else {
		output += " b "
	}
	castleString := ""
	if s.castleAvailability[0] {
		castleString += "K"
	}
	if s.castleAvailability[2] {
		castleString += "Q"
	}
	if s.castleAvailability[1] {
		castleString += "k"
	}
	if s.castleAvailability[3] {
		castleString += "q"
	}
	if castleString != "" {
		output += castleString + " "
	} else {
		output += "- "
	}
	if s.canEnpassant {
		output += s.enPassantSquare.String() + " "
	} else {
		output += "- "
	}
	output += strconv.FormatInt(int64(s.lastCapOrPawn), 10) + " "
	output += strconv.FormatInt(int64(s.ply/2+1), 10)
	return output
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

func (s *State) getPV() string {
	pvString := ""
	moveStack := make([]Move, 0, 30)
	var result TableData
	var found bool
	for {
		result, found = transpositionTable.SearchState(s)
		if found {
			bestMove := result.bestMove
			if bestMove != NilMove {
				pvString += bestMove.ShortString() + " "
				moveStack = append(moveStack, bestMove)
				s.MakeMove(bestMove)
			} else {
				break
			}
		} else {
			break
		}
	}
	stackPointer := len(moveStack) - 1
	for stackPointer >= 0 {
		s.UnMakeMove(moveStack[stackPointer])
		stackPointer--
	}
	return pvString
}
