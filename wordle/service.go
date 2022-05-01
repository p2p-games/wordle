package wordle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	msngr "github.com/celestiaorg/go-libp2p-messenger"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	core "github.com/libp2p/go-libp2p-core"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/p2p-games/wordle/model"
)

var log = logging.Logger("wordle")

var topic = "wordle"

var protoID protocol.ID = "/wordle/v0.0.1"

// TODO(@Wondertan); If we are Full Node, sync every header

type Service struct {
	store  *Store
	host   core.Host
	pubsub *pubsub.PubSub
	topic  *pubsub.Topic

	// TODO(@Wondertan): improve messenger so it can handle msg types, thus avoiding the requirement to make an instance
	//  for a type
	reqs, resps *msngr.Messenger

	bootsrapped chan struct{}
	cancel      context.CancelFunc
}

func NewService(host core.Host, ds datastore.Batching, pubsub *pubsub.PubSub) *Service {
	reqs, err := msngr.New(host, msngr.WithProtocols(protoID+"/req"), msngr.WithMessageType(&HeaderRequest{}))
	if err != nil {
		panic(err)
	}
	resps, err := msngr.New(host, msngr.WithProtocols(protoID+"/resp"), msngr.WithMessageType(&HeaderResponse{}))
	if err != nil {
		panic(err)
	}
	return &Service{
		store:       NewStore(ds),
		host:        host,
		pubsub:      pubsub,
		reqs:        reqs,
		resps:       resps,
		bootsrapped: make(chan struct{}),
	}
}

func (s *Service) Start(ctx context.Context) (err error) {
	s.topic, err = s.pubsub.Join(topic)
	if err != nil {
		return err
	}

	err = s.pubsub.RegisterTopicValidator(topic, s.validate)
	if err != nil {
		return err
	}

	ctx, s.cancel = context.WithCancel(context.Background())
	go s.bootstrap(ctx)
	go s.listen(ctx)
	fmt.Println("Started P2P Wordle")
	return nil
}

func (s *Service) Stop(context.Context) error {
	s.cancel()
	s.host.RemoveStreamHandler(protoID)
	err := s.pubsub.UnregisterTopicValidator(topic)
	if err != nil {
		return err
	}

	err = s.reqs.Close()
	if err != nil {
		return err
	}

	err = s.resps.Close()
	if err != nil {
		return err
	}

	return s.topic.Close()
}

