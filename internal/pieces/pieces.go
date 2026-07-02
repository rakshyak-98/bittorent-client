package pieces

import (
	"fmt"
	"os"

	"github.com/bittorent-client/internal/message"
	"github.com/bittorent-client/internal/torrentparser"
)

// Pieces tracks which blocks have been requested and received.
type Pieces struct {
	requested [][]bool
	received  [][]bool
}

// New creates a Pieces tracker for the given torrent.
func New(torrent torrentparser.Torrent) *Pieces {
	nPieces := torrentparser.NumPieces(torrent)
	build := func() [][]bool {
		arr := make([][]bool, nPieces)
		for i := range arr {
			nBlocks := torrentparser.BlocksPerPiece(torrent, i)
			arr[i] = make([]bool, nBlocks)
		}
		return arr
	}
	return &Pieces{
		requested: build(),
		received:  build(),
	}
}

func blockIndex(pieceBlock message.PieceBlock) int {
	return int(pieceBlock.Begin / torrentparser.BlockLen)
}

// AddRequest marks a block as requested.
func (p *Pieces) AddRequest(pieceBlock message.PieceBlock) {
	idx := blockIndex(pieceBlock)
	p.requested[pieceBlock.Index][idx] = true
}

// AddReceived marks a block as received.
func (p *Pieces) AddReceived(pieceBlock message.PieceBlock) {
	idx := blockIndex(pieceBlock)
	p.received[pieceBlock.Index][idx] = true
}

// Needed reports whether a block still needs to be requested.
func (p *Pieces) Needed(pieceBlock message.PieceBlock) bool {
	allRequested := true
	for _, blocks := range p.requested {
		for _, b := range blocks {
			if !b {
				allRequested = false
				break
			}
		}
		if !allRequested {
			break
		}
	}
	if allRequested {
		for i, blocks := range p.received {
			p.requested[i] = append([]bool(nil), blocks...)
		}
	}

	idx := blockIndex(pieceBlock)
	return !p.requested[pieceBlock.Index][idx]
}

// IsDone reports whether every block has been received.
func (p *Pieces) IsDone() bool {
	for _, blocks := range p.received {
		for _, b := range blocks {
			if !b {
				return false
			}
		}
	}
	return true
}

// PrintPercentDone writes download progress to stdout.
func (p *Pieces) PrintPercentDone() {
	downloaded := 0
	total := 0
	for _, blocks := range p.received {
		total += len(blocks)
		for _, b := range blocks {
			if b {
				downloaded++
			}
		}
	}
	if total == 0 {
		return
	}
	percent := downloaded * 100 / total
	fmt.Fprintf(os.Stdout, "progress: %d%%\r", percent)
}
