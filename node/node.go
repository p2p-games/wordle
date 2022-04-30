package node

import (
	"context"
	"fmt"
	"time"

	"github.com/ipfs/go-datastore"
	exchange "github.com/ipfs/go-ipfs-exchange-interface"
	format "github.com/ipfs/go-ipld-format"
	logging "github.com/ipfs/go-log/v2"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"go.uber.org/fx"

	"github.com/p2p-games/wordle/wordle"
)

var log = logging.Logger("node")

const LifecycleTimeout = time.Second * 15

type Node struct {
	Type Type

	Host         core.Host
	PubSub       *pubsub.PubSub
	Datastore    datastore.Batching
	ConnGater    connmgr.ConnectionGater
	Routing      routing.PeerRouting
	DataExchange exchange.Interface
	DAG          format.DAGService

	Wordle *wordle.Service

	start, stop lifecycleFunc
}

// New assembles a new Node with the given type 'tp' over Store 'store'.
func New(tp Type, store Store) (*Node, error) {
	cfg, err := store.Config()
	if err != nil {
		return nil, err
	}

	switch tp {
	case Light:
		fallthrough
	case Full:
		return newNode(baseComponents(cfg, store), fx.Supply(tp))
	default:
		panic("node: unknown Node Type")
	}
}

// Start launches the Node and all its components and services.
func (n *Node) Start(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, LifecycleTimeout)
	defer cancel()

	err := n.start(ctx)
	if err != nil {
		log.Errorf("starting %s Node: %s", n.Type, err)
		return fmt.Errorf("node: failed to start: %w", err)
	}

	addrs, err := peer.AddrInfoToP2pAddrs(host.InfoFromHost(n.Host))
	if err != nil {
		log.Errorw("Retrieving multiaddress information", "err", err)
		return err
	}
	fmt.Println("The p2p host is listening on:")
	for _, addr := range addrs {
		fmt.Println("* ", addr.String())
	}
	fmt.Println()
	return nil
}

// Run is a Start which blocks on the given context 'ctx' until it is canceled.
// If canceled, the Node is still in the running state and should be gracefully stopped via Stop.
func (n *Node) Run(ctx context.Context) error {
	err := n.Start(ctx)
	if err != nil {
		return err
	}

	<-ctx.Done()
	return ctx.Err()
}

// Stop shuts down the Node, all its running Components/Services and returns.
// Canceling the given context earlier 'ctx' unblocks the Stop and aborts graceful shutdown forcing remaining
// Components/Services to close immediately.
func (n *Node) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, LifecycleTimeout)
	defer cancel()

	err := n.stop(ctx)
	if err != nil {
		log.Errorf("Stopping %s Node: %s", n.Type, err)
		return err
	}

	log.Infof("stopped %s Node", n.Type)
	return nil
}

// newNode creates a new Node from given DI options.
// DI options allow initializing the Node with a customized set of components and services.
// NOTE: newNode is currently meant to be used privately to create various custom Node types e.g. Light, unless we
// decide to give package users the ability to create custom node types themselves.
func newNode(opts ...fx.Option) (*Node, error) {
	node := new(Node)
	app := fx.New(
		fx.NopLogger,
		fx.Extract(node),
		fx.Options(opts...),
	)
	if err := app.Err(); err != nil {
		return nil, err
	}

	node.start, node.stop = app.Start, app.Stop
	return node, nil
}

// lifecycleFunc defines a type for common lifecycle funcs.
type lifecycleFunc func(context.Context) error