func (s *Service) Guess(ctx context.Context, guess, proposal string) error {
	select {
	case <-s.bootsrapped:
	case <-ctx.Done():
		return ctx.Err()
	}

	head, err := s.store.Head(ctx)
	if err != nil {
		return err
	}

	head, err = model.NewHeader(head, guess, proposal, s.host.ID().String())
	if err != nil {
		return err
	}

	data, err := json.Marshal(head)
	if err != nil {
		return err
	}

	return s.topic.Publish(ctx, data)
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

			if peer.ID(msg.From) == s.host.ID() {
				// ignore self messages
				continue
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

func (s *Service) validate(ctx context.Context, _ peer.ID, msg *pubsub.Message) pubsub.ValidationResult {
	proposal := &model.Header{}
	err := json.Unmarshal(msg.Data, proposal)
	if err != nil {
		log.Errorw("unmarshalling proposal", "err", err)
		return pubsub.ValidationReject
	}

	head, err := s.store.Head(ctx)
	if err != nil {
		log.Errorw("getting local head", "err", err)
		return pubsub.ValidationIgnore
	}

	if head.Height < proposal.Height && model.Verify(head.Proposal, proposal.Guess) {
		err = s.store.Append(ctx, proposal)
		if err != nil {
			log.Errorw("appending the successful proposal", "err", err)
			return pubsub.ValidationIgnore
		}

		log.Debugf("rcvd successful guess")
	} else {
		log.Debugf("rcvd unsuccessful guess")
	}

	// we allow unsuccessful guesses to be passed around the network, but we store only successful ones
	return pubsub.ValidationAccept
}

func (s *Service) bootstrap(ctx context.Context) {
	// ensure we discovered some peers to sync from
	// discovery is done automagically by PubSub
	// we just wait here until we discover and connect us to at least one peer for now
	s.ensurePeers(ctx)

	headers := s.askPeers(ctx)
	if len(headers) == 0 {
		// this means our peers does not have a height higher than ours, so we are done
		close(s.bootsrapped)
		return
	}

	// now, find if there is mismatch between headers on the same height
	newHead := headers[0]
	hashA, err := newHead.Hash()
	if err != nil {
		return
	}
	for _, h := range headers {
		hashB, err := h.Hash()
		if err != nil {
			return
		}

		if !bytes.Equal(hashA, hashB) {
			// TODO(@Wondertan):
			//  The whole point of this project was to implement p2p IVGs to make trust minimized access to the latest
			//  state from the light clients. However, we don't have enough time to make this and we simply do
			//  nothing when there is a mismatch between information. It shouldn't be hard to add the verification part
			//  in here at later point.
			fmt.Println(`
Peers we are connected, told us different information about the network state.
Something suspicious is happening. Just don't do anything for now, until we implement dispute resolution.
			`)
			return
		}
	}

	err = s.store.Append(ctx, newHead)
	if err != nil {
		log.Errorw("appending header", "err", err)
		return
	}

	fmt.Printf("Updated the state! New height is %d. 'Guess what?' \n", newHead.Height)
	close(s.bootsrapped)
}

func (s *Service) ensurePeers(ctx context.Context) {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		if len(s.reqs.Peers()) >= 1 {
			fmt.Println("Yay! Discovered some peers")
			return
		}

		select {
		case <-t.C:
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) askPeers(ctx context.Context) []*model.Header {
	s.reqs.Broadcast(ctx, &HeaderRequest{Height: 0}) // request status from every connected peer

	head, err := s.store.Head(ctx)
	if err != nil {
		log.Errorw("getting head", "err", err)
		return nil
	}

	fmt.Printf("JFYI, anon, we are on the height %d \n", head.Height)

	height := head.Height
	headers := make(map[int][]*model.Header)
	for _, p := range s.reqs.Peers() {
		s.host.ConnManager().TagPeer(p, topic, 100)
		msg, _, err := s.resps.Receive(ctx)
		if err != nil {
			return nil
		}

		h := msg.(*HeaderResponse).Header
		if h == nil {
			continue
		}

		if h.Height > height {
			height = h.Height
			headers[height] = append(headers[height], h)
		}
	}

	return headers[height]
}

func (s *Service) listen(ctx context.Context) {
	for {
		msg, from, err := s.reqs.Receive(ctx)
		if err != nil {
			return
		}
		req := msg.(*HeaderRequest)

		resp := &HeaderResponse{}
		switch req.Height {
		case 0:
			resp.Header, err = s.store.Head(ctx)
			if err != nil {
				log.Errorw("getting head", "err", err)
				continue
			}
		default:
			resp.Header, err = s.store.Get(ctx, req.Height)
			if err != nil {
				log.Errorw("getting header", "height", req.Height, "err", err)
				continue
			}
		}

		err = <-s.resps.Send(ctx, resp, from)
		if err != nil {
			log.Errorw("responding peer", "peer", from, "err", err)
			continue
		}
	}
}

type HeaderRequest struct {
	Height int // 0 means give me the latest
}

func (h *HeaderRequest) Size() int {
	data, _ := json.Marshal(h) // super stupidm but it is what it is right now
	return len(data)
}

func (h *HeaderRequest) MarshalTo(bytes []byte) (int, error) {
	data, err := json.Marshal(h)
	if err != nil {
		return 0, err
	}
	// Note to myself: ugly, buuuut, it is hackaton
	return copy(bytes, data), nil
}

func (h *HeaderRequest) Unmarshal(bytes []byte) error {
	return json.Unmarshal(bytes, h)
}

type HeaderResponse struct {
	Header *model.Header
}

func (h *HeaderResponse) Size() int {
	data, _ := json.Marshal(h) // super stupid but it is what it is right now
	return len(data)
}

func (h *HeaderResponse) MarshalTo(bytes []byte) (int, error) {
	data, err := json.Marshal(h)
	if err != nil {
		return 0, err
	}
	// Note to myself: ugly, buuuut, it is hackaton
	return copy(bytes, data), nil
}

func (h *HeaderResponse) Unmarshal(bytes []byte) error {
	return json.Unmarshal(bytes, h)
}
