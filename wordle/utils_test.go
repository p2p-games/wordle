package wordle

import (
	"fmt"
	"testing"

	"github.com/p2p-games/wordle/model"
)

func TestColorWords(t *testing.T) {
	// generate a comparable word
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

	fmt.Println(ComposeWordleVisualWord("eaaaa", target) + "\n")
	fmt.Println(ComposeWordleVisualWord("hello", target) + "\n")
	fmt.Println(ComposeWordleVisualWord("eeeee", target) + "\n")
	fmt.Println(ComposeWordleVisualWord("hella", target) + "\n")
}
