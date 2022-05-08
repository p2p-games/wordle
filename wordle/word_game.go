package wordle

import (
	"fmt"
	"strings"

	"github.com/p2p-games/wordle/model"
	"github.com/pkg/errors"
)

const maxAttempts int = 5

type WordGame struct {
	PeerId string

	// to verify if the guess is correct
	Target *model.Word
	Salts  []string

	StateIdx int32
	NextWord string

	AttemptedWords []string
	isCorrect      map[string][]bool

	userGuessC chan guess
}

type guess struct {
	Guess    string
	Proposal string
}

// generate new game session
func NewWordGame(peerId string, proposerId string, target *model.Word, guessC chan guess) *WordGame {
	salts := GetSaltsFromWord(target)
	wg := &WordGame{
		PeerId:         peerId,
		Target:         target,
		Salts:          salts,
		StateIdx:       int32(0), // start requesting the word
		AttemptedWords: make([]string, 0),
		isCorrect:      make(map[string][]bool),
		userGuessC:     guessC,
	}
	if proposerId == peerId {
		// go straight to the 2 state (I already won)
		wg.StateIdx = 2
	}
	return wg
}

func (w *WordGame) ComposeStateUI() string {
	var s string
	switch w.StateIdx {
	case int32(0):
		s = "Introduce your word proposal as next word to guess:\n"
	case int32(1):
		s = "Guess which is the current Word:\n"
		for _, guessedWord := range w.AttemptedWords {
			if guessedWord != "" {
				// check wheather the word was correct or not
				correct := "[red]x[white]"
				comp, err := model.VerifyString(guessedWord, w.Target)
				if err != nil {
					continue
				}
				if IsGuessSuccess(comp) {
					correct = "[green]v[white]"
				}
				// compose the color strings with color chars

				s += fmt.Sprintf("\t[%s] %s\n", correct, ComposeWordleVisualWord(guessedWord, w.Target))
			}
		}
		s += fmt.Sprintf("\nAttempts left %d\n", maxAttempts-len(w.AttemptedWords))
	case int32(2):
		s = "\n\nCongrats, you guessed the word!\nWait untill someone guesses your word to play again\n"
		s += "list of guessed words:\n"
		for _, guessedWord := range w.AttemptedWords {
			if guessedWord != "" {
				// check wheather the word was correct or not
				correct := "[red]x[white]"
				comp, err := model.VerifyString(guessedWord, w.Target)
				if err != nil {
					continue
				}
				if IsGuessSuccess(comp) {
					correct = "[green]v[white]"
				}
				// compose the color strings with color chars

				s += fmt.Sprintf("\t[%s] %s\n", correct, ComposeWordleVisualWord(guessedWord, w.Target))
			}
		}
	case int32(3):
		s = "\n\tNo more attempts left for this word!\nWait untill someone guesses it to play again\n"
	default:
		s = "unrecognized state to generate the UI\n"
	}
	return s + "\ntype '/quit to exit game' \n"
}

func (w *WordGame) NewStdinInput(input string) error {
	// check if non alphanumeric character
	input = strings.ToLower(input)
	/*
		if !IsLetter(input) {
			return errors.New("not a letter")
		}
	*/
	// check in which state do we are
	switch w.StateIdx {
	case int32(0):
		err := w.addNextTarget(input, w.Salts)
		if err != nil {
			return err
		}
	case int32(1):
		err := w.addNewGuess(input)
		if err != nil {
			fmt.Println("error adding new guess")
			return err
		}
	default:
		// nothing
		return errors.New("unable to add any new guess")
	}
	return nil
}

func (w *WordGame) addNextTarget(nextWord string, salts []string) error {
	// check if we are in state 0
	if w.StateIdx != int32(0) {
		return errors.New("unable to add next target, not in state 0")
	}

	w.NextWord = nextWord
	// go to state 1
	w.StateIdx = 1
	return nil
}

func (w *WordGame) WasGuessed() bool {
	// check if we have already guessed 5 times or if guess correct
	for _, word := range w.AttemptedWords {
		comp, err := model.VerifyString(word, w.Target)
		if err != nil {
			continue
		}
		if IsGuessSuccess(comp) {
			return true
		}
	}
	return false
}

func (w *WordGame) addNewGuess(guessedWord string) error {
	// check if we are in state 1
	if w.StateIdx != int32(1) {
		return errors.New("unable to add next target, not in state 1")
	}

	// add the new word to the list of Attempted, Verify if its the correct word, add to the map the result
	w.AttemptedWords = append(w.AttemptedWords, guessedWord)

	correct, err := model.VerifyString(guessedWord, w.Target)
	if err != nil {
		return nil
	}

	w.isCorrect[guessedWord] = correct

	comp, err := model.VerifyString(guessedWord, w.Target)
	if err != nil {
		return err
	}

	if IsGuessSuccess(comp) {
		w.StateIdx = 2 // Congrats, wait untill someone guesses your word
	}

	// check if we did all the attempts
	if len(w.AttemptedWords) == maxAttempts && !IsGuessSuccess(comp) {
		w.StateIdx = 3 // Wait untill you can play again
	}

	g := guess{
		Guess:    guessedWord,
		Proposal: w.NextWord,
	}

	w.userGuessC <- g
	return nil
}
