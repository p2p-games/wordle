package wordle

import (
	"context"
	"strings"
	"testing"

	"github.com/p2p-games/wordle/model"
	"github.com/stretchr/testify/require"
)

func TestStateTransition(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	require := require.New(t)
	salts := []string{"a", "b", "c", "d", "e"}
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

	guessMsgC := make(chan guess)
	go func() {
		for {
			select {
			case _ = <-guessMsgC:
			case <-ctx.Done():
				return
			}
		}
	}()
	chars, err := model.GetChars("hello", salts)
	word := &model.Word{Chars: chars}
	require.NoError(err)

	wordGame := NewWordGame(word, guessMsgC)

	t.Log(wordGame.ComposeStateUI())

	require.Equal(wordGame.StateIdx, 0)
	require.Equal(*wordGame.Target, *target)

	// add the next add new input
	err = wordGame.NewStdinInput("nextt")
	require.NoError(err)
	require.Equal(wordGame.StateIdx, 1)
	require.Equal(wordGame.NextWord, "nextt")

	// add the next add new input
	err = wordGame.NewStdinInput("/ipfs")
	require.NoError(err)
	require.Equal(wordGame.StateIdx, 1)
	require.Equal(0, len(wordGame.AttemptedWords))

	for i, word := range []string{"Guess", "ramon", "pedro", "lucas"} {
		// add the next add new input
		err = wordGame.NewStdinInput(word)
		require.NoError(err)
		require.Equal(wordGame.StateIdx, 1)
		require.Equal(wordGame.AttemptedWords[i], strings.ToLower(word))
		t.Log(wordGame.ComposeStateUI())
		//time.Sleep(10 * time.Second)
		//ClearTerminal()
	}

	// add the next add new input
	err = wordGame.NewStdinInput("hello")
	require.NoError(err)
	require.Equal(wordGame.StateIdx, 2)
	require.Equal(wordGame.AttemptedWords[4], "hello")

	t.Log(wordGame.ComposeStateUI())
	//time.Sleep(10 * time.Second)
	//ClearTerminal()

	// add the next add new input
	err = wordGame.NewStdinInput("juanx")
	require.Error(err)
	require.Equal(wordGame.StateIdx, 2)
	require.Equal(len(wordGame.AttemptedWords), 5)

	t.Log(wordGame.ComposeStateUI())
	//time.Sleep(10 * time.Second)
	//ClearTerminal()

	// try to see if it was success
	guessed := wordGame.WasGuessed()
	require.Equal(guessed, true)

	close(guessMsgC)
	cancel()
}
