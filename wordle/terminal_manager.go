package wordle

import (
	"context"
	"fmt"
	"io"
	"time"

	tcell "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TerminalManager is a Text User Interface (TUI) for a ChatRoom.
// The Run method will draw the UI to the terminal in "fullscreen"
// mode. You can quit with Ctrl-C, or by typing "/quit" into the
// chat prompt.
type TerminalManager struct {
	ctx      context.Context
	Game     *WordGame
	app      *tview.Application
	debugBox *tview.TextView

	stateBox     io.Writer
	inputCh      chan string
	DebugCh      chan string
	OthersGuessC chan struct{}

	doneCh chan struct{}
}

// NewTerminalManager returns a new TerminalManager struct that controls the text UI.
// It won't actually do anything until you call Run().
func NewTerminalManager(ctx context.Context, game *WordGame) *TerminalManager {
	app := tview.NewApplication()

	// make a text view to contain our chat messages
	stateBox := tview.NewTextView()
	stateBox.SetDynamicColors(true)
	stateBox.SetBorder(true)
	stateBox.SetTitle(fmt.Sprintf("P2P Wordle"))

	// text views are io.Writers, but they don't automatically refresh.
	// this sets a change handler to force the app to redraw when we get
	// new messages to display.
	stateBox.SetChangedFunc(func() {
		app.Draw()
	})

	debugB := tview.NewTextView()
	debugB.SetBorder(true)
	debugB.SetTitle("Debug Events")
	debugB.SetChangedFunc(func() { app.Draw() })

	// an input field for typing messages into
	inputCh := make(chan string, 32)
	input := tview.NewInputField().
		SetLabel(" > ").
		SetFieldWidth(0).
		SetFieldBackgroundColor(tcell.ColorBlack)

	// the done func is called when the user hits enter, or tabs out of the field
	input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			// we don't want to do anything if they just tabbed away
			return
		}
		line := input.GetText()
		if len(line) == 0 {
			// ignore blank lines
			return
		}

		// bail if requested
		if line == "/quit" {
			app.Stop()
			return
		}

		// send the line onto the input chan and reset the field text
		inputCh <- line
		input.SetText("")
	})

	// make a text view to hold the list of peers in the room, updated by ui.refreshPeers()
	/*
		peersList := tview.NewTextView()
		peersList.SetBorder(true)
		peersList.SetTitle("Peers")
		peersList.SetChangedFunc(func() { app.Draw() })
	*/

	// chatPanel is a horizontal box with messages on the left and peers on the right
	// the peers list takes 20 columns, and the messages take the remaining space
	wordlePannel := tview.NewFlex().
		AddItem(stateBox, 0, 1, false).
		AddItem(debugB, 60, 1, false)

	// flex is a vertical box with the chatPanel on top and the input field at the bottom.

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(wordlePannel, 0, 1, false).
		AddItem(input, 1, 1, true)

	app.SetRoot(flex, true)

	return &TerminalManager{
		ctx:          ctx,
		app:          app,
		Game:         game,
		stateBox:     stateBox,
		debugBox:     debugB,
		inputCh:      inputCh,
		OthersGuessC: make(chan struct{}),
		DebugCh:      make(chan string),
		doneCh:       make(chan struct{}, 1),
	}
}

func (ui *TerminalManager) Run() error {
	go ui.handleEvents()
	defer ui.end()

	err := ui.app.Run()
	if err != nil {
		return err
	}
	ui.displayStateStatus()
	return nil

}

// end signals the event loop to exit gracefully
func (ui *TerminalManager) end() {
	ui.doneCh <- struct{}{}
}

func (ui *TerminalManager) displayStateString(s string) {
	fmt.Fprintf(ui.stateBox, "%s\n", s)
}

func (ui *TerminalManager) displayStateStatus() {
	s := ui.Game.ComposeStateUI()
	ui.displayStateString(s)
}

func (ui *TerminalManager) displayDebugString(s string) {
	stateB := ui.stateBox.(*tview.TextView)
	stateB.Clear()
	fmt.Fprintf(ui.debugBox, "%s\n", s)
}

func (ui *TerminalManager) addDebugItem(s string) {
	ui.displayDebugString(s)
}

func (ui *TerminalManager) handleEvents() {
	peerRefreshTicker := time.NewTicker(time.Second)
	defer peerRefreshTicker.Stop()

	for {
		select {
		case input := <-ui.inputCh:
			ui.AddDebugEvent(fmt.Sprintf("New input: %s", input))
			// when the user types in a line, publish it to the chat room and print to the message window
			err := ui.Game.NewStdinInput(input)
			if err != nil {
				ui.AddDebugEvent(fmt.Sprintln("publish error: %s", err))
			}
			ui.AddDebugEvent("input call done:")

		case _ = <-ui.OthersGuessC:
			// when we receive a message from the chat room, print it to the message window

		case debugS := <-ui.DebugCh:
			ui.addDebugItem(debugS)

		case <-ui.ctx.Done():
			return

		case <-ui.doneCh:
			return
		}
		ui.displayStateStatus()
	}
}

func (ui *TerminalManager) AddDebugEvent(s string) {
	go func() {
		ui.DebugCh <- s
	}()
}
