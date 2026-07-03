# BitTorrent Client (Go)

A small BitTorrent downloader written in Go. It reads a `.torrent` file, discovers peers through a UDP tracker, connects to peers over TCP, and downloads piece blocks into an output file.

This repository is intentionally compact and is useful both as a learning project and as a minimal reference implementation for core BitTorrent client behavior.

## What it does

- Parses `.torrent` metadata using bencode
- Computes the torrent `info_hash`
- Announces to the tracker from the `announce` field
- Connects to discovered peers using the BitTorrent peer wire protocol
- Handles handshake, interested, choke/unchoke, have, bitfield, and piece messages
- Splits pieces into 16 KiB blocks and writes received blocks to disk
- Tracks requested vs received blocks so missing blocks can be retried

## Current scope and limitations

This project implements a minimal downloader, not a full-featured BitTorrent client.

- Tracker discovery is based on the single `announce` URL and expects a UDP tracker flow
- HTTP trackers, DHT, PEX, magnet links, resume support, and seeding are not implemented
- Downloaded data is written to `info.name`
- Piece hash verification is not currently performed after blocks are written
- Multi-file torrents are not reconstructed into their original directory layout

## Requirements

- Go `1.26.3` or later

## Build

```bash
go build -o bittorrent-client .
```

## Usage

```bash
./bittorrent-client path/to/file.torrent
```

The client writes the downloaded bytes to the file name stored in the torrent metadata (`info.name`) in the current working directory.

## How it works

1. `main.go` reads the torrent path from the CLI and opens the torrent file.
2. `internal/torrentparser` decodes bencoded metadata and exposes helpers such as:
   - torrent size
   - piece length
   - piece count
   - block sizing
   - `info_hash`
3. `internal/tracker` sends the UDP tracker connect and announce requests, then parses compact peer responses.
4. `internal/download` connects to all returned peers and starts a download loop per peer.
5. `internal/message` builds and parses peer wire protocol messages.
6. `internal/queue` keeps a per-peer queue of blocks that peer can provide.
7. `internal/pieces` tracks which blocks have been requested and received across the download.
8. Received blocks are written to the correct file offsets with `WriteAt`.

## Project structure

```text
.
├── main.go                        # CLI entry point
├── go.mod                         # Module definition and dependencies
├── README.md
└── internal
    ├── download
    │   └── download.go            # Peer connection handling and download orchestration
    ├── message
    │   └── message.go             # Handshake, request, interested, and message parsing
    ├── pieces
    │   └── pieces.go              # Requested/received block tracking and progress reporting
    ├── queue
    │   └── queue.go               # Per-peer queue of downloadable blocks
    ├── torrentparser
    │   └── torrentparser.go       # Torrent decoding, info hash, piece math, block sizing
    ├── tracker
    │   └── tracker.go             # UDP tracker connect/announce and peer parsing
    └── util
        └── util.go                # Stable peer ID generation
```

## Package overview

### `main`

Starts the program, validates CLI input, opens the torrent file, and calls the downloader.

### `internal/torrentparser`

Responsible for decoding `.torrent` files and exposing metadata helpers used throughout the client:

- `Open`
- `InfoHash`
- `Size`
- `PieceLength`
- `NumPieces`
- `PieceLen`
- `BlocksPerPiece`
- `BlockSize`
- `InfoName`
- `AnnounceURL`

### `internal/tracker`

Implements tracker communication:

- builds the UDP connect request
- parses the tracker connection response
- builds the announce request using `info_hash`, peer ID, torrent size, and port
- parses compact peer lists into `IP:port` pairs

### `internal/message`

Defines the wire-level message helpers used for peer communication:

- handshake creation
- interested message creation
- block request creation
- message parsing for piece-related payloads

### `internal/queue`

Stores the blocks advertised by a single peer. A peer queue is filled from `have` and `bitfield` messages, then drained one block at a time as requests are sent.

### `internal/pieces`

Tracks global download state:

- which blocks have already been requested
- which blocks have actually been received
- whether the torrent is complete
- a simple percentage-based progress display

### `internal/download`

Coordinates the full download process:

- gets peers from the tracker
- opens the destination file
- dials each peer concurrently
- sends the initial handshake and interested message
- reacts to choke/unchoke, have, bitfield, and piece messages
- writes piece blocks to their offsets in the output file

## Protocol notes

The implementation follows the standard BitTorrent flow at a high level:

1. Read torrent metadata
2. Compute `info_hash`
3. Announce to a tracker
4. Receive peers
5. Connect to peers over TCP
6. Exchange handshake
7. Send `interested`
8. Wait for `unchoke`
9. Request blocks
10. Receive piece data and write it to disk

Blocks are requested in `16 KiB` chunks (`1 << 14`), which is the conventional block size used by BitTorrent clients.

## Dependency

This project currently uses:

- [`github.com/jackpal/bencode-go`](https://github.com/jackpal/bencode-go) for bencode decoding and encoding

## Running locally

With a reachable UDP tracker and active peers available for the torrent:

```bash
go run . ./example.torrent
```

During download, the client prints progress updates and ends with `Download complete` when every tracked block has been received.

## Notes for future improvements

Some natural next steps for the project are:

- piece hash verification against the torrent metadata
- support for HTTP trackers and multiple trackers
- better retry and timeout handling for tracker and peer communication
- proper support for multi-file torrents
- resume support and persistent state
- seeding/upload behavior

