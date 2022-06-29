package wordle

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TerminalManager is a Text User Interface (TUI) for a ChatRoom.
// The Run method will draw the UI to the terminal in "fullscreen"
// mode. You can quit with Ctrl-C, or by typing "/quit" into the
// chat prompt.
type TerminalManager struct {
	Game     *WordGame
	app      *tview.Application
	debugBox *tview.TextView

	stateBox     io.Writer
	inputCh      chan string
}

// NewTerminalManager returns a new TerminalManager struct that controls the text UI.
// It won't actually do anything until you call Run().
func NewTerminalManager(game *WordGame) *TerminalManager {
	app := tview.NewApplication()

	// make a text view to contain our chat messages
	stateBox := tview.NewTextView()
	stateBox.SetDynamicColors(true)
	stateBox.SetBorder(true)
	stateBox.SetTitle("P2P Wordle")

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
			close(inputCh)
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
		app:          app,
		Game:         game,
		stateBox:     stateBox,
		debugBox:     debugB,
		inputCh:      inputCh,
	}
}

func (ui *TerminalManager) Run(ctx context.Context, initialized *bool) error {
	go ui.handleEvents(ctx)

	go func() {
		// mark as initialized after ~200 milliseconds
		time.Sleep(200 * time.Millisecond)
		*initialized = true
	}()
	return ui.app.Run()
}

func (ui *TerminalManager) RefreshWordleState() {
	stateB := ui.stateBox.(*tview.TextView)
	stateB.Clear()
	s := ui.Game.ComposeStateUI()
	fmt.Fprintf(ui.stateBox, "%s\n", s)
}

func (ui *TerminalManager) AddDebugMsg(s string) {
	fmt.Fprintf(ui.debugBox, "%s\n", s)
}

func (ui *TerminalManager) handleEvents(ctx context.Context) {
	defer ui.app.Stop()
	for {
		ui.RefreshWordleState()
		select {
		case input, ok := <-ui.inputCh:
			if !ok {
				return
			}

			switch ui.Game.StateIdx {
			case 0:
				ui.AddDebugMsg(fmt.Sprintf("Your next proposed word: %s", input))
			case 1:
				ui.AddDebugMsg(fmt.Sprintf("Last guess: %s (freezes are expected if no peers connected)", input))
			default:
				continue
			}
			// when the user types in a line, publish it to the chat room and print to the message window
			err := ui.Game.NewStdinInput(input)
			if err != nil {
				ui.AddDebugMsg(fmt.Sprintf("publish error: %s", err))
			}
		case <-ctx.Done():
			return
		}
	}
}
