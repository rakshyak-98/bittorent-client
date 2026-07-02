package torrentparser

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/jackpal/bencode-go"
)

const BlockLen = 1 << 14 // 16384 bytes

// Torrent is a decoded .torrent file.
type Torrent map[string]interface{}

// Open reads and decodes a .torrent file from disk.
func Open(filepath string) (Torrent, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := bencode.Decode(f)
	if err != nil {
		return nil, err
	}
	torrent, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid torrent file")
	}
	return Torrent(torrent), nil
}

// InfoHash returns the 20-byte SHA-1 hash of the torrent info dictionary.
func InfoHash(torrent Torrent) ([]byte, error) {
	info, ok := torrent["info"].(map[string]interface{})
	if !ok {
		return nil, os.ErrInvalid
	}
	var buf bytes.Buffer
	if err := bencode.Marshal(&buf, info); err != nil {
		return nil, err
	}
	h := sha1.Sum(buf.Bytes())
	return h[:], nil
}

// Size returns the total size of all files in the torrent.
func Size(torrent Torrent) int64 {
	info := torrent["info"].(map[string]interface{})
	if files, ok := info["files"].([]interface{}); ok {
		var total int64
		for _, f := range files {
			file := f.(map[string]interface{})
			total += file["length"].(int64)
		}
		return total
	}
	return info["length"].(int64)
}

// PieceLength returns the configured piece length from the torrent metadata.
func PieceLength(torrent Torrent) int {
	info := torrent["info"].(map[string]interface{})
	return int(info["piece length"].(int64))
}

// NumPieces returns the number of pieces in the torrent.
func NumPieces(torrent Torrent) int {
	info := torrent["info"].(map[string]interface{})
	pieces := info["pieces"].(string)
	return len(pieces) / 20
}

// PieceLen returns the byte length of a specific piece index.
func PieceLen(torrent Torrent, pieceIndex int) int {
	totalLength := Size(torrent)
	pieceLength := PieceLength(torrent)
	lastPieceIndex := totalLength / int64(pieceLength)
	lastPieceLength := int(totalLength % int64(pieceLength))
	if lastPieceLength == 0 {
		lastPieceLength = pieceLength
	}
	if int64(pieceIndex) == lastPieceIndex {
		return lastPieceLength
	}
	return pieceLength
}

// BlocksPerPiece returns how many blocks comprise a given piece.
func BlocksPerPiece(torrent Torrent, pieceIndex int) int {
	return (PieceLen(torrent, pieceIndex) + BlockLen - 1) / BlockLen
}

// BlockSize returns the byte length of a specific block within a piece.
func BlockSize(torrent Torrent, pieceIndex, blockIndex int) int {
	pieceLength := PieceLen(torrent, pieceIndex)
	lastBlockIndex := pieceLength / BlockLen
	lastBlockLength := pieceLength % BlockLen
	if lastBlockLength == 0 {
		lastBlockLength = BlockLen
	}
	if blockIndex == lastBlockIndex {
		return lastBlockLength
	}
	return BlockLen
}

// InfoName returns the name field from the torrent info dictionary.
func InfoName(torrent Torrent) string {
	info := torrent["info"].(map[string]interface{})
	return info["name"].(string)
}

// AnnounceURL returns the tracker announce URL as a string.
func AnnounceURL(torrent Torrent) string {
	switch v := torrent["announce"].(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}
