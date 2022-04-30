package wordle

import (
	"context"
	"testing"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/sync"
	"github.com/libp2p/go-libp2p-core/event"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/p2p-games/wordle/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	const peers = 2

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	net, err := mocknet.FullMeshLinked(peers)
	require.NoError(t, err)

	servs := make([]*Service, peers)
	subs := make([]event.Subscription, peers)
	for i, h := range net.Hosts() {
		ds := sync.MutexWrap(datastore.NewMapDatastore())
		ps, err := pubsub.NewFloodSub(ctx, h, pubsub.WithMessageSignaturePolicy(pubsub.StrictNoSign))
		require.NoError(t, err)

		servs[i] = NewService(h, ds, ps)
		err = servs[i].Start()
		require.NoError(t, err)

		subs[i], err = net.Hosts()[0].EventBus().Subscribe(&event.EvtPeerIdentificationCompleted{})
		require.NoError(t, err)
	}

	err = net.ConnectAllButSelf()
	require.NoError(t, err)

	for _, sub := range subs {
		for range make([]bool, peers-1) {
			select {
			case <-sub.Out():
			case <-ctx.Done():
				t.Fatalf(ctx.Err().Error())
			}
		}
	}

	guesses := make([]<-chan *model.Header, peers)
	for i, serv := range servs {
		guesses[i], err = serv.Guesses(ctx)
		require.NoError(t, err)
	}

	var prev string
	for _, serv := range servs {
		prop := model.RandomString(5)
		err := serv.Guess(ctx, prev, prop)
		require.NoError(t, err)
		prev = prop
		time.Sleep(time.Millisecond*500)
	}

	for _, sub := range guesses {
		for range make([]bool, peers-1) {
			<-sub
		}
	}

	head, err := servs[0].store.Head(ctx)
	require.NoError(t, err)
	for _, serv := range servs[1:] {
		headCpr, err := serv.store.Head(ctx)
		require.NoError(t, err)
		assert.Equal(t, head, headCpr)
	}
}

