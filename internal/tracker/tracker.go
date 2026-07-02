package tracker

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/bittorent-client/internal/torrentparser"
	"github.com/bittorent-client/internal/util"
)

// Peer represents a discovered peer address.
type Peer struct {
	IP   string
	Port int
}

// GetPeers announces to the UDP tracker and returns the peer list.
func GetPeers(torrent torrentparser.Torrent) ([]Peer, error) {
	announceURL := torrentparser.AnnounceURL(torrent)
	if announceURL == "" {
		return nil, fmt.Errorf("torrent has no announce URL")
	}

	u, err := url.Parse(announceURL)
	if err != nil {
		return nil, err
	}

	port := 6891
	if u.Port() != "" {
		fmt.Sscanf(u.Port(), "%d", &port)
	} else if u.Scheme == "http" || u.Scheme == "https" {
		port = 80
	}

	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", u.Hostname(), port))
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	connReq := buildConnReq()
	if _, err := conn.Write(connReq); err != nil {
		return nil, err
	}

	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	connResp, err := readUDPResponse(conn)
	if err != nil {
		return nil, fmt.Errorf("connect response: %w", err)
	}

	parsedConn := parseConnResp(connResp)
	announceReq, err := buildAnnounceReq(parsedConn.connectionID, torrent)
	if err != nil {
		return nil, err
	}

	if _, err := conn.Write(announceReq); err != nil {
		return nil, err
	}

	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	announceResp, err := readUDPResponse(conn)
	if err != nil {
		return nil, fmt.Errorf("announce response: %w", err)
	}

	return parseAnnounceResp(announceResp), nil
}

func readUDPResponse(conn *net.UDPConn) ([]byte, error) {
	buf := make([]byte, 1500)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func buildConnReq() []byte {
	buf := make([]byte, 16)
	binary.BigEndian.PutUint32(buf[0:4], 0x417)
	binary.BigEndian.PutUint32(buf[4:8], 0x27101980)
	binary.BigEndian.PutUint32(buf[8:12], 0)
	rand.Read(buf[12:16])
	return buf
}

func buildAnnounceReq(connID []byte, torrent torrentparser.Torrent) ([]byte, error) {
	buf := make([]byte, 98)

	copy(buf[0:8], connID)
	binary.BigEndian.PutUint32(buf[8:12], 1)
	rand.Read(buf[12:16])

	infoHash, err := torrentparser.InfoHash(torrent)
	if err != nil {
		return nil, err
	}
	copy(buf[16:36], infoHash)

	peerID := util.GenID()
	copy(buf[36:56], peerID[:])

	// downloaded
	// left
	binary.BigEndian.PutUint64(buf[64:72], uint64(torrentparser.Size(torrent)))
	// uploaded

	binary.BigEndian.PutUint32(buf[80:84], 0) // event: none
	binary.BigEndian.PutUint32(buf[84:88], 0) // IP address: default
	rand.Read(buf[88:92])
	binary.BigEndian.PutUint32(buf[92:96], 0xFFFFFFFF) // num_want: -1
	binary.BigEndian.PutUint16(buf[96:98], 6881)

	return buf, nil
}

type connResponse struct {
	action         uint32
	transactionID  uint32
	connectionID   []byte
}

func parseConnResp(resp []byte) connResponse {
	return connResponse{
		action:        binary.BigEndian.Uint32(resp[0:4]),
		transactionID: binary.BigEndian.Uint32(resp[4:8]),
		connectionID:  resp[8:16],
	}
}

func parseAnnounceResp(resp []byte) []Peer {
	peerData := resp[20:]
	var peers []Peer
	for i := 0; i+6 <= len(peerData); i += 6 {
		ip := fmt.Sprintf("%d.%d.%d.%d", peerData[i], peerData[i+1], peerData[i+2], peerData[i+3])
		port := int(binary.BigEndian.Uint16(peerData[i+4 : i+6]))
		peers = append(peers, Peer{IP: ip, Port: port})
	}
	return peers
}
