package wordle

import (
	"context"
	"fmt"
	"time"

	"github.com/p2p-games/wordle/model"
)

const (
	MinWordLen int = 3
	MaxWordLen int = 25
	MaxApptemps int = 5
)

type WordleUI struct {

	WordleServ  *Service
	CurrentGame *WordGame

	CannonicalHeader *model.Header

	uiInitialized bool
	tm            *TerminalManager
}

func NewWordleUI(wordleServ *Service) *WordleUI {
	ui := &WordleUI{
		WordleServ: wordleServ,
	}

	wordleServ.SetLog(func(s string) {
		// wait untill UI is initialilzed
		for !ui.uiInitialized {
			time.Sleep(200 * time.Millisecond)
		}
		ui.AddDebugItem(s)
	})

	return ui
}

func (w *WordleUI) Run(ctx context.Context) (err error) {
	// get the latest header from the server
	w.CannonicalHeader, err = w.WordleServ.Head(ctx)
	if err != nil {
		return fmt.Errorf("unable to load latest head: %s", err)
	}

	// generate a guess channel
	userGuessC := make(chan guess)

	// generate a new game
	w.CurrentGame = NewWordGame(w.WordleServ.ID(), w.CannonicalHeader.PeerID, w.CannonicalHeader.Proposal, userGuessC)

	// get the channel for incoming headers
	incomingHeaders, err := w.WordleServ.Guesses(ctx)
	if err != nil {
		return fmt.Errorf("unable to get guesses channel: %s", err)
	}
	go func() {
		for {
			select {
			case userGuess := <-userGuessC:
				w.tm.AddDebugMsg("about to send new guess")
				// notify the service of a new User Guess
				err = w.WordleServ.Guess(ctx, userGuess.Guess, userGuess.Proposal)
				if err != nil {
					w.tm.AddDebugMsg("error sending guess" + userGuess.Guess + " - " + err.Error())
				}

				// check if the guess if the guess was right for debug-msg purpose
				bools, err := model.VerifyString(userGuess.Guess, w.CannonicalHeader.Proposal)
				if err != nil {
					w.tm.AddDebugMsg("error verifying the guess" + err.Error())
				}

				if IsGuessSuccess(bools) {
					w.tm.AddDebugMsg("succ guess sent")
				} else {
					w.tm.AddDebugMsg("wrong guess sent")
				}

			case recHeader := <-incomingHeaders: // incoming New Message from surrounding peers
				//w.AddDebugItem(fmt.Sprintf("guess received from %s", recHeader.PeerID))
				// verify weather the header is correct or not
				if model.Verify(recHeader.Guess, w.CannonicalHeader.Proposal) {
					w.CannonicalHeader = recHeader
					// generate a new one game

					w.CurrentGame = NewWordGame(w.WordleServ.ID(), w.CannonicalHeader.PeerID, recHeader.Proposal, userGuessC)
					// refresh the terminal manager
					w.tm.Game = w.CurrentGame
				}
			case <-ctx.Done(): // context shutdown
				return
			}

			// render the new state of the UI
			w.tm.RefreshWordleState()
		}
	}()

	// generate a terminal manager
	w.tm = NewTerminalManager(w.CurrentGame)

	return w.tm.Run(ctx, &w.uiInitialized)
}

func (w *WordleUI) AddDebugItem(s string) {
	w.tm.AddDebugMsg(s)
}
