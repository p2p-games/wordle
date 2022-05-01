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

func (t *TerminalManager) Refresh(game *WordGame) {
	ClearTerminal()
	go t.launchStdinReader(game)
	fmt.Println(game.ComposeStateUI())
}

func (t *TerminalManager) launchStdinReader(game *WordGame) {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	game.NewStdinInput(input)
}

/*
func (t *TerminalManager) () {

}
*/
