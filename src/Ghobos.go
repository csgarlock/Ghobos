package main

import (
	"fmt"
	"strings"
	"time"
)

func main() {
	InitializeMoveBoards()
	InitializeEvalVariables()
	SetupTable(1024)
	PerftTester()
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
	gameState := StartingFen()
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
				fmt.Println(playerMove)
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
					fmt.Println(foundMove)
					gameState.MakeMove(foundMove)
					break
				} else {
					fmt.Println("Invalid Move")
				}
			}
		} else {
			searchTime := GetUserFloat("How long would you like to search (in seconds)?: ")
			bestMove := gameState.IterativeDeepiningSearch(time.Duration(searchTime * float64(time.Second)))
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
