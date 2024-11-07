package main

import (
	"fmt"
	"strconv"
)

func clampInt32(x int32, min int32, max int32) int32 {
	if x > max {
		return max
	} else if x < min {
		return min
	} else {
		return x
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
