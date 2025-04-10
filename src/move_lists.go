package main

import (
	"sort"
)

type QuietMoveList struct {
	slice      []QuietMove
	firstEmpty uint16
}

type CaptureMoveList struct {
	slice      []CaptureMove
	firstEmpty uint16
}

type MoveListStack struct {
	moveLists []MoveList
	current   uint16
}

// Should switch to unsafe pointers at some point
// not going to just yet while still developing
type MoveList struct {
	slice         []Move
	firstEmpty    uint16
	searchPointer uint16
	// The first capture that has not been searched
	captureRemainsPointer uint16
	quietStartPinter      uint16
	secondCapturePass     bool
}

func newMoveListStack(stackSize uint16, sliceSize uint16) MoveListStack {
	moveStackList := MoveListStack{make([]MoveList, stackSize), 0}
	for i := range moveStackList.moveLists {
		moveStackList.moveLists[i].reset()
		moveStackList.moveLists[i].slice = make([]Move, sliceSize)
	}
	return moveStackList
}

func (moveStack *MoveListStack) incrementStack() {
	moveStack.current++
	if moveStack.current >= uint16(len(moveStack.moveLists)) {
		panic("Too Many MoveLists")
	}
}

func (moveStack *MoveListStack) decrementStack() {
	moveStack.current--
	if moveStack.current > uint16(len(moveStack.moveLists)) {
		panic("Too few MoveLists")
	}
}

func (moveStack *MoveListStack) resetCurrent() {
	moveStack.moveLists[moveStack.current].reset()
}

func (moveStack *MoveListStack) getCurrent() *MoveList {
	return &moveStack.moveLists[moveStack.current]
}

// Resets all pointers to 0 other than search pointer. Sets search pointer to
// 2 ^ 16 - 1. Reasoning for this is calling nextMove() as a loop condition will
// skip the first item if searchPointer starts at 0
func (moveList *MoveList) reset() {
	moveList.firstEmpty = 0
	moveList.searchPointer = uint16(65535)
	moveList.captureRemainsPointer = 0
	moveList.quietStartPinter = 0
}

func (moveList *MoveList) addMove(move Move) {
	moveList.slice[moveList.firstEmpty] = move
	moveList.firstEmpty++
}

// Moves the searchPointer to the next move and returns whether there is another move
func (moveList *MoveList) nextMove() bool {
	moveList.searchPointer++
	if !moveList.secondCapturePass {
		if moveList.searchPointer < moveList.firstEmpty {
			return true
		} else {
			if moveList.captureRemainsPointer == moveList.quietStartPinter {
				return false
			} else {
				moveList.secondCapturePass = true
				moveList.searchPointer = moveList.captureRemainsPointer
				return true
			}
		}
	} else {
		if moveList.searchPointer < moveList.quietStartPinter {
			return true
		} else {
			return false
		}
	}
}

func (moveList *MoveList) getMove() Move {
	return moveList.slice[moveList.searchPointer]
}

// Assumes the capture move currently at search pointer has not yet been searched
func (moveList *MoveList) setupQuietLoading() {
	moveList.captureRemainsPointer = moveList.searchPointer + 1
	moveList.searchPointer = moveList.firstEmpty - 1
	moveList.quietStartPinter = moveList.firstEmpty
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
