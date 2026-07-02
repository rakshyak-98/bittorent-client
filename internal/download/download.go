package download

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"github.com/bittorent-client/internal/message"
	"github.com/bittorent-client/internal/pieces"
	"github.com/bittorent-client/internal/queue"
	"github.com/bittorent-client/internal/torrentparser"
	"github.com/bittorent-client/internal/tracker"
)

// Download discovers peers and downloads the torrent to the given output path.
func Download(torrent torrentparser.Torrent, outputPath string) error {
	peers, err := tracker.GetPeers(torrent)
	if err != nil {
		return fmt.Errorf("get peers: %w", err)
	}
	if len(peers) == 0 {
		return fmt.Errorf("no peers returned from tracker")
	}

	p := pieces.New(torrent)
	file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open output file: %w", err)
	}
	defer file.Close()

	var wg sync.WaitGroup
	var closeOnce sync.Once
	var closeErr error

	for _, peer := range peers {
		wg.Add(1)
		go func(peer tracker.Peer) {
			defer wg.Done()
			downloadFromPeer(peer, torrent, p, file, &closeOnce, &closeErr)
		}(peer)
	}

	wg.Wait()
	return closeErr
}

func downloadFromPeer(peer tracker.Peer, torrent torrentparser.Torrent, p *pieces.Pieces, file *os.File, closeOnce *sync.Once, closeErr *error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(peer.IP, fmt.Sprintf("%d", peer.Port)))
	if err != nil {
		fmt.Printf("connect %s:%d: %v\n", peer.IP, peer.Port, err)
		return
	}
	defer conn.Close()

	handshake, err := message.BuildHandshake(torrent)
	if err != nil {
		fmt.Printf("build handshake: %v\n", err)
		return
	}
	if _, err := conn.Write(handshake); err != nil {
		fmt.Printf("write handshake: %v\n", err)
		return
	}

	q := queue.New(torrent)
	onWholeMessage(conn, func(msg []byte) {
		msgHandler(msg, conn, torrent, p, q, file, closeOnce, closeErr)
	})
}

func onWholeMessage(conn net.Conn, handler func([]byte)) {
	var saveBuf []byte
	handshake := true

	buf := make([]byte, 32*1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("read: %v\n", err)
			}
			return
		}
		saveBuf = append(saveBuf, buf[:n]...)

		for {
			msgLen, ok := messageLength(saveBuf, handshake)
			if !ok || len(saveBuf) < msgLen {
				break
			}
			handler(saveBuf[:msgLen])
			saveBuf = saveBuf[msgLen:]
			handshake = false
		}
	}
}

func messageLength(buf []byte, handshake bool) (int, bool) {
	if len(buf) < 4 {
		return 0, false
	}
	if handshake {
		if len(buf) < 1 {
			return 0, false
		}
		return int(buf[0]) + 49, true
	}
	length := int(uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3]))
	return length + 4, true
}

func msgHandler(msg []byte, conn net.Conn, torrent torrentparser.Torrent, p *pieces.Pieces, q *queue.Queue, file *os.File, closeOnce *sync.Once, closeErr *error) {
	if message.IsHandshake(msg) {
		conn.Write(message.BuildInterested())
		return
	}

	m := message.Parse(msg)
	if m.ID == nil {
		return
	}

	switch *m.ID {
	case 0:
		chokeHandler(conn)
	case 1:
		unchokeHandler(conn, p, q)
	case 4:
		haveHandler(m.Payload.([]byte), conn, p, q)
	case 5:
		bitfieldHandler(m.Payload.([]byte), conn, p, q)
	case 7:
		pieceHandler(m.Payload.(message.PieceBlock), conn, torrent, p, q, file, closeOnce, closeErr)
	}
}

func chokeHandler(conn net.Conn) {
	conn.Close()
}

func unchokeHandler(conn net.Conn, p *pieces.Pieces, q *queue.Queue) {
	q.Choked = false
	requestPiece(conn, p, q)
}

func haveHandler(payload []byte, conn net.Conn, p *pieces.Pieces, q *queue.Queue) {
	if len(payload) < 4 {
		return
	}
	pieceIndex := int(uint32(payload[0])<<24 | uint32(payload[1])<<16 | uint32(payload[2])<<8 | uint32(payload[3]))
	queueEmpty := q.Len() == 0
	q.EnqueuePiece(pieceIndex)
	if queueEmpty {
		requestPiece(conn, p, q)
	}
}

func bitfieldHandler(payload []byte, conn net.Conn, p *pieces.Pieces, q *queue.Queue) {
	queueEmpty := q.Len() == 0
	for i, b := range payload {
		for j := 0; j < 8; j++ {
			if b&1 == 1 {
				q.EnqueuePiece(i*8 + 7 - j)
			}
			b >>= 1
		}
	}
	if queueEmpty {
		requestPiece(conn, p, q)
	}
}

func pieceHandler(block message.PieceBlock, conn net.Conn, torrent torrentparser.Torrent, p *pieces.Pieces, q *queue.Queue, file *os.File, closeOnce *sync.Once, closeErr *error) {
	p.PrintPercentDone()
	p.AddReceived(block)

	offset := int64(block.Index)*int64(torrentparser.PieceLength(torrent)) + int64(block.Begin)
	if _, err := file.WriteAt(block.Block, offset); err != nil {
		fmt.Printf("write piece: %v\n", err)
	}

	if p.IsDone() {
		closeOnce.Do(func() {
			fmt.Println("\nDownload complete")
			conn.Close()
		})
		return
	}

	requestPiece(conn, p, q)
}

func requestPiece(conn net.Conn, p *pieces.Pieces, q *queue.Queue) {
	if q.Choked {
		return
	}

	for q.Len() > 0 {
		block, ok := q.Dequeue()
		if !ok {
			return
		}
		if p.Needed(block) {
			conn.Write(message.BuildRequest(block))
			p.AddRequest(block)
			return
		}
	}
}
