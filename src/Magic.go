package main

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
