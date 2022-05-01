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
		OthersGuessC: make(chan struct{}, 10),
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
	ui.displayStateString(ui.Game.ComposeStateUI())
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

func (ui *TerminalManager) AddDebugItem(s string) {
	ui.displayDebugString(s)
}

func (ui *TerminalManager) handleEvents() {
	peerRefreshTicker := time.NewTicker(time.Second)
	defer peerRefreshTicker.Stop()

	for {
		ui.displayStateStatus()
		select {
		case input := <-ui.inputCh:
			switch ui.Game.StateIdx {
			case 0:
				ui.AddDebugItem(fmt.Sprintf("Your next proposed word: %s", input))
			case 1:
				ui.AddDebugItem(fmt.Sprintf("Last guess: %s (freezes are expected if no peers connected)", input))
			default:
				continue
			}
			// when the user types in a line, publish it to the chat room and print to the message window
			err := ui.Game.NewStdinInput(input)
			if err != nil {
				ui.AddDebugItem(fmt.Sprintf("publish error: %s", err))
			}

		case others := <-ui.OthersGuessC:
			ui.AddDebugItem(fmt.Sprintln("new gueess from someone", others))
			// when we receive a message from the chat room, print it to the message window

		case <-ui.ctx.Done():
			fmt.Println("context done")
			return

		case <-ui.doneCh:
			fmt.Println("channel done")

			return
		}
	}
}
