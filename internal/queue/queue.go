package queue

import (
	"github.com/bittorent-client/internal/message"
	"github.com/bittorent-client/internal/torrentparser"
)

// Queue holds piece blocks available from a single peer.
type Queue struct {
	torrent torrentparser.Torrent
	blocks  []message.PieceBlock
	Choked  bool
}

// New creates an empty per-peer download queue.
func New(torrent torrentparser.Torrent) *Queue {
	return &Queue{
		torrent: torrent,
		Choked:  true,
	}
}

// EnqueuePiece adds all blocks for a piece index to the queue.
func (q *Queue) EnqueuePiece(pieceIndex int) {
	nBlocks := torrentparser.BlocksPerPiece(q.torrent, pieceIndex)
	for i := 0; i < nBlocks; i++ {
		q.blocks = append(q.blocks, message.PieceBlock{
			Index:  uint32(pieceIndex),
			Begin:  uint32(i * torrentparser.BlockLen),
			Length: uint32(torrentparser.BlockSize(q.torrent, pieceIndex, i)),
		})
	}
}

// Dequeue removes and returns the next block in the queue.
func (q *Queue) Dequeue() (message.PieceBlock, bool) {
	if len(q.blocks) == 0 {
		return message.PieceBlock{}, false
	}
	block := q.blocks[0]
	q.blocks = q.blocks[1:]
	return block, true
}

// Len returns the number of blocks waiting in the queue.
func (q *Queue) Len() int {
	return len(q.blocks)
}
