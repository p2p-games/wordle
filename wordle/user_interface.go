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
		//panic("non able to load any header from the datastore, not even genesis??!")
		target := &model.Word{
			Chars: []*model.Char{
				{
					Salt: "a",
					Hash: "8693873cd8f8a2d9c7c596477180f851e525f4eaf55a4",
				},
				{
					Salt: "b",
					Hash: "fce2551fcc23040870d151006816cc39d3831abff948d",
				},
				{
					Salt: "c",
					Hash: "ef07b359570add31929a5422d400b16c7c84e35644cb2",
				},
				{
					Salt: "d",
					Hash: "e5a08ffd3d7509c66e79642edbdcd8ed889269a7164c7",
				},
				{
					Salt: "e",
					Hash: "3ab80789f271e9870121d65c28808d23c3ee2f4d1498e",
				},
			},
		}
		h := &model.Header{
			Proposal: target,
		}
		w.CannonicalHeader = h
	}

	// generate a new game
	guessMsgC := make(chan guess, 100)
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
			w.AddDebugItem(fmt.Sprintf("sending guess %#v to peers", guess))
			w.WordleServ.Guess(w.ctx, guess.Guess, guess.Proposal)
			w.AddDebugItem("guess sent")
			//w.tm.RefreshAndRead(w.CurrentGame)

		case recHeader := <-incomingHeaders: // incoming New Message from surrounding peers
			w.AddDebugItem(fmt.Sprintln("guess sent %#v", recHeader))
			// verify weather the header is correct or not
			/*
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
			*/
		case <-w.ctx.Done(): // context shutdown
			fmt.Println("contest closed! Ciao")
			close(guessMsgC)
			return
		}
	}
}

func (w *WordleUI) AddDebugItem(s string) {
	w.tm.AddDebugItem(s)
}
