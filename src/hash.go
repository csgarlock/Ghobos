package main

import "math/rand"

type RepetitionMap map[uint64]uint16

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

func (repetitionMap *RepetitionMap) add(hash uint64) {
	_, exists := (*repetitionMap)[hash]
	if exists {
		(*repetitionMap)[hash] += 1
	} else {
		(*repetitionMap)[hash] = 1
	}
}

func (repetitionMap *RepetitionMap) remove(hash uint64) {
	value, exists := (*repetitionMap)[hash]
	if exists {
		if value == 1 {
			delete(*repetitionMap, hash)
		} else {
			(*repetitionMap)[hash] -= 1
		}
	} else {
		panic("Can't remove from repetition hash. is not in table")
	}
}

func (repetitionMap *RepetitionMap) get(hash uint64) uint16 {
	return (*repetitionMap)[hash]
}
