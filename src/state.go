package main

import (
	"fmt"
	"strconv"
	"strings"
)

type KillerTable [][2]Move

// TODO migrate to worker when added instead of being linked to a state
type SearchParameters struct {
	killerTable KillerTable
	trueDepth   int16 // The true depth from the root node
}

type PinInfo struct {
	pinnedBoards [2]Bitboard
	pinners      [2][8]Square
	pinsSet      [2]bool
}

// 0 = White King Side, 1 = Black King Side, 2 = White Queen Side, 3 = Black Queen Side
type CastleAvailability [4]bool
type State struct {
	board                  Board
	sideOccupied           [2]Bitboard
	occupied               Bitboard
	notOccupied            Bitboard
	pinInfo                PinInfo
	turn                   uint8 // 0 for White, 1 for Black
	enPassantSquare        Square
	check                  bool
	captureHistory         CaptureHistory
	canEnpassant           bool
	enPassantSquareHistory EnPassantSquareHistory
	lastCapOrPawn          uint16
	ply                    uint16
	castleAvailability     CastleAvailability
	castleHistory          CastleHistory
	fiftyMoveHistory       FiftyMoveHistory
	repetitionMap          *RepetitionMap
	hashcode               uint64
	hashHistory            *HashHistory
	searchParameters       SearchParameters
}

type SafetyCheckBoards struct {
	bishopsBoards Bitboard
	rooksBoard    Bitboard
	knightsBoard  Bitboard
	kingBoard     Bitboard
	pawnsBoard    Bitboard
	combinedBoard Bitboard
}

type QuietMove struct {
	move         Move
	historyValue uint64
}

type CaptureMove struct {
	move         Move
	captureValue int32
}

// To be added to eventual worker struct
var quietMoves QuietMoveList = newQuietMoveList(100)
var captureMoves CaptureMoveList = newCaptureMoveList(50)

func (s *State) MakeMove(move Move) {
	s.lastCapOrPawn += 1
	friendIndex := s.turn * 6
	enemyIndex := (1 - s.turn) * 6
	if s.canEnpassant {
		s.hashcode ^= enPassantHashes[s.enPassantSquare%8]
	}
	if move != PassingMove {
		startSquare := move.OriginSquare()
		startBoard := boardFromSquare(startSquare)
		var startBoardPtr *Bitboard = nil
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
		desBoard := boardFromSquare(desSquare)
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
		s.sideOccupied[s.turn] ^= startBoard
		s.sideOccupied[s.turn] |= desBoard
		isCapture := false
		if desBoardPtr != nil {
			*desBoardPtr ^= desBoard
			s.sideOccupied[1-s.turn] ^= desBoard
			s.captureHistory.Push(uint8(desBoardIndex), s.ply)
			s.fiftyMoveHistory.Push(s.lastCapOrPawn-1, s.ply)
			s.lastCapOrPawn = 0
			isCapture = true
			s.hashcode ^= squareHashes[desBoardIndex][desSquare]
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
			startRookBoard := boardFromSquare(startingRookSquare)
			endRookBoard := boardFromSquare(rookSquare)
			s.board[friendIndex+Rook] ^= startRookBoard
			s.board[friendIndex+Rook] |= endRookBoard
			s.sideOccupied[s.turn] ^= startRookBoard
			s.sideOccupied[s.turn] |= endRookBoard
			s.hashcode ^= squareHashes[friendIndex+Rook][startingRookSquare] ^ squareHashes[friendIndex+Rook][rookSquare]
		} else if specialMove == PromotionSpecialMove {
			s.hashcode ^= squareHashes[startBoardIndex][desSquare]
			promotionType := move.PromotionType()
			if promotionType == QueenPromotion {
				s.board[friendIndex+Queen] |= desBoard
				s.hashcode ^= squareHashes[friendIndex+Queen][desSquare]
			} else if promotionType == RookPromotion {
				s.board[friendIndex+Rook] |= desBoard
				s.hashcode ^= squareHashes[friendIndex+Rook][desSquare]
			} else if promotionType == KnightPromotion {
				s.board[friendIndex+Knight] |= desBoard
				s.hashcode ^= squareHashes[friendIndex+Knight][desSquare]
			} else if promotionType == BishopPromotion {
				s.board[friendIndex+Bishop] |= desBoard
				s.hashcode ^= squareHashes[friendIndex+Bishop][desSquare]
			}
			*startBoardPtr ^= desBoard
		} else if specialMove == EnPassantSpacialMove {
			enemyPawnBoard := &s.board[enemyIndex+Pawn]
			relativeDownStep := DownStep
			if s.turn == Black {
				relativeDownStep = UpStep
			}
			enPassantCaptureSquare := desSquare.Step(relativeDownStep)
			enPassantCaptureBoard := boardFromSquare(enPassantCaptureSquare)
			*enemyPawnBoard ^= enPassantCaptureBoard
			s.sideOccupied[1-s.turn] ^= enPassantCaptureBoard
			s.hashcode ^= squareHashes[enemyIndex+Pawn][enPassantCaptureSquare]
			s.captureHistory.Push(enemyIndex+Pawn, s.ply)
		}
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
		s.occupied = s.sideOccupied[0] | s.sideOccupied[1]
		s.notOccupied = ^s.occupied
		s.pinInfo.pinsSet[0] = false
		s.pinInfo.pinsSet[1] = false
	} else {
		s.canEnpassant = false
		s.enPassantSquare = Square(100)
	}
	s.turn = 1 - s.turn
	s.hashcode ^= blackHash
	s.hashHistory.Push(s.hashcode)
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
	s.ply++
	s.repetitionMap.add(s.hashcode)
}

