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
	board     *Board
	boardInfo BoardInfo
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
	var boardInfo BoardInfo = 0
	turnString := splitFenString[1]
	if turnString == "b" {
		boardInfo |= 1
	} else if turnString != "w" {
		panic("Invalid Fen String (Invalid Turn)")
	}
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
	enpassantString := splitFenString[3]
	if enpassantString != "-" {
		rankMap := map[rune]int{'a': 0, 'b': 1, 'c': 2, 'd': 3, 'e': 4, 'f': 5, 'g': 6, 'h': 7}
		rankRune := enpassantString[0]
		rank, ok := rankMap[rune(rankRune)]
		if ok {
			boardInfo |= BoardInfo(rank) << 5
		} else {
			panic("Invalid Fen String (Invalid En Passant Square)")
		}
	}
	return &State{board: &board, boardInfo: boardInfo}
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
