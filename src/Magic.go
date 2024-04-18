package main

import (
	"math"
	"math/bits"
	"math/rand"
)

type Magic struct {
	mask        Bitboard
	magicNumber uint64
	index       uint8
	offset      uint32
}

var bishopMagics [64]Magic = [64]Magic{}
var rookMagics [64]Magic = [64]Magic{}

var bishopTable [0x1480]Bitboard = [0x1480]Bitboard{}
var rookTable [0x19000]Bitboard = [0x19000]Bitboard{}

func InitializeMagics() {
	var square Square
	var runningBishopOffset uint32 = 0
	var runningRookOffset uint32 = 0
	for square = 0; square < 64; square++ {
		bishopMagics[square] = Magic{0, 0, 0, 0}
		rookMagics[square] = Magic{0, 0, 0, 0}
		squareBishopTable, bishopShift := FindMagic(square, &moveBoards[Bishop], &bishopMagics[square], &bishopSteps)
		bishopMagics[square].offset = runningBishopOffset
		squareRookTable, rookShift := FindMagic(square, &moveBoards[Rook], &rookMagics[square], &rookSteps)
		rookMagics[square].offset = runningRookOffset
		for i, board := range *squareBishopTable {
			bishopTable[i+int(runningBishopOffset)] = board
		}
		for i, board := range *squareRookTable {
			rookTable[i+int(runningRookOffset)] = board
		}
		runningBishopOffset += bishopShift
		runningRookOffset += rookShift
	}
}

func FindMagic(s Square, attacks *[64]Bitboard, magic *Magic, pieceSteps *[4]Step) (*[]Bitboard, uint32) {
	attackBoard := attacks[s]
	mask := attackBoard & (^(Rank0 | Rank7) | ranks[s.Rank()]) & (^(File0 | File7) | files[s.File()])
	magic.mask = mask
	bitcount := bits.OnesCount64(uint64(mask))
	magic.index = uint8(bitcount)
	tableSize := uint32(math.Pow(2, float64(bitcount)))
	table := make([]Bitboard, tableSize)
	moveTable := make([]Bitboard, tableSize)
	subIter := NewSubsetIterator(mask)
	for i := range tableSize {
		moveTable[i] = findBlockedSlidingAttack(s, pieceSteps, subIter.getSubset())
		subIter.nextSubset()
	}
	for {
		magicNum := rand.Uint64() & rand.Uint64() & rand.Uint64()
		magic.magicNumber = magicNum
		foundTable := make([]bool, tableSize)
		subsetIterator := NewSubsetIterator(mask)
		goodTable := true
		for i := 0; i < int(tableSize); i++ {
			moves := moveTable[i]
			tableIndex := GetMagicIndex(magic, subsetIterator.getSubset())
			if foundTable[tableIndex] {
				goodTable = false
				break
			} else {
				table[tableIndex] = moves
				foundTable[tableIndex] = true
			}
			if subsetIterator.nextSubset() == 0 {
				break
			}
		}
		if goodTable {
			return &table, tableSize
		}

	}
}

func GetMagicIndex(magic *Magic, occupied Bitboard) int64 {
	blockers := occupied & magic.mask
	hash := blockers * Bitboard(magic.magicNumber)
	return int64((hash >> (64 - magic.index)) + Bitboard(magic.offset))
}

func getBishopMoves(square Square, occupied Bitboard) Bitboard {
	magic := &bishopMagics[square]
	return bishopTable[GetMagicIndex(magic, occupied)]
}

func getRookMoves(square Square, occupied Bitboard) Bitboard {
	magic := &rookMagics[square]
	return rookTable[GetMagicIndex(magic, occupied)]
}
