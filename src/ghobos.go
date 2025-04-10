package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"time"
)

func main() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()
	InitializeMoveBoards()
	InitializeEvalVariables()
	setupFillBoards()
	// s := FenState("rnbqk1nr/pppp1ppp/8/4p3/1b1P4/2N5/PPP1PPPP/R1BQKBNR w KQkq - 2 3")
	// fmt.Println(s)
	// fmt.Println(s.MakeMove(BuildSimpleMove(SFS("c3"), SFS("d5"))))
	// fmt.Println(s)
	// s := FenState("rnbqkb1r/pppppp1p/7n/6pQ/8/4P3/PPPP1PPP/RNB1KBNR b KQkq g6 0 2")
	// fmt.Println(s)
	// fmt.Println(s.isSquareSafeEasy(GetLSB(s.board[BlackKing])))
	// s.MakeMove((SimpleMoveFromString("e2e4")))
	// s := FenState("8/8/8/KPpPr3/5p1k/8/6P1/1R6 w - c6 0 4")
	// s.NewGenMoves(true, UniversalBitboard)
	// for moveStack.getCurrent().nextMove() {
	// 	fmt.Println(moveStack.getCurrent().getMove())
	// }
	// PerftCheckerNew(7, FenState("8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1"))
	PerftTester()
	// SetupTable(4096)
	// UIGame()
}

func UIGame() {
	var playerSide uint8
	for {
		fmt.Print("What Color do you want (white/black): ")
		var colorInput string
		_, err := fmt.Scanln(&colorInput)
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}
		colorInput = strings.ToLower(colorInput)
		if colorInput == "white" || colorInput == "w" {
			playerSide = White
			break
		} else if colorInput == "black" || colorInput == "b" {
			playerSide = Black
			break
		}
	}
	gameState := FenState("6k1/6p1/8/6KQ/1r6/q2b4/8/8 w - - 0 1")
	gameOver := false
	playerTurn := false
	if playerSide == gameState.turn {
		playerTurn = true
	}
	for !gameOver {
		fmt.Println(gameState)
		fmt.Println(gameState.fenString())
		if playerTurn {
			for {
				playerMove := getUserMove()
				validMoves := gameState.quickGenMoves()
				found := false
				var foundMove Move
				for _, move := range *validMoves {
					if sameSourceDes(playerMove, Move(move)) {
						found = true
						foundMove = move
					}
				}
				if found {
					if foundMove.SpecialMove() == PromotionSpecialMove {
						for {
							fmt.Print("What piece do you want to promote to (queen/rook/bishop/knight): ")
							var promotionString string
							_, err := fmt.Scanln(&promotionString)
							if err != nil {
								fmt.Println("Error reading input: ", err)
								continue
							}
							promotionMap := map[string]int{"queen": QueenPromotion, "rook": RookPromotion, "bishop": BishopPromotion, "knight": KnightPromotion}
							promotion, ok := promotionMap[strings.ToLower(promotionString)]
							if !ok {
								fmt.Println("Invalid promotion")
							} else {
								foundMove = BuildMove(foundMove.OriginSquare(), foundMove.DestinationSquare(), uint16(promotion), PromotionSpecialMove)
								break
							}
						}
					}
					gameState.MakeMove(foundMove)
					break
				} else {
					fmt.Println("Invalid Move")
				}
			}
		} else {
			searchTime := GetUserFloat("How long would you like to search (in seconds)?: ")
			bestMove := gameState.IterativeDeepiningSearch(time.Duration(searchTime*float64(time.Second)), true)
			gameState.MakeMove(bestMove)
		}
		moves := gameState.quickGenMoves()
		if len(*moves) == 0 {
			fmt.Println(gameState)
			if gameState.check {
				if playerTurn {
					fmt.Println("You Win")
				} else {
					fmt.Println("Ghobos Wins")
				}
			} else {
				fmt.Println("Stalemate")
			}
			gameOver = true
		} else if gameState.lastCapOrPawn >= 100 {
			fmt.Println(gameState)
			fmt.Println("Draw by 50 move rule")
			gameOver = true
		} else if gameState.repetitionMap.get(gameState.hashcode) >= 3 {
			fmt.Println(gameState)
			fmt.Println("Draw by 3 fold repetition")
			gameOver = true
		}
		playerTurn = !playerTurn
	}
}
