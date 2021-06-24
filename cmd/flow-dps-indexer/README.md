# Flow DPS Indexer

## Description

The Flow DPS Indexer binary implements the core functionality to create the index for past sporks.
It needs a reference to the protocol state database of the spork, as well as the trie directory and an execution state checkpoint.
The index is generated in the form of a Badger database that allows random access to any ledger register at any block height.

## Usage

```sh
Usage of flow-dps-indexer:
  -c, --checkpoint string   checkpoint file for state trie
  -d, --data string         database directory for protocol data
  -f, --force bool          overwrite existing index database (default "false")
  -i, --index string        database directory for state index (default "index")
  -l, --log string          log output level (default "info")
  -t, --trie string         data directory for state ledger
```

## Example

The below command line starts indexing a past spork from the on-disk information.

```sh
./flow-dps-indexer -d /var/flow/data/protocol -t /var/flow/data/execution -c /var/flow/bootstrap/root.checkpoint -i /var/flow/data/index
```