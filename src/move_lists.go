package main

import "sort"

type QuietMoveList struct {
	slice      []QuietMove
	firstEmpty uint16
}

type CaptureMoveList struct {
	slice      []CaptureMove
	firstEmpty uint16
}

func newQuietMoveList(size uint16) QuietMoveList {
	return QuietMoveList{make([]QuietMove, size), 0}
}

func newCaptureMoveList(size uint16) CaptureMoveList {
	return CaptureMoveList{make([]CaptureMove, size), 0}
}

func (moveList *QuietMoveList) size() uint16 {
	return uint16(len(moveList.slice))
}

func (moveList *CaptureMoveList) size() uint16 {
	return uint16(len(moveList.slice))
}

func (moveList *QuietMoveList) len() uint16 {
	return moveList.firstEmpty
}

func (moveList *CaptureMoveList) len() uint16 {
	return moveList.firstEmpty
}

func (moveList *QuietMoveList) addMove(quietMove QuietMove) {
	if moveList.firstEmpty >= moveList.size() {
		panic("Too Many Quiet Moves")
	}
	moveList.slice[moveList.firstEmpty] = quietMove
	moveList.firstEmpty++
}

func (moveList *CaptureMoveList) addMove(captureMove CaptureMove) {
	if moveList.firstEmpty >= moveList.size() {
		panic("Too Many Capture Moves")
	}
	moveList.slice[moveList.firstEmpty] = captureMove
	moveList.firstEmpty++
}

func (moveList *QuietMoveList) reset() {
	moveList.firstEmpty = 0
}

func (moveList *CaptureMoveList) reset() {
	moveList.firstEmpty = 0
}

func (moveList *QuietMoveList) sort() {
	sort.Slice(moveList.slice[:moveList.firstEmpty], func(i, j int) bool {
		return moveList.slice[i].historyValue > moveList.slice[j].historyValue
	})
}

func (moveList *CaptureMoveList) sort() {
	sort.Slice(moveList.slice[:moveList.firstEmpty], func(i, j int) bool {
		return moveList.slice[i].captureValue > moveList.slice[j].captureValue
	})
}
