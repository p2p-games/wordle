package wordle

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"strconv"

	"github.com/ipfs/go-datastore"

	"github.com/p2p-games/wordle/model"
)

// use topic name as genesis
var genesis, _ = model.NewHeader(&model.Header{Proposal: &model.Word{}}, "", topic, "")

type Store struct {
	ds datastore.Batching
}

func NewStore(ds datastore.Batching) *Store {
	return &Store{
		ds: ds,
	}
}

func (s *Store) Init(ctx context.Context) error {
	return s.Append(ctx, genesis)
}

func (s *Store) Head(ctx context.Context) (*model.Header, error) {
	data, err := s.ds.Get(ctx, headKey)
	switch err {
	default:
		return nil, err
	case datastore.ErrNotFound:
		err := s.Init(ctx)
		if err != nil {
			return nil, err
		}

		return genesis, nil
	case nil:
		headHeight, _ := binary.Uvarint(data)
		if err != nil {
			return nil, err
		}

		return s.Get(ctx, int(headHeight))
	}
}

func (s *Store) Append(ctx context.Context, h *model.Header) error {
	data, err := json.Marshal(h)
	if err != nil {
		return err
	}

	err = s.ds.Put(ctx, datastore.NewKey(strconv.Itoa(h.Height)), data)
	if err != nil {
		return err
	}

	data = make([]byte, 8)
	n := binary.PutUvarint(data, uint64(h.Height))
	err = s.ds.Put(ctx, headKey, data[:n])
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) Get(ctx context.Context, height int) (*model.Header, error) {
	data, err := s.ds.Get(ctx, datastore.NewKey(strconv.Itoa(height)))
	if err != nil {
		return nil, err
	}

	h := &model.Header{}
	return h, json.Unmarshal(data, &h)
}

var headKey = datastore.NewKey("head")
