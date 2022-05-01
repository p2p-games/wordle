package wordle

import (
	"bufio"
	"context"
	"fmt"
	"os"
)

type TerminalManager struct {
	ctx context.Context
}

func NewTerminalManager(ctx context.Context) *TerminalManager {
	return &TerminalManager{ctx}
}

func (t *TerminalManager) RefreshAndRead(game *WordGame) {
	//ClearTerminal()
	go t.launchStdinReader(game)
	fmt.Println(game.ComposeStateUI())
}

func (t *TerminalManager) Refresh(game *WordGame) {
	//ClearTerminal()
	fmt.Println(game.ComposeStateUI())
}

func (t *TerminalManager) launchStdinReader(game *WordGame) {
	reader := bufio.NewReader(os.Stdin)

	input, _ := reader.ReadString('\n')
	fmt.Println(input)
	if len(input) > 15 {
		t.RefreshAndRead(game)
		return
	}
	err := game.NewStdinInput(input)
	if err != nil {
		fmt.Println("error with new input", err.Error())
		t.RefreshAndRead(game)
	}
}
