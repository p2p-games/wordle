package wordle

import (
	"context"
	"fmt"

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

	tm *TerminalManager
}

func NewWordleUI(ctx context.Context, wordleServ *Service, peerId string) *WordleUI {

	ui := &WordleUI{
		ctx:        ctx,
		PeerId:     peerId,
		WordleServ: wordleServ,
	}

	wordleServ.SetLog(func(s string) {
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

	// generate a new game
	w.CurrentGame = NewWordGame(w.ctx, w.PeerId, w.CannonicalHeader.PeerID, w.CannonicalHeader.Proposal, w.WordleServ)

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
		case recHeader := <-incomingHeaders: // incoming New Message from surrounding peers
			w.AddDebugItem(fmt.Sprintf("guess received from %s", recHeader.PeerID))
			// verify weather the header is correct or not
			if model.Verify(recHeader.Guess, w.CannonicalHeader.Proposal) {
				w.CannonicalHeader = recHeader
				// generate a new one game
				w.CurrentGame = NewWordGame(w.ctx, w.PeerId, w.CannonicalHeader.PeerID, recHeader.Proposal, w.WordleServ)

				// refresh the terminal manager
				w.tm.Game = w.CurrentGame
			} else {
				// Actually, there isn't anything else to do
				continue
			}
		case <-w.ctx.Done(): // context shutdow
			return
		}
	}
}

func (w *WordleUI) AddDebugItem(s string) {
	w.tm.AddDebugItem(s)
}
