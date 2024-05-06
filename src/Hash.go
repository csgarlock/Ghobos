package main

import "math/rand"

var squareHashes [12][64]uint64
var blackHash uint64
var castleHashes [4]uint64
var enPassantHashes [8]uint64

func SetupHashRandoms() {
	for i := 0; i < 12; i++ {
		for j := 0; j < 64; j++ {
			squareHashes[i][j] = rand.Uint64()
		}
	}
	blackHash = rand.Uint64()
	for i := 0; i < 4; i++ {
		castleHashes[i] = rand.Uint64()
	}
	for i := 0; i < 8; i++ {
		enPassantHashes[i] = rand.Uint64()
	}
}

func (s *State) hash() uint64 {
	var hash uint64 = 0
	for i := 0; i < 12; i++ {
		board := s.board[i]
		for board != 0 {
			square := PopLSB(&board)
			hash ^= squareHashes[i][square]
		}
	}
	if s.turn == Black {
		hash ^= blackHash
	}
	for i := 0; i < 4; i++ {
		if s.castleAvailability[i] {
			hash ^= castleHashes[i]
		}
	}
	if s.canEnpassant {
		hash ^= enPassantHashes[s.enPassantSquare.File()]
	}
	return hash
}
