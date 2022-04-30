package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewHeader(t *testing.T) {
	require := require.New(t)

	h, err := NewHeader("hello", []string{"a", "b", "c", "d", "e"}, "proposal", "peerID", nil)
	require.NoError(err)

	require.Equal(&Word{
		Chars: []*Char{
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
	}, h.Guess)
}

func TestVerify(t *testing.T) {
	require := require.New(t)

	ch, err := GetChars("table", []string{"a", "b", "c", "d", "e"})
	require.NoError(err)

	v, err := VerifyString("apple", &Word{Chars: ch})
	require.NoError(err)

	require.Equal([]bool{false, false, false, true, true}, v)
}

func TestVerifyDifferentLengths(t *testing.T) {
	require := require.New(t)

	ch, err := GetChars("table", []string{"a", "b", "c", "d", "e"})
	require.NoError(err)

	v, err := VerifyString("shortable", &Word{Chars: ch})
	require.NoError(err)

	require.Equal([]bool{false, false, false, false, false}, v)

	ch, err = GetChars("shortable", []string{"a", "b", "c", "d", "e", "f", "g", "h", "j"})
	require.NoError(err)

	v, err = VerifyString("short", &Word{Chars: ch})
	require.NoError(err)

	require.Equal([]bool{false, false, false, false, false, false, false, false, false}, v)
}
