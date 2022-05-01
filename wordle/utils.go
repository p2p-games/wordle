package wordle

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"unicode"

	"github.com/p2p-games/wordle/model"
)

func ClearTerminal() {
	cmd := exec.Command("clear") //Linux example, its tested
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func IsLetter(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func IsGuessSuccess(result []bool) bool {
	for _, b := range result {
		if !b {
			return false
		}
	}
	return true
}

func GetSaltsFromWord(word *model.Word) []string {
	salts := make([]string, 0)
	for _, i := range word.Chars {
		salts = append(salts, i.Salt)
	}
	return salts
}

// Color utilities for the wordle

//var Reset = "\033[0m"
//var Red = "\033[31m"
var Green = "green"   // "\033[32m"
var Yellow = "yellow" // "\033[33m"

func ComposeWordleVisualWord(word string, target *model.Word) string {
	// get the salts for the word
	salts := GetSaltsFromWord(target)
	var c []string
	for _, l := range word {
		c = append(c, string(l))
	}
	// compose the word
	//chars, _ := model.GetChars(word, salts)

	compWord := ""
	h := sha256.New()
	for i, char := range c {
		charStatus := 0
	out:
		for _, s := range salts {
			// check all the hashes of the letter with the different salts
			h.Reset()
			h.Write([]byte(char))
			h.Write([]byte(s))
			hashP := fmt.Sprintf("%x", h.Sum(nil))[:45]

			for j, charB := range target.Chars {
				if hashP == charB.Hash {
					charStatus = 1
					if i == j {
						charStatus = 2
						break out
					}
				}
			}
		}
		// once we know wheather it is or not, select color

		switch charStatus {
		case 0: // not in the word
			compWord += composeCharWithColor(c[i], "")
		case 1: // in the word but on wrong possition
			compWord += composeCharWithColor(c[i], Yellow)
		case 2: // bingo
			compWord += composeCharWithColor(c[i], Green)
		}

	}
	return compWord
}

// compose the character over the color and reset the terminal color
func composeCharWithColor(char string, color string) string {
	return fmt.Sprintf("[%s]%s", color, char)
}
