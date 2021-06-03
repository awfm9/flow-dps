# Flow Rosetta Server

## Description

The Flow Rosetta Server implements the Rosetta Data API specifications for the Flow network.
It uses the Flow DPS Server's GRPC API as the backend to query the required data.
Flow core contract addresses are derived from the chain ID with which the service is started.
This allows the Rosetta API to access state remotely, or locally by running the Flow DPS Server on the same host.

## Usage

```sh
Usage of flow-dps-rosetta:
  -a, --api string     host URL for GRPC API endpoint (default "127.0.0.1:5005")
  -c, --chain string   chain ID for Flow network core contracts (default "flow-testnet")
  -l, --log string     log output level (default "info")
  -p, --port uint16    port to serve Rosetta API on (default 8080)
```

## Example

The following command line starts the Flow DPS Rosetta server for a main network spork on port 8080.
It uses a local instance of the Flow DPS Server for access to the execution state.

```sh
./flow-dps-rosetta -a "127.0.0.1:5005" -c "flow-mainnet" -p 8080
```