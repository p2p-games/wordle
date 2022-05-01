package wordle

import (
	"context"
	"fmt"

	"github.com/p2p-games/wordle/model"
)

var MinWordLen int = 3
var MaxWordLen int = 25
var MaxApptemps int = 5
var testWord string = "test"

type WordleUI struct {
	ctx context.Context

	PeerId string

	WordleServ  *Service
	CurrentGame *WordGame

	CannonicalHeader *model.Header

	tm *TerminalManager
}

func NewWordleUI(ctx context.Context, wordleServ *Service, peerId string) *WordleUI {
	return &WordleUI{
		ctx:        ctx,
		PeerId:     peerId,
		WordleServ: wordleServ,
	}
}

func (w *WordleUI) Run() {
	var err error
	// get the latest header from the server
	w.CannonicalHeader, err = w.WordleServ.Head(w.ctx)

	if err != nil {
		panic("non able to load any header from the datastore, not even genesis??!")
	}

	// generate a new game
	guessMsgC := make(chan guess)
	w.CurrentGame = NewWordGame(w.PeerId, w.CannonicalHeader.PeerID, w.CannonicalHeader.Proposal, guessMsgC)

	// generate a terminal manager
	w.tm = NewTerminalManager(w.ctx, w.CurrentGame)
	err = w.tm.Run()
	if err != nil {
		panic(err)
	}
	// get the channel for incoming headers
	incomingHeaders, err := w.WordleServ.Guesses(w.ctx)
	if err != nil {
		panic("unable to retrieve the channel of headers from the user interface")
	}

	for {
		select {
		case guess := <-guessMsgC: // new guess from the user
			w.AddDebugItem(fmt.Sprintf("sending guess %s to peers", guess.Guess))
			w.WordleServ.Guess(w.ctx, guess.Guess, guess.Proposal)
			//w.tm.RefreshAndRead(w.CurrentGame)

		case recHeader := <-incomingHeaders: // incoming New Message from surrounding peers
			w.AddDebugItem(fmt.Sprintf("guess received from %s", recHeader.PeerID))
			// verify weather the header is correct or not
			if model.Verify(recHeader.Guess, w.CannonicalHeader.Proposal) {
				w.CannonicalHeader = recHeader
				// generate a new one game
				w.CurrentGame = NewWordGame(w.PeerId, w.CannonicalHeader.PeerID, recHeader.Proposal, guessMsgC)

				// refresh the terminal manager
				w.tm.Game = w.CurrentGame
			} else {
				// Actually, there isn't anything else to do
				continue
			}
		case <-w.ctx.Done(): // context shutdown
			close(guessMsgC)
			return
		}
	}
}

func (w *WordleUI) AddDebugItem(s string) {
	w.tm.AddDebugItem(s)
}
