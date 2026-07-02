package main

import (
	"fmt"
	"os"

	"github.com/bittorent-client/internal/download"
	"github.com/bittorent-client/internal/torrentparser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <torrent-file>\n", os.Args[0])
		os.Exit(1)
	}

	torrent, err := torrentparser.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open torrent: %v\n", err)
		os.Exit(1)
	}

	outputPath := torrentparser.InfoName(torrent)
	if err := download.Download(torrent, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "download: %v\n", err)
		os.Exit(1)
	}
}
