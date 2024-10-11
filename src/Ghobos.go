package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func main() {
	InitializeMoveBoards()
	InitializeEvalVariables()
	SetupTable(4096)
	UIGame()
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
	playerTurn := true
	if playerSide == Black {
		playerTurn = false
	}
	for !gameOver {
		fmt.Println(gameState)
		if playerTurn {
			for {
				playerMove := getUserMove()
				fmt.Println(playerMove)
				validMoves := gameState.genAllMoves(true)
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
							promotion, ok := promotionMap[promotionString]
							if !ok {
								fmt.Println("Invalid promotion")
							} else {
								foundMove = foundMove | (1 << promotion)
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
			bestMove := gameState.IterativeDeepiningSearch(time.Duration(searchTime * float64(time.Second)))
			gameState.MakeMove(bestMove)
		}
		playerTurn = !playerTurn
	}
}

func PerftChecker(depth int64, s *State) {
	var d int64 = 0
	currentDepth := depth
	for {
		moves := s.genAllMoves(true)
		for i, move := range *moves {
			fmt.Print(move.ShortString())
			fmt.Printf(", Move %d: ", i)
			s.MakeMove(move)
			var counter int64 = 0
			Perft(currentDepth-1, &counter, s, &d, &d, &d, &d, (*time.Duration)(&d))
			fmt.Println(counter)
			s.UnMakeMove(move)
		}
		move_selection := GetUserNumber("Enter move number: ")
		s.MakeMove((*moves)[move_selection])
		currentDepth--
	}
}

// Boards Complete to Depth 5: 1, 2, 3, 5, 6
func PerftRunner(depth int64, s *State) {
	var perftCounter int64 = 0
	var checkCounter int64 = 0
	var mateCounter int64 = 0
	var enPassantCounter int64 = 0
	var castleCounter int64 = 0
	var timer time.Duration = 0
	start := time.Now()
	Perft(depth, &perftCounter, s, &checkCounter, &mateCounter, &enPassantCounter, &castleCounter, &timer)
	fmt.Printf("Highest Depth Node Count: %d\n", perftCounter)
	fmt.Printf("Check Counter: %d\n", checkCounter)
	fmt.Printf("Mate Counter: %d\n", mateCounter)
	fmt.Printf("En Passant Counter: %d\n", enPassantCounter)
	fmt.Printf("Castle Counter: %d\n", castleCounter)
	fmt.Println(time.Since(start))
	fmt.Println(timer)
}

func Perft(depth int64, moveCounter *int64, s *State, checkCounter *int64, mateCounter *int64, enPassantCounter *int64, castleCounter *int64, sectionTimer *time.Duration) {
	if depth == 0 {
		*moveCounter++
		if s.check {
			//start := time.Now()
			moves := s.genAllMoves(true)
			//*sectionTimer += time.Since(start)
			if len(*moves) == 0 {
				*mateCounter++
			}
		}
	} else {
		start := time.Now()
		moves := s.genAllMoves(true)
		*sectionTimer += time.Since(start)
		if len(*moves) == 0 {
			*mateCounter++
		}
		for _, move := range *moves {
			// fmt.Println(s)
			// fmt.Println("Trying to Make " + move.String())
			s.MakeMove(move)
			Perft(depth-1, moveCounter, s, checkCounter, mateCounter, enPassantCounter, castleCounter, sectionTimer)
			// fmt.Println(s)
			// fmt.Println("Trying to Unmake " + move.String())
			s.UnMakeMove(move)
		}
	}
}

func getUserMove() Move {
	fileMap := map[rune]int{'a': 0, 'b': 1, 'c': 2, 'd': 3, 'e': 4, 'f': 5, 'g': 6, 'h': 7}
	for {
		var moveInput string
		fmt.Print("Please enter your move enter in the format source square destination square (eg e2e4): ")
		_, err := fmt.Scanln(&moveInput)
		if err != nil {
			fmt.Println("Error reading unput: ", err)
			continue
		}
		if len(moveInput) != 4 {
			fmt.Println("Invalid Move (Bad Format)")
			continue
		}
		sourceFileRune := moveInput[0]
		sourceRankRune := moveInput[1]
		desFileRune := moveInput[2]
		desRankRune := moveInput[3]
		sourceFile, sourceFileOk := fileMap[rune(sourceFileRune)]
		sourceRank, sourceRankErr := strconv.Atoi(string(sourceRankRune))
		sourceRank -= 1
		desFile, desFileOk := fileMap[rune(desFileRune)]
		desRank, desRankErr := strconv.Atoi(string(desRankRune))
		desRank -= 1
		if !sourceFileOk || sourceRankErr != nil || !desFileOk || desRankErr != nil {
			fmt.Println("Invalid Move (Bad Format)")
			continue
		}
		playerMove := BuildMove(sFromRankFile(sourceFile, sourceRank), sFromRankFile(desFile, desRank), 0, 0)
		return playerMove
	}
}

func GetUserNumber(prompt string) int {
	for {
		var userInput string
		var num int

		fmt.Print(prompt)
		_, err := fmt.Scanln(&userInput)
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}
		num, err = strconv.Atoi(userInput)
		if err != nil {
			fmt.Println("Error converting to integer:", err)
			continue
		}
		return num
	}
}
func GetUserFloat(prompt string) float64 {
	for {
		var userInput string
		var num float64

		fmt.Print(prompt)
		_, err := fmt.Scanln(&userInput)
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}
		num, err = strconv.ParseFloat(userInput, 64)
		if err != nil {
			fmt.Println("Error converting to integer:", err)
			continue
		}
		return num
	}
}
