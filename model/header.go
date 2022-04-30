package model

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/multiformats/go-multihash"
)

type Header struct {
	Height         int
	LastHeaderHash multihash.Multihash

	Guess    *Word
	Proposal *Word

	PeerID string
}

func (h *Header) Hash() (multihash.Multihash, error) {
	data, err := json.Marshal(h)
	if err != nil {
		return nil, err
	}

	hash := sha256.New()
	hash.Write(data)
	mhash, err := multihash.Encode(hash.Sum(nil), multihash.SHA2_256)
	if err != nil {
		return nil, err
	}
	return mhash, nil
}

type Word struct {
	Chars []*Char
}

type Char struct {
	Salt string
	Hash string
}

func NewHeader(guess string, gSalts []string, proposal, peerID string, target multihash.Multihash) (*Header, error) {
	pSalt := make([]string, 0, len(proposal))
	for i := 0; i < len(proposal); i++ {
		pSalt = append(pSalt, RandomString(30))
	}
	pw, err := GetChars(proposal, pSalt)
	if err != nil {
		return nil, err
	}

	gw, err := GetChars(guess, gSalts)
	if err != nil {
		return nil, err
	}

	return &Header{
		LastHeaderHash: target,
		PeerID:         peerID,
		Guess: &Word{
			Chars: gw,
		},
		Proposal: &Word{
			Chars: pw,
		},
	}, nil
}

func VerifyString(guess string, challenge *Word) ([]bool, error) {
	result := make([]bool, len(challenge.Chars))
	if len(guess) != len(challenge.Chars) {
		return result, nil
	}

	salts := make([]string, 0, len(challenge.Chars))
	for _, ch := range challenge.Chars {
		salts = append(salts, ch.Salt)
	}

	gch, err := GetChars(guess, salts)
	if err != nil {
		return result, err
	}

	for i, ch := range challenge.Chars {
		result[i] = gch[i].Hash == ch.Hash
	}

	return result, nil
}

func Verify(guess, challenge *Word) bool {
	for i, ch := range challenge.Chars {
		if guess.Chars[i].Hash != ch.Hash {
			return false
		}
	}
	return true
}

var ErrSaltsAndCharsDidntMatch = errors.New("number of salts and number of letters didn't match")

func GetChars(word string, salts []string) ([]*Char, error) {
	if len(word) != len(salts) {
		return nil, ErrSaltsAndCharsDidntMatch
	}
	chars := make([]*Char, len(word))
	h := sha256.New()
	for i, r := range word {
		salt := salts[i]
		h.Reset()
		h.Write([]byte{byte(r)})
		h.Write([]byte(salt))

		hash := fmt.Sprintf("%x", h.Sum(nil))[:45]

		ch := &Char{
			Salt: salt,
			Hash: hash,
		}

		chars = append(chars, ch)
	}

	return chars, nil
}
