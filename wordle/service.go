package wordle

import (
	"context"
	"encoding/json"

	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	core "github.com/libp2p/go-libp2p-core"
	net "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/p2p-games/wordle/model"
)

var log = logging.Logger("wordle")

var topic = "wordle"

var protoID protocol.ID = "/wordle/v0.0.1"

type Service struct {
	store *Store
	host core.Host
	pubsub *pubsub.PubSub
	topic *pubsub.Topic
}

func NewService(host core.Host, ds datastore.Batching, pubsub *pubsub.PubSub) *Service {
	return &Service{
		store: NewStore(ds),
		host:  host,
		pubsub: pubsub,
	}
}

func (s *Service) Start() error {
	err := s.pubsub.RegisterTopicValidator(topic, s.validate)
	if err != nil {
		return  err
	}

	s.topic, err = s.pubsub.Join(topic)
	if err != nil {
		return err
	}

	s.host.SetStreamHandler(protoID, s.handle)
	return err
}

func (s *Service) handle(stream net.Stream) {

}

func (s *Service) validate(ctx context.Context, _ peer.ID, msg *pubsub.Message) pubsub.ValidationResult {
	proposal := &model.Header{}
	err := json.Unmarshal(msg.Data, proposal)
	if err != nil {
		log.Errorw("unmarshalling proposal", "err", err)
		return pubsub.ValidationReject
	}

	head, err := s.store.Head(ctx)
	switch err {
	default:
		log.Errorw("getting local head", "err", err)
		return pubsub.ValidationIgnore
	case datastore.ErrNotFound:
	case nil:
	}

	if head == nil || model.Verify(head.Proposal, proposal.Guess) {
		err = s.store.Append(ctx, proposal)
		if err != nil {
			log.Errorw("appending the successful proposal", "err", err)
			return pubsub.ValidationIgnore
		}

		log.Debugf("rcvd successful guess")
	} else {
		log.Debugf("rcvd unsuccessful guess")
	}

	// TODO(@Wondertan): Add notification API so that users can know what guesses were rcvd, including self
	// we allow unsuccessful guesses to be passed around the network, but we store only successful ones
	return pubsub.ValidationAccept
}

func (s *Service) Guesses(ctx context.Context) (<-chan *model.Header, error) {
	sub, err := s.topic.Subscribe()
	if err != nil {
		return nil, err
	}

	out := make(chan *model.Header, 4)
	go func() {
		defer sub.Cancel()
		defer close(out)
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				return
			}

			header := &model.Header{}
			err = json.Unmarshal(msg.Data, header)
			if err != nil {
				log.Errorw("unmarshalling header", "err", err)
				continue
			}

			select {
			case out <- header:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, nil
}

func (s *Service) Guess(ctx context.Context, guess, proposal string) error {
	head, err := s.store.Head(ctx)
	switch err {
	default:
		return err
	case datastore.ErrNotFound:
		// a special case for genesis header
		head, err = model.NewHeader(guess, nil, proposal, s.host.ID().String(), nil)
		if err != nil {
			return err
		}
	case nil:
		salts := make([]string, len(head.Proposal.Chars))
		for i, ch := range head.Proposal.Chars {
			salts[i] = ch.Salt
		}

		hash, err := head.Hash()
		if err != nil {
			return err
		}

		head, err = model.NewHeader(guess, salts, proposal, s.host.ID().String(), hash)
		if err != nil {
			return err
		}
	}

	data, err := json.Marshal(head)
	if err != nil {
		return err
	}

	return s.topic.Publish(ctx, data)
}
