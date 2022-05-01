# Wordle
> An async Wordle game solely on p2p.

## What
A fully distributed p2p implementation of the popular Wordle game with a few amendments to the rules. Instead of having 
a centralized authority which defines a word that everyone guesses within a day, we implement a simple asynchronous 
'first come, first serve' consensus, where a guesser becomes a new leader and chooses the next word for the whole 
network to guess.

## Why
This project is a [hackaton](https://p2p.paris/en/event/hackathon-1/) submission written in two days. The original goal
was to implement something based on the [Lazy Client for Lazy Blockchains paper](https://arxiv.org/abs/2203.15968), and 
we decided to implement a game with global state transitions, like Wordle.

### Side Quests
- PL Side quest as we rely on public IPFS network and using libp2p heavily

## How
The protocol can be explained as a simple permissionless blockchain with two actors, Light and Full(WIP) Nodes. The
Light Nodes are actual players participating in the global consensus by solving each other word puzzles and growing the 
canonical chain. The Full Nodes are mainly the infrastructure nodes that sync and serve the whole chain to the network. 
The essential property of the protocol is that Light Nodes only need to access the latest state header to interact with the network in a trust minimized manner without the requirement to sync the whole chain, but with an assumption that it connects to at least one honest Full Node.

## Play
* Clone it
* `make build` it
* `./build/wordle light start` it
* Wait until you discover peers
* Play it

## Comments for reviewers
* The actual protocol is in `./wordle` pkg
* `node`, `libs`, `cmd` are mostly boilerplate code, mostly unrelated to the protocol itself
* `wordle.Service` uses a relatively new WIP tool - [`go-libp2p-messagner`](https://github.com/celestiaorg/go-libp2p-messenger), 
which abstracts away libp2p streams, with (subjectively) simpler API. Mainly, we need it to for the `Broadcast` feature
that enable message sending to all dynamically changing immediate peers over long-lived streams.
* The protocol is *not fully secure yet*. 
  * If a Light Node detects suspicious behavior, e.g. different immediate peers tell different network state, it hangs.
Going further, [the paper](https://arxiv.org/abs/2203.15968) describes a novel way on how to resolve such disputes in
efficient manner.
  * Proposer/Guesser signing is missing
  * The protocol relies on the longest chain fork-choice rule, meaning that the chain with the bigger amount of guessed words is preffered by the protocol. Unfortunately, word guessing can be easily brutforced, s.t. an attacker can precompute a fork with a longer chain that everyone will eventually switch to. 
  * ...
* Message propagation is done with PubSub
* Peer/Topic discover is done over kDHT with management delegated to PubSub's internal discovery feature

## Future Work
* [ ] Finish dispute resolution imeplemtnation
* [ ] Finish implementation of the Full Node
* [ ] Leaderboard with UTXO based state machine
* [ ] Move towards generalization of the protocol
 * Generic Consensus interface for swappable/composable consensuses
 * Generic dispute resolution protocol
* [ ] Use [Celestia](https://celestia.org/) as DA layer 🤔 😁

## Team
* [@Wondertan](https://github.com/Wondertan)
* [@cortze](https://github.com/cortze)
* [@ajnavarro](https://github.com/ajnavarro)
