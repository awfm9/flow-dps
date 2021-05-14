# Flow Data Provisioning Service

[[!CI Status](https://github.com/awfm9/flow-dps/workflows/CI/badge.svg)](https://github.com/awfm9/flow-dps/actions/workflows/pr.yml)

The Flow Data Provisioning Service (DPS) aims at providing a scalable and efficient way to access the history of the Flow
execution state, both for the current live sporks and for past sporks. It also serves as a basis for the implementation
of the Rosetta Data API, used in the larger blockchain ecosystem as a common generic interface for blockchain integration.

The state of past sporks is indexed by reading an execution node's protocol state and state trie write-ahead log.
Optionally, a root checkpoint can be used to bootstrap state before a spork's start. In more specific terms, indexing
of past sporks requires a Badger key-value database containing the Flow protocol state of the spork and a LedgerWAL with
all the trie updates that happened on the spork.

Indexing the live spork works similarly. The DPS will connect to the publish socket of an execution node that has
whitelisted it and subscribe to state trie and transaction event updates. At the same time, the DPS will use access nodes
to assemble a view of the finalized blockchain state. By combining these two sources of information, it can reconstruct
the execution state on-the-fly.

The Flow DPS maintains multiple specialized indexes for different purposes. One index is used for accessing the entire
execution state at any given height, while another is used to follow the history of a specific Ledger register over time.
Contrary to the execution node's state trie, the indexes allow random access to the execution state at any block height
which enables state retrieval at any point in history and beyond the execution node's pruning limit.

The DPS also supports a set of custom smart contract resources that serve as wrapper for locked token vaults and as
proxy to staking and delegating resources. This allows the DPS to track multiple balances per account, including locked,
staked and delegated tokens, for accounts which deploy these custom resources.

## Dependencies

Go `v1.16` or higher is required to compile `flow-dps`.

If you want to make changes to the GRPC API, the two following binaries are required as well.

* [`protoc`](https://grpc.io/docs/protoc-installation/) version `3.17.0`
* `go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26`
* `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1`

Once they are installed, you can run `go generate ./...` from the root of this repository to update the generated protobuf files.

## Road Map

| Milestone |                  Description                  | Past Spork State | Past Spork Events | Live Spork State | Live Spork Events | Raw API | Ledger API | Rosetta API | Liquid Balance | Locked Balance | Staked Balance | Delegated Balance | State Verification | State Proofs | Event Proofs |
|:---------:|:---------------------------------------------:|:----------------:|-------------------|------------------|-------------------|---------|------------|-------------|----------------|----------------|----------------|-------------------|--------------------|--------------|--------------|
|    P.1    |        Past spork support for registers       |         X        |                   |                  |                   |    X    |      X     |             |        X       |                |                |                   |          X         |              |              |
|    P.2    |         Past spork support with events        |         X        |         X         |                  |                   |    X    |      X     |             |        X       |                |                |                   |          X         |              |              |
|    R.1    |    Rosetta API support for default balance    |         X        |         X         |                  |                   |    X    |      X     |      X      |        X       |                |                |                   |          X         |              |              |
|    L.1    |        Live spork support for registers       |         X        |         X         |         X        |                   |    X    |      X     |      X      |        X       |                |                |                   |          X         |              |              |
|    L.2    |         Live spork support with events        |         X        |         X         |         X        |         X         |    X    |      X     |      X      |        X       |                |                |                   |          X         |              |              |
|    R.2    | Rosetta API support with sub-account balances |         X        |         X         |         X        |         X         |    X    |      X     |      X      |        X       |        X       |        X       |         X         |          X         |              |              |
|    C.1    |       Cryptographic proofs for registers      |         X        |         X         |         X        |         X         |    X    |      X     |      X      |        X       |        X       |        X       |         X         |          X         |       X      |              |
|    C.2    |         Cryptographic proofs for events       |         X        |         X         |         X        |         X         |    X    |      X     |      X      |        X       |        X       |        X       |         X         |          X         |       X      |       X      |

## Architecture

### Components

The Flow Data Provisioning Service (DPS) is composed of five main components.

1. The **Chain** interface is responsible for reconstructing a view of the sequence of blocks, along with their metadata. It allows the consumer to step from the root block to the last sealed block, while presenting height, block identifier and state commitment for each step. The file i/o version does so by using the execution node's on-disk key-value store, while the network version relies on data retrieved from access nodes.
2. The **Feeder** interface is responsible for streaming in-order trie updates from different sources; the file i/o version reads them from the LedgerWAL, while the network version receives trie updates through its network subscription on the execution node.
3. The **Mapper** interface is responsible for mapping incoming state trie updates to blocks. Generally, trie updates come in by chunk, so each block maps from zero to multiple trie updates. Once a block is mapped to its respective trie updates, the mapper forwards the information to the indexer.
4. The **Store** interface is responsible for receiving a set of trie updates for each block and creating the necessary main indexes and auxiliary in the on-disk database. These indexes allow efficient retrieval of the state at arbitrary block heights in the state history. It also provides random access to the execution state by providing smart access to these indexes. It combines writing and retrieving of indexes, so that an efficient caching strategy is possible.

### Diagram

The following diagram provides a simple overview of the data flow for the DPS:

```text
┌─────────────────┐
│   Past Spork    │
│                 │
│ ┌─────────────┐ │
│ │  Exec Node  │ │
│ │             │ │  blocks
│ │ ┌─────────┐ │ │  transaction events
│ │ │Badger DB├─┼─┼───────────────────────────────────┐
│ │ └─────────┘ │ │                                   │
│ │             │ │  checkpoint                       │
│ │ ┌─────────┐ │ │  trie updates                     │
│ │ │LedgerWAL├─┼─┼─────────────────┐                 │
│ │ └─────────┘ │ │                 │                 │
│ └─────────────┘ │                 │                 │
└─────────────────┘                 │                 │
                                    ▼                 ▼                  ┌──────────────────┐
                           ┌──────────────────────────────────────────┐  │ REST/GRPC Client │
                           │   WALFeeder    ◄   Filesystem Chain      │  └────────┬─────────┘
                           ├───────▼▼▼──────┬───────────┬─────────────┤           │
                           │                │           ► Raw API     │◄──────────┘
                           │                │           ├─────────────┤              ┌────────────┐
                           │     Mapper     ►   Store   ► Ledger API  │◄─────────────┤ FVM Client │
                           │                │           ├─────────────┤              └────────────┘
                           │                │           ► Rosetta API │◄──────────┐
                           ├───────▲▲▲──────┴───────────┴─────────────┤           │
                           │  LiveFeeder    ◄     Network Chain       │  ┌────────┴────────┐
                           └──────────────────────────────────────────┘  │ Rosetta Client  │
                                    ▲                 ▲                  └─────────────────┘
                                    │                 │
┌─────────────────┐                 │                 │
│   Live Spork    │                 │                 │
│                 │                 │                 │
│ ┌─────────────┐ │                 │                 │
│ │  Exec Node  │ │                 │                 │
│ ├─────────────┤ │                 │                 │
│ │  Pub Socket ├─┼─────────────────┘                 │
│ └─────────────┘ │   trie updates                    │
│                 │   transaction events              │
│ ┌─────────────┐ │   bootstrapping info              │
│ │ Access Node │ │                                   │
│ ├─────────────┤ │                                   │
│ │ Access API  ├─┼───────────────────────────────────┘
│ └─────────────┘ │   blocks
└─────────────────┘
```
