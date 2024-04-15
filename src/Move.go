package main

import "fmt"

// 0 - 5 origin square
// 6 - 11 destination square
// 12 - 13 Promotion type (0 - Queen, 1 - Rook, 2 Knight, 3 - Bishop)
// 14 - 16 Special Move Type (0 - None, 1 - Castle, 2 - Promotion, 3 - En Passant)
type Move uint16

const (
	BitMask6 uint16 = 0x3f
	BitMask2 uint16 = 0x3

	CastleSpecialMove    = 0
	PromotionSpecialMove = 1
	EnPassantSpacialMove = 2
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

func (m Move) String() string {
	result := ""
	result += fmt.Sprintf("Source: Rank = %d, File = %d\n", m.OriginSquare().Rank(), m.OriginSquare().File())
	result += fmt.Sprintf("Destination: Rank = %d, File = %d\n", m.DestinationSquare().Rank(), m.DestinationSquare().File())
	result += fmt.Sprintf("Promotion Type: %b\n", m.PromotionType())
	result += fmt.Sprintf("Special Move Type: %b", m.SpecialMove())
	return result
}
