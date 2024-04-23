package main

type Capture struct {
	piece uint8
	ply   uint16
}
type CaptureHistory struct {
	slice        []Capture
	currentIndex int32
}

func NewCaptureHistory(startingLength int32) *CaptureHistory {
	slice := make([]Capture, startingLength)
	return &CaptureHistory{slice: slice, currentIndex: 0}
}

func (pH *CaptureHistory) MostRecentCapturePly() uint16 {
	if pH.currentIndex == 0 {
		return 65535
	}
	return pH.slice[pH.currentIndex-1].ply
}

func (pH *CaptureHistory) Pop() Capture {
	pH.currentIndex--
	if pH.currentIndex < 0 {
		panic("Can't pop empty PieceHistory")
	}
	capture := pH.slice[pH.currentIndex]
	return capture
}

func (pH *CaptureHistory) Push(piece uint8, ply uint16) {
	pH.slice[pH.currentIndex] = Capture{piece: piece, ply: ply}
	pH.currentIndex++
}
