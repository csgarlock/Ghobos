package main

import "fmt"

// 0 - 5 origin square
// 6 - 11 destination square
// 12 - 13 Promotion type (0 - Queen, 1 - Rook, 2 - Knight, 3 - Bishop)
// 14 - 16 Special Move Type (0 - None, 1 - Castle, 2 - Promotion, 3 - En Passant)
type Move uint16

const (
	NilMove Move = 0xffff

	BitMask12 uint16 = 0xfff
	BitMask6  uint16 = 0x3f
	BitMask2  uint16 = 0x3

	CastleSpecialMove    = 1
	PromotionSpecialMove = 2
	EnPassantSpacialMove = 3

	QueenPromotion  = 0
	RookPromotion   = 1
	KnightPromotion = 2
	BishopPromotion = 3
)

func (m Move) OriginSquare() Square {
	return Square(m & Move(BitMask6))
}

func (m Move) DestinationSquare() Square {
	return Square(m >> 6 & Move(BitMask6))
}

func (m Move) PromotionType() uint16 {
	return uint16(m >> 12 & Move(BitMask2))
}

func (m Move) SpecialMove() uint16 {
	return uint16(m >> 14 & Move(BitMask2))
}

func BuildMove(origin Square, destination Square, promotion uint16, specialMove uint16) Move {
	return Move(origin) | (Move(destination) << 6) | (Move(promotion) << 12) | (Move(specialMove) << 14)
}

func sameSourceDes(move1 Move, move2 Move) bool {
	return (move1 & Move(BitMask12)) == (move2 & Move(BitMask12))
}

func (m Move) String() string {
	result := ""
	result += m.OriginSquare().String() + " To " + m.DestinationSquare().String() + "\n"
	result += fmt.Sprintf("Promotion Type: %b\n", m.PromotionType())
	result += fmt.Sprintf("Special Move Type: %b", m.SpecialMove())
	return result
}

func (m Move) ShortString() string {
	return m.OriginSquare().String() + m.DestinationSquare().String()
}
