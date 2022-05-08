package wordle

import (
	"context"
	"time"

	"github.com/p2p-games/wordle/model"
)

var MinWordLen int = 3
var MaxWordLen int = 25
var MaxApptemps int = 5

type WordleUI struct {
	ctx context.Context

	PeerId string

	WordleServ  *Service
	CurrentGame *WordGame

	CannonicalHeader *model.Header

	uiInitialized bool
	tm            *TerminalManager
}

func NewWordleUI(ctx context.Context, wordleServ *Service, peerId string) *WordleUI {

	ui := &WordleUI{
		ctx:        ctx,
		PeerId:     peerId,
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

func (w *WordleUI) Run() {
	var err error
	// get the latest header from the server
	w.CannonicalHeader, err = w.WordleServ.Head(w.ctx)

	if err != nil {
		panic("non able to load any header from the datastore, not even genesis??!")
	}

	// generate a guess channel
	userGuessC := make(chan guess)

	// generate a new game
	w.CurrentGame = NewWordGame(w.PeerId, w.CannonicalHeader.PeerID, w.CannonicalHeader.Proposal, userGuessC)

	// get the channel for incoming headers
	incomingHeaders, err := w.WordleServ.Guesses(w.ctx)
	if err != nil {
		panic("unable to retrieve the channel of headers from the user interface")
	}
	go func() {
		for {
			select {
			case userGuess := <-userGuessC:
				w.tm.AddDebugMsg("about to send new guess")
				// notify the service of a new User Guess
				err = w.WordleServ.Guess(w.ctx, userGuess.Guess, userGuess.Proposal)
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

					w.CurrentGame = NewWordGame(w.PeerId, w.CannonicalHeader.PeerID, recHeader.Proposal, userGuessC)
					// refresh the terminal manager
					w.tm.Game = w.CurrentGame
				}
			case <-w.ctx.Done(): // context shutdow
				return
			}

			// render the new state of the UI
			w.tm.RefreshWordleState()
		}
	}()

	// generate a terminal manager
	w.tm = NewTerminalManager(w.ctx, w.CurrentGame)
	err = w.tm.Run(&w.uiInitialized)
	if err != nil {
		panic(err)
	}

}

func (w *WordleUI) AddDebugItem(s string) {
	w.tm.AddDebugMsg(s)
}
