package message

import (
	"encoding/binary"

	"github.com/bittorent-client/internal/torrentparser"
	"github.com/bittorent-client/internal/util"
)

// PieceBlock describes a block request within a piece.
type PieceBlock struct {
	Index  uint32
	Begin  uint32
	Length uint32
	Block  []byte
}

// Message is a parsed BitTorrent wire protocol message.
type Message struct {
	Size    uint32
	ID      *uint8
	Payload interface{}
}

// BuildHandshake constructs the initial handshake message.
func BuildHandshake(torrent torrentparser.Torrent) ([]byte, error) {
	buf := make([]byte, 68)
	buf[0] = 19
	copy(buf[1:20], "BitTorrent protocol")

	infoHash, err := torrentparser.InfoHash(torrent)
	if err != nil {
		return nil, err
	}
	copy(buf[28:48], infoHash)

	peerID := util.GenID()
	copy(buf[48:68], peerID[:])
	return buf, nil
}

// BuildInterested constructs an "interested" message.
func BuildInterested() []byte {
	return []byte{0, 0, 0, 1, 2}
}

// BuildRequest constructs a piece block request message.
func BuildRequest(payload PieceBlock) []byte {
	buf := make([]byte, 17)
	binary.BigEndian.PutUint32(buf[0:4], 13)
	buf[4] = 6
	binary.BigEndian.PutUint32(buf[5:9], payload.Index)
	binary.BigEndian.PutUint32(buf[9:13], payload.Begin)
	binary.BigEndian.PutUint32(buf[13:17], payload.Length)
	return buf
}

// Parse decodes a wire protocol message (post-handshake).
func Parse(msg []byte) Message {
	size := binary.BigEndian.Uint32(msg[0:4])
	if size == 0 {
		return Message{Size: 0}
	}

	id := msg[4]
	payload := msg[5:]

	switch id {
	case 6, 7, 8:
		block := PieceBlock{
			Index: binary.BigEndian.Uint32(payload[0:4]),
			Begin: binary.BigEndian.Uint32(payload[4:8]),
		}
		if id == 7 {
			block.Block = payload[8:]
		} else {
			block.Length = binary.BigEndian.Uint32(payload[8:12])
		}
		return Message{Size: size, ID: &id, Payload: block}
	default:
		return Message{Size: size, ID: &id, Payload: payload}
	}
}

// IsHandshake reports whether the buffer contains a complete handshake message.
func IsHandshake(msg []byte) bool {
	if len(msg) < 20 {
		return false
	}
	pstrlen := msg[0]
	expected := int(pstrlen) + 49
	if len(msg) != expected {
		return false
	}
	return string(msg[1:20]) == "BitTorrent protocol"
}