func (s *State) UnMakeMove(move Move) {
	s.repetitionMap.remove(s.hashcode)
	friendIndex := s.turn * 6
	enemyIndex := (1 - s.turn) * 6
	s.lastCapOrPawn -= 1
	if move != PassingMove {
		startSquare := move.OriginSquare()
		startBoard := boardFromSquare(startSquare)
		desSquare := move.DestinationSquare()
		desBoard := boardFromSquare(desSquare)
		var desBoardPtr *Bitboard = nil
		for i := enemyIndex; i < enemyIndex+6; i++ {
			board := &s.board[i]
			if *board&desBoard != 0 {
				desBoardPtr = board
			}
		}
		*desBoardPtr |= startBoard
		*desBoardPtr ^= desBoard
		s.sideOccupied[1-s.turn] |= startBoard
		s.sideOccupied[1-s.turn] ^= desBoard
		if s.ply-1 == s.captureHistory.MostRecentCapturePly() {
			capture := s.captureHistory.Pop()
			capturedPiece := capture.piece
			captureBoardPtr := &s.board[capturedPiece]
			*captureBoardPtr |= desBoard
			s.sideOccupied[s.turn] |= desBoard
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
			startingRookBoard := boardFromSquare(startingRookSquare)
			endingRookBoard := boardFromSquare(endingRookSquare)
			s.board[enemyIndex+Rook] ^= startingRookBoard
			s.board[enemyIndex+Rook] |= endingRookBoard
			s.sideOccupied[1-s.turn] ^= startingRookBoard
			s.sideOccupied[1-s.turn] |= endingRookBoard
		} else if specialMove == EnPassantSpacialMove {
			relativeUpStep := UpStep
			if s.turn == Black {
				relativeUpStep = DownStep
			}
			enPassantSquare := desSquare.Step(relativeUpStep)
			enPassantBoard := boardFromSquare(enPassantSquare)
			s.board[friendIndex+Pawn] ^= desBoard
			s.board[friendIndex+Pawn] |= enPassantBoard
			s.sideOccupied[s.turn] ^= desBoard
			s.sideOccupied[s.turn] |= enPassantBoard
		} else if specialMove == PromotionSpecialMove {
			*desBoardPtr ^= startBoard
			enemyPawnBoard := &s.board[enemyIndex+Pawn]
			*enemyPawnBoard |= startBoard
		}
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
	s.occupied = s.sideOccupied[0] | s.sideOccupied[1]
	s.notOccupied = ^s.occupied
	s.turn = 1 - s.turn
	s.pinInfo.pinsSet[0] = false
	s.pinInfo.pinsSet[1] = false
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

func (s *State) clearPins(perspective uint8) {
	s.pinInfo.pinnedBoards[perspective] = EmptyBitboard
	s.pinInfo.pinners[perspective] = [8]Square{NullSquare, NullSquare, NullSquare, NullSquare, NullSquare, NullSquare, NullSquare, NullSquare}
}

// Check for possible pins by interating through sliding pieces instead of by checking in all directions from the king like a doofus because 5 < 8
func (s *State) ensurePins(perspective uint8) {
	if !s.pinInfo.pinsSet[perspective] {
		s.clearPins(perspective)
		enemyIndex := uint8((1 - perspective) * 6)
		kingSquare := GetLSB(s.board[(6*perspective)+King])
		bishopBoard := s.board[enemyIndex+Bishop]
		for bishopBoard != EmptyBitboard {
			bishopSquare := PopLSB(&bishopBoard)
			step := squareToSquareStep[kingSquare][bishopSquare]
			if step != 0 {
				stepID := getStepId(step)
				if stepID%2 == 1 {
					rayCast := squareToSquareFillBoards[kingSquare][bishopSquare]
					intersection := rayCast & s.sideOccupied[perspective]
					if BitCount(intersection) == 1 && rayCast&s.sideOccupied[1-perspective] == 0 {
						s.pinInfo.pinnedBoards[perspective] |= intersection
						s.pinInfo.pinners[perspective][stepID] = bishopSquare
					}
				}
			}
		}
		rookBoard := s.board[enemyIndex+Rook]
		for rookBoard != EmptyBitboard {
			rookSquare := PopLSB(&rookBoard)
			step := squareToSquareStep[kingSquare][rookSquare]
			if step != 0 {
				stepID := getStepId(step)
				if stepID%2 == 0 {
					rayCast := squareToSquareFillBoards[kingSquare][rookSquare]
					intersection := rayCast & s.sideOccupied[perspective]
					if BitCount(intersection) == 1 && rayCast&s.sideOccupied[1-perspective] == 0 {
						s.pinInfo.pinnedBoards[perspective] |= intersection
						s.pinInfo.pinners[perspective][stepID] = rookSquare
					}
				}
			}
		}
		queenBoard := s.board[enemyIndex+Queen]
		for queenBoard != EmptyBitboard {
			queenSquare := PopLSB(&queenBoard)
			step := squareToSquareStep[kingSquare][queenSquare]
			if step != 0 {
				rayCast := squareToSquareFillBoards[kingSquare][queenSquare]
				intersection := rayCast & s.sideOccupied[perspective]
				if BitCount(intersection) == 1 && rayCast&s.sideOccupied[1-perspective] == 0 {
					s.pinInfo.pinnedBoards[perspective] |= intersection
					s.pinInfo.pinners[perspective][getStepId(step)] = queenSquare
				}
			}
		}
	}
}

func (s *State) quickGenMoves() *[]Move {
	s.genAllMoves(true)
	moves := make([]Move, captureMoves.len()+quietMoves.len())
	totalIndex := 0
	badCutoff := captureMoves.len()
	for i := uint16(0); i < captureMoves.len(); i++ {
		if captureMoves.slice[i].captureValue >= 0 {
			moves[totalIndex] = captureMoves.slice[i].move
			totalIndex++
		} else {
			badCutoff = i
			break
		}
	}
	for i := 0; i < int(quietMoves.len()); i++ {
		moves[totalIndex] = quietMoves.slice[i].move
		totalIndex++
	}
	for i := badCutoff; i < captureMoves.len(); i++ {
		moves[totalIndex] = captureMoves.slice[i].move
		totalIndex++
	}
	return &moves
}

func (s *State) genAllMoves(includeQuiets bool) {
	// We want the pop function to pop the the bits at the top of the board relative to whos turn
	// it is. So when it's white's turn we pop the most significant bit first and with black
	// we pop the least significant bit first
	s.ensurePins(s.turn)
	quietMoves.reset()
	captureMoves.reset()
	var friendIndex uint8 = s.turn * 6
	var enemyIndex uint8 = (1 - s.turn) * 6
	friendBoard := s.sideOccupied[s.turn]
	enemyBoard := s.sideOccupied[1-s.turn]
	occupied := s.occupied
	notOccupied := ^occupied
	kingBoard := s.board[friendIndex+King]
	kingSquare := PopLSB(&kingBoard)
	enemyBishopSliders := s.board[enemyIndex+Bishop] | s.board[enemyIndex+Queen]
	enemyRookSliders := s.board[enemyIndex+Rook] | s.board[enemyIndex+Queen]
	safetyCheckBoard := &SafetyCheckBoards{enemyBishopSliders, enemyRookSliders, s.board[enemyIndex+Knight], s.board[enemyIndex+King], s.board[enemyIndex+Pawn], enemyBoard}
	checkBlockerSquares := UniversalBitboard
	enPassantCheckBlockerSquares := UniversalBitboard
	if s.check {
		checkBlockerSquares, enPassantCheckBlockerSquares = s.GetCheckBlockerSquares(kingSquare, friendBoard, safetyCheckBoard, s.turn)
	}
	// Start Bishop
	if checkBlockerSquares != EmptyBitboard {
		bishopBoard := s.board[friendIndex+Bishop]
		for bishopBoard != 0 {
			sliderSquare := PopLSB(&bishopBoard)
			safeSquares := s.getPinBoard(sliderSquare, kingSquare, s.turn)
			sliderMoves := getBishopMoves(sliderSquare, s.occupied)
			sliderAttacks := sliderMoves & s.sideOccupied[1-s.turn] & safeSquares & checkBlockerSquares
			for sliderAttacks != 0 {
				attackSquare := PopLSB(&sliderAttacks)
				attackedPiece := s.board.getPieceAt(attackSquare)
				captureMoves.addMove(CaptureMove{BuildSimpleMove(sliderSquare, attackSquare), valueTable[attackedPiece%6] - valueTable[Bishop]})
			}
			if includeQuiets {
				sliderQuiets := sliderMoves & s.notOccupied & safeSquares & checkBlockerSquares
				for sliderQuiets != 0 {
					quietSquare := PopLSB(&sliderQuiets)
					quietMoves.addMove(QuietMove{BuildSimpleMove(sliderSquare, quietSquare), historyTable[friendIndex+Bishop][quietSquare]})
				}
			}
		}
	}
	// End Bishop
	// Start Knight
	if checkBlockerSquares != EmptyBitboard {
		pieceIndex := Knight + friendIndex
		knightBoard := s.board[pieceIndex]
		for knightBoard != 0 {
			knightSquare := PopLSB(&knightBoard)
			if s.pinInfo.pinnedBoards[s.turn]&(1<<Bitboard(knightSquare)) == 0 {
				knightMoves := moveBoards[Knight][knightSquare]
				knightAttacks := knightMoves & enemyBoard & checkBlockerSquares
				for knightAttacks != 0 {
					attackSquare := PopLSB(&knightAttacks)
					attackedPiece := s.board.getColorPieceAt(attackSquare, 1-s.turn)
					captureMoves.addMove(CaptureMove{BuildSimpleMove(knightSquare, attackSquare), valueTable[attackedPiece%6] - valueTable[Knight]})
				}
				if includeQuiets {
					knightQuiets := knightMoves & notOccupied & checkBlockerSquares
					for knightQuiets != 0 {
						quietSquare := PopLSB(&knightQuiets)
						quietMoves.addMove(QuietMove{BuildSimpleMove(knightSquare, quietSquare), historyTable[pieceIndex][quietSquare]})
					}
				}
			}
		}
	}
	// End Knight
	// Start Queen
	if checkBlockerSquares != EmptyBitboard {
		queenBoard := s.board[friendIndex+Queen]
		for queenBoard != 0 {
			sliderSquare := PopLSB(&queenBoard)
			safeSquares := s.getPinBoard(sliderSquare, kingSquare, s.turn)
			sliderMoves := getQueenMoves(sliderSquare, s.occupied)
			sliderAttacks := sliderMoves & s.sideOccupied[1-s.turn] & safeSquares & checkBlockerSquares
			for sliderAttacks != 0 {
				attackSquare := PopLSB(&sliderAttacks)
				attackedPiece := s.board.getPieceAt(attackSquare)
				captureMoves.addMove(CaptureMove{BuildSimpleMove(sliderSquare, attackSquare), valueTable[attackedPiece%6] - valueTable[Queen]})
			}
			if includeQuiets {
				sliderQuiets := sliderMoves & s.notOccupied & safeSquares & checkBlockerSquares
				for sliderQuiets != 0 {
					quietSquare := PopLSB(&sliderQuiets)
					quietMoves.addMove(QuietMove{BuildSimpleMove(sliderSquare, quietSquare), historyTable[friendIndex+Queen][quietSquare]})
				}
			}
		}
	}
	// End Queen
	// Start Rook
	if checkBlockerSquares != EmptyBitboard {
		rookBoard := s.board[friendIndex+Rook]
		for rookBoard != 0 {
			sliderSquare := PopLSB(&rookBoard)
			safeSquares := s.getPinBoard(sliderSquare, kingSquare, s.turn)
			sliderMoves := getRookMoves(sliderSquare, s.occupied)
			sliderAttacks := sliderMoves & s.sideOccupied[1-s.turn] & safeSquares & checkBlockerSquares
			for sliderAttacks != 0 {
				attackSquare := PopLSB(&sliderAttacks)
				attackedPiece := s.board.getPieceAt(attackSquare)
				captureMoves.addMove(CaptureMove{BuildSimpleMove(sliderSquare, attackSquare), valueTable[attackedPiece%6] - valueTable[Rook]})
			}
			if includeQuiets {
				sliderQuiets := sliderMoves & s.notOccupied & safeSquares & checkBlockerSquares
				for sliderQuiets != 0 {
					quietSquare := PopLSB(&sliderQuiets)
					quietMoves.addMove(QuietMove{BuildSimpleMove(sliderSquare, quietSquare), historyTable[friendIndex+Rook][quietSquare]})
				}
			}
		}
	}
	// End Rook
	// Start Pawn
	if checkBlockerSquares != EmptyBitboard {
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
			pawnSquare := PopLSB(&pawnBoard)
			safeBoard := s.getPinBoard(pawnSquare, kingSquare, s.turn)
			pawnAttacks := pawnAttackBoards[s.turn][pawnSquare] & enemyEnPassantBoard & safeBoard & (checkBlockerSquares | enPassantCheckBlockerSquares)
			for pawnAttacks != 0 {
				attackSquare := PopLSB(&pawnAttacks)
				attackedPiece := s.board.getColorPieceAt(attackSquare, 1-s.turn)
				if attackSquare == s.enPassantSquare {
					if s.canEnpassant {
						if s.EnPassantSafetyCheck(pawnSquare, attackSquare, friendIndex, enemyIndex, occupied) {
							captureMoves.addMove(CaptureMove{BuildMove(pawnSquare, attackSquare, 0, EnPassantSpacialMove), 0})
						}
					}
				} else {
					attackValue := valueTable[attackedPiece%6] - valueTable[Pawn]
					if attackSquare.Rank() == promotionRank {
						captureMoves.addMove(CaptureMove{BuildMove(pawnSquare, attackSquare, 0, PromotionSpecialMove), attackValue})
						captureMoves.addMove(CaptureMove{BuildMove(pawnSquare, attackSquare, 1, PromotionSpecialMove), attackValue})
						captureMoves.addMove(CaptureMove{BuildMove(pawnSquare, attackSquare, 2, PromotionSpecialMove), attackValue})
						captureMoves.addMove(CaptureMove{BuildMove(pawnSquare, attackSquare, 3, PromotionSpecialMove), attackValue})
					} else {
						captureMoves.addMove(CaptureMove{BuildMove(pawnSquare, attackSquare, 0, 0), attackValue})
					}
				}
			}
			if includeQuiets {
				pawnMoves := GetPawnMoves(pawnSquare, occupied, moveStep, homeRank) & safeBoard & checkBlockerSquares
				for pawnMoves != 0 {
					desSquare := PopLSB(&pawnMoves)
					historyValue := historyTable[Pawn+friendIndex][desSquare]
					if desSquare.Rank() == promotionRank {
						quietMoves.addMove(QuietMove{BuildMove(pawnSquare, desSquare, 0, PromotionSpecialMove), historyValue})
						quietMoves.addMove(QuietMove{BuildMove(pawnSquare, desSquare, 1, PromotionSpecialMove), historyValue})
						quietMoves.addMove(QuietMove{BuildMove(pawnSquare, desSquare, 2, PromotionSpecialMove), historyValue})
						quietMoves.addMove(QuietMove{BuildMove(pawnSquare, desSquare, 3, PromotionSpecialMove), historyValue})
					} else {
						quietMoves.addMove(QuietMove{BuildMove(pawnSquare, desSquare, 0, 0), historyValue})
					}
				}
			}

		}
	}
	// End Pawn
	// Start King
	kingBoard = s.board[friendIndex+King]
	noKingFriendBoard := friendBoard ^ kingBoard
	kingSquare = PopLSB(&kingBoard)
	kingAttacks := moveBoards[King][kingSquare] & enemyBoard
	for kingAttacks != 0 {
		desSquare := PopLSB(&kingAttacks)
		attackedPiece := s.board.getColorPieceAt(desSquare, 1-s.turn)
		if isSquareSafe(desSquare, noKingFriendBoard, safetyCheckBoard, s.turn) {
			captureMoves.addMove(CaptureMove{BuildSimpleMove(kingSquare, desSquare), valueTable[attackedPiece%6] - valueTable[King]})
		}
	}
	if includeQuiets {
		kingQuiets := moveBoards[King][kingSquare] & notOccupied
		for kingQuiets != 0 {
			desSquare := PopLSB(&kingQuiets)
			if isSquareSafe(desSquare, noKingFriendBoard, safetyCheckBoard, s.turn) {
				quietMoves.addMove(QuietMove{BuildSimpleMove(kingSquare, desSquare), historyTable[King+friendIndex][desSquare]})
			}
		}
	}
	// Castle Start
	if !s.check && includeQuiets {
		rankIndex := s.turn * 56
		// King Castle Start
		if s.castleAvailability[s.turn] {
			if occupied&Bitboard(0x60<<rankIndex) == 0 && s.board[friendIndex+Rook]&Bitboard(0x80<<rankIndex) != 0 {
				desSquare := kingSquare + 2
				if isSquareSafe(Square(5+rankIndex), noKingFriendBoard, safetyCheckBoard, s.turn) && isSquareSafe(desSquare, noKingFriendBoard, safetyCheckBoard, s.turn) {
					quietMoves.addMove(QuietMove{BuildMove(kingSquare, desSquare, 0, CastleSpecialMove), historyTable[King+friendIndex][desSquare]})
				}
			}
		}
		// King Castle End
		// Queen Castle Start
		if s.castleAvailability[s.turn+2] {
			if occupied&Bitboard(0xE<<rankIndex) == 0 && s.board[friendIndex+Rook]&Bitboard(0x1<<rankIndex) != 0 {
				desSquare := kingSquare - 2
				if isSquareSafe(Square(3+rankIndex), noKingFriendBoard, safetyCheckBoard, s.turn) && isSquareSafe(desSquare, noKingFriendBoard, safetyCheckBoard, s.turn) {
					quietMoves.addMove(QuietMove{BuildMove(kingSquare, desSquare, 0, CastleSpecialMove), historyTable[King+friendIndex][desSquare]})
				}
			}
		}
		// Queen Castle End
	}
	// Castle End
	// End King
}

// Given a square returns a Bitboard with all the squares that the piece can move to and not leave the king expose
func (s *State) getPinBoard(pinnedSquare Square, kingSquare Square, perspective uint8) Bitboard {
	if s.pinInfo.pinnedBoards[perspective]&boardFromSquare(pinnedSquare) != 0 {
		pinnerSquare := s.pinInfo.pinners[perspective][getStepId(squareToSquareStep[kingSquare][pinnedSquare])]
		return squareToSquareFillBoards[pinnedSquare][kingSquare] | squareToSquareFillBoards[pinnedSquare][pinnerSquare] | boardFromSquare(pinnerSquare)
	}
	return UniversalBitboard
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
	sideOccupied := [2]Bitboard{EmptyBitboard, EmptyBitboard}
	occupied := EmptyBitboard
	for i := 0; i < 6; i++ {
		sideOccupied[0] |= board[i]
		sideOccupied[1] |= board[6+i]
	}
	occupied = sideOccupied[0] | sideOccupied[1]
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
	searchParameters := SearchParameters{killerTable: killerTable, trueDepth: -1}
	fiftyMoveRule := newFiftyMoveRuleHistory(104)
	repetitionMap := make(RepetitionMap, 50)
	hashHistory := NewHashHistory(5)
	pinInfo := PinInfo{pinnedBoards: [2]Bitboard{}, pinners: [2][8]Square{}, pinsSet: [2]bool{false, false}}
	s := &State{
		board:                  board,
		sideOccupied:           sideOccupied,
		occupied:               occupied,
		notOccupied:            ^occupied,
		pinInfo:                pinInfo,
		turn:                   turn,
		enPassantSquare:        enPassantSquare,
		check:                  false,
		captureHistory:         NewCaptureHistory(32),
		canEnpassant:           canEnpassant,
		enPassantSquareHistory: NewEnpassantHistory(16),
		lastCapOrPawn:          uint16(halfMoveClock),
		ply:                    ply,
		castleAvailability:     castleAvailability,
		castleHistory:          NewCastleHistory(4),
		fiftyMoveHistory:       fiftyMoveRule,
		repetitionMap:          &repetitionMap,
		hashHistory:            hashHistory,
		searchParameters:       searchParameters,
	}
	s.hashcode = s.hash()
	s.hashHistory.Push(s.hashcode)
	s.ensurePins(White)
	s.ensurePins(Black)
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
	result += fmt.Sprintf("Castle Availability: %v", s.castleAvailability)
	return result + "\n"
}

func (s *State) getPV() string {
	pvString := ""
	moveStack := make([]Move, 0, 30)
	var result TableData
	var found bool
	seenBoards := map[uint64]uint64{}
	for {
		result, found = transpositionTable.SearchState(s)
		_, dupFound := seenBoards[s.hashcode]
		if found && !dupFound {
			bestMove := result.bestMove
			if bestMove != NilMove {
				pvString += bestMove.ShortString() + " "
				moveStack = append(moveStack, bestMove)
				seenBoards[s.hashcode] = s.hashcode
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
