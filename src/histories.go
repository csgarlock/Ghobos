package main

type Capture struct {
	piece uint8
	ply   uint16
}
type CaptureHistory struct {
	slice        []Capture
	currentIndex int32
}
type EnPassantEntry struct {
	square Square
	ply    uint16
}
type EnPassantSquareHistory struct {
	slice        []EnPassantEntry
	currentIndex int32
}
type CastleHistoryEntry struct {
	castle uint8
	ply    uint16
}
type CastleHistory struct {
	slice        []CastleHistoryEntry
	currentIndex int32
}
type FiftyMoveEntry struct {
	lastCount uint16
	ply       uint16
}
type FiftyMoveHistory struct {
	slice        []FiftyMoveEntry
	currentIndex int32
}
type HashHistory struct {
	slice        []uint64
	currentIndex int32
}

func NewCaptureHistory(startingLength int32) CaptureHistory {
	slice := make([]Capture, startingLength)
	return CaptureHistory{slice: slice, currentIndex: 0}
}

func (pH *CaptureHistory) MostRecentCapturePly() uint16 {
	if pH.currentIndex == 0 {
		return 65530
	}
	return pH.slice[pH.currentIndex-1].ply
}

func (pH *CaptureHistory) Pop() Capture {
	pH.currentIndex--
	capture := pH.slice[pH.currentIndex]
	return capture
}

func (pH *CaptureHistory) Push(piece uint8, ply uint16) {
	pH.slice[pH.currentIndex] = Capture{piece: piece, ply: ply}
	pH.currentIndex++
}

func NewEnpassantHistory(startingLength int32) EnPassantSquareHistory {
	slice := make([]EnPassantEntry, startingLength)
	return EnPassantSquareHistory{slice: slice, currentIndex: 0}
}

func (eH *EnPassantSquareHistory) MostRecentCapturePly() uint16 {
	if eH.currentIndex == 0 {
		return 65530
	}
	return eH.slice[eH.currentIndex-1].ply
}

func (eH *EnPassantSquareHistory) Pop() EnPassantEntry {
	eH.currentIndex--
	capture := eH.slice[eH.currentIndex]
	return capture
}

func (eH *EnPassantSquareHistory) Peek() EnPassantEntry {
	return eH.slice[eH.currentIndex-1]
}

func (eH *EnPassantSquareHistory) Push(square Square, ply uint16) {
	eH.slice[eH.currentIndex] = EnPassantEntry{square: square, ply: ply}
	eH.currentIndex++
}

func NewCastleHistory(startingLength int32) CastleHistory {
	slice := make([]CastleHistoryEntry, startingLength)
	return CastleHistory{slice: slice, currentIndex: 0}
}

func (cH *CastleHistory) MostRecentCapturePly() uint16 {
	if cH.currentIndex == 0 {
		return 65530
	}
	return cH.slice[cH.currentIndex-1].ply
}

func (cH *CastleHistory) Pop() CastleHistoryEntry {
	cH.currentIndex--
	capture := cH.slice[cH.currentIndex]
	return capture
}

func (cH *CastleHistory) Push(castle uint8, ply uint16) {
	cH.slice[cH.currentIndex] = CastleHistoryEntry{castle: castle, ply: ply}
	cH.currentIndex++
}

func newFiftyMoveRuleHistory(startingLength int32) FiftyMoveHistory {
	slice := make([]FiftyMoveEntry, startingLength)
	return FiftyMoveHistory{slice: slice, currentIndex: 0}
}

func (fH *FiftyMoveHistory) lastReset() uint16 {
	if fH.currentIndex == 0 {
		return 65530
	}
	return fH.slice[fH.currentIndex-1].ply
}

func (fH *FiftyMoveHistory) Pop() FiftyMoveEntry {
	fH.currentIndex--
	return fH.slice[fH.currentIndex]
}

func (fH *FiftyMoveHistory) Peek() FiftyMoveEntry {
	return fH.slice[fH.currentIndex-1]
}

func (fH *FiftyMoveHistory) Push(lastCapOrPawn uint16, ply uint16) {
	fH.slice[fH.currentIndex] = FiftyMoveEntry{lastCapOrPawn, ply}
	fH.currentIndex++
}

func NewHashHistory(startingLength int32) *HashHistory {
	slice := make([]uint64, startingLength)
	return &HashHistory{slice: slice, currentIndex: 0}
}

func (hH *HashHistory) Pop() uint64 {
	hH.currentIndex--
	return hH.slice[hH.currentIndex]
}

func (hH *HashHistory) Peek() uint64 {
	return hH.slice[hH.currentIndex-1]
}

func (hH *HashHistory) Push(hash uint64) {
	if hH.currentIndex > int32(len(hH.slice)-1) {
		hH.slice = append(hH.slice, hash)
	} else {
		hH.slice[hH.currentIndex] = hash
	}
	hH.currentIndex++
}
