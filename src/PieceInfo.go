package main

type Step int8

const (
	RightStep     Step = 1
	UpRightStep   Step = 9
	UpStep        Step = 8
	UpLeftStep    Step = 7
	LeftStep      Step = -1
	DownLeftStep  Step = -9
	DownStep      Step = -8
	DownRightStep Step = -7

	KnightStepRightUp   Step = 10
	KnightStepUpRight   Step = 17
	KnightStepUpLeft    Step = 15
	KnightStepLeftUp    Step = 6
	KnightStepLeftDown  Step = -10
	KnightStepDownLeft  Step = -17
	KnightStepDownRight Step = -15
	KnightStepRightDown Step = -6

	King   = 0
	Queen  = 1
	Rook   = 2
	Bishop = 3
	Knight = 4
	Pawn   = 5

	WhiteKing   = 0
	WhiteQueen  = 1
	WhiteRook   = 2
	WhiteBishop = 3
	WhiteKnight = 4
	WhitePawn   = 5

	BlackKing   = 6
	BlackQueen  = 7
	BlackRook   = 8
	BlackBishop = 9
	BlackKnight = 10
	BlackPawn   = 11
)

var stepMap map[Step]int = map[Step]int{RightStep: 0, UpRightStep: 1, UpStep: 2, UpLeftStep: 3, LeftStep: 4, DownLeftStep: 5, DownStep: 6, DownRightStep: 7, KnightStepRightUp: 8, KnightStepUpRight: 9, KnightStepUpLeft: 10, KnightStepLeftUp: 11, KnightStepLeftDown: 12, KnightStepDownLeft: 13, KnightStepDownRight: 14, KnightStepRightDown: 15}

var allSteps [16]Step = [16]Step{RightStep, UpRightStep, UpStep, UpLeftStep, LeftStep, DownLeftStep, DownStep, DownRightStep, KnightStepRightUp, KnightStepUpRight, KnightStepUpLeft, KnightStepLeftUp, KnightStepLeftDown, KnightStepDownLeft, KnightStepDownRight, KnightStepRightDown}
var kingSteps [8]Step = [8]Step{RightStep, UpRightStep, UpStep, UpLeftStep, LeftStep, DownLeftStep, DownStep, DownRightStep}
var queenSteps [8]Step = [8]Step{RightStep, UpRightStep, UpStep, UpLeftStep, LeftStep, DownLeftStep, DownStep, DownRightStep}
var bishopSteps [4]Step = [4]Step{UpRightStep, UpLeftStep, DownRightStep, DownLeftStep}
var rookSteps [4]Step = [4]Step{RightStep, UpStep, LeftStep, DownStep}
var knightSteps [8]Step = [8]Step{KnightStepRightUp, KnightStepUpRight, KnightStepUpLeft, KnightStepLeftUp, KnightStepLeftDown, KnightStepDownLeft, KnightStepDownRight, KnightStepRightDown}
