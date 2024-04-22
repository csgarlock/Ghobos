package main

import (
	"fmt"
	"strconv"
	"strings"
)

// 0 turn (0 = White, 1 = Black)
// 1 white king side castle
// 2 white queen side castle
// 3 black king side castle
// 4 black queen side castle
// 5 - 7 the file of the en passant square
type BoardInfo uint8
type State struct {
	board           *Board
	turn            uint8 // 0 for White, 1 for Black
	enPassantSquare Square
	boardInfo       BoardInfo
}

// Stores important information about the current move being generated to more easily to functions
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
	var boardInfo BoardInfo = 0
	castleString := splitFenString[2]
	if castleString != "-" {
		castleOptions := [4]rune{'K', 'Q', 'k', 'q'}
		for i, r := range castleOptions {
			for _, c := range castleString {
				if r == c {
					boardInfo |= 1 << (i + 1)
				}
			}
		}
	}
	var enPassantSquare Square
	enpassantString := splitFenString[3]
	if enpassantString != "-" {
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
	return &State{board: &board, turn: turn, enPassantSquare: enPassantSquare, boardInfo: boardInfo}
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
	// Start Bishop
	bishopBoard := s.board[friendIndex+Bishop]
	genSliderMoves(bishopBoard, &moves, genInfo, getBishopMoves)
	// End Bishop
	// Start Knight
	knightBoard := s.board[friendIndex+Knight]
	for knightBoard != 0 {
		knightSquare := popFunction(&knightBoard)
		knightMoves := moveBoards[Knight][knightSquare]
		knightAttacks := knightMoves & enemyBoard
		for knightAttacks != 0 {
			attackSquare := popFunction(&knightAttacks)
			moves = append(moves, BuildMove(knightSquare, attackSquare, 0, 0))
		}
		if includeQuiets {
			knightQuiets := knightMoves & notOccupied
			for knightQuiets != 0 {
				quietSquare := popFunction(&knightQuiets)
				moves = append(moves, BuildMove(knightSquare, quietSquare, 0, 0))
			}
		}
	}
	// End Knight
	// Start Queen
	queenBoard := s.board[friendIndex+Queen]
	genSliderMoves(queenBoard, &moves, genInfo, getQueenMoves)
	// End Queen
	// Start Rook
	rookBoard := s.board[friendIndex+Rook]
	genSliderMoves(rookBoard, &moves, genInfo, getRookMoves)
	// End Rook
	// Start Pawn
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
		pawnAttacks := pawnAttackBoards[s.turn][pawnSquare] & enemyEnPassantBoard
		for pawnAttacks != 0 {
			attackSquare := popFunction(&pawnAttacks)
			if attackSquare == s.enPassantSquare {
				moves = append(moves, BuildMove(pawnSquare, attackSquare, 0, 3))
			} else {
				moves = append(moves, BuildMove(pawnSquare, attackSquare, 0, 0))
			}
		}
		if includeQuiets {
			pawnMoves := GetPawnMoves(pawnSquare, occupied, moveStep, homeRank)
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
	// End Pawn
	// Start King
	kingBoard := s.board[friendIndex+King]
	safetyCheckBoard := &SafetyCheckBoards{s.board[friendIndex+Bishop] | s.board[friendIndex+Queen], s.board[friendIndex+Rook] | s.board[friendIndex+Queen], s.board[friendIndex+Knight], kingBoard, s.board[friendIndex+Pawn], enemyBoard}
	noKingFriendBoard := friendBoard ^ kingBoard
	kingSquare := popFunction(&kingBoard)
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
			desSquare := popFunction(&kingAttacks)
			if isSquareSafe(desSquare, noKingFriendBoard, safetyCheckBoard, s.turn) {
				moves = append(moves, BuildMove(kingSquare, desSquare, 0, 0))
			}
		}
	}
	// End King
	return &moves
}

func genSliderMoves(board Bitboard, moves *[]Move, genInfo *MoveGenInfo, magicRetriever func(Square, Bitboard) Bitboard) {
	for board != 0 {
		sliderSquare := genInfo.popFunction(&board)
		sliderMoves := magicRetriever(sliderSquare, genInfo.occupied)
		sliderAttacks := sliderMoves & genInfo.enemyBoard
		for sliderAttacks != 0 {
			attackSquare := genInfo.popFunction(&sliderAttacks)
			*moves = append(*moves, BuildMove(sliderSquare, attackSquare, 0, 0))
		}
		if genInfo.includeQuiets {
			sliderQuiets := sliderMoves & genInfo.notOccupied
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
	pawnCast := pawnAttackBoards[1-turn][square]
	if pawnCast*enemyBoards.pawnsBoard != 0 {
		return true
	}
	return true
}

func (s *State) String() string {
	result := ""
	result += s.board.String() + "\n"
	result += s.boardInfo.String()
	return result
}

func (info BoardInfo) String() string {
	resultString := ""
	turn := info & 1
	if turn == 0 {
		resultString += "Turn: White\n"
	} else if turn == 1 {
		resultString += "Turn: Black\n"
	} else {
		resultString += "Invalid Turn"
	}
	resultString += fmt.Sprintf("Castle Status: %b\n", (info>>1)&0xf)
	ranks := [8]string{"a", "b", "c", "d", "e", "f", "g", "h"}
	resultString += "Enpassant Square: " + ranks[info>>5]
	return resultString
}
