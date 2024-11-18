"use strict";

const fs = require("node:fs");
const crypto = require("node:crypto");
const benCode = require("bencode");
const bigNum = require("bignumber");

/*
 * smallest distribution of file size (piece) length, 2^14 is the block size.
 * its the block size the amount of data transferred between peers.
 * minimize the chances of corruption
 */
module.exports.BLOCK_LEN = Math.pow(2, 14);

module.exports.open = (filepath) => {
	return benCode.decode(fs.readFileSync(filepath));
};

/*
 * Generate info_hash for Tracker Connection Request.
 *  bittorrent ideal with sha1 encryption.
 */
module.exports.infoHash = (torrent) => {
	const info = benCode.encode(torrent.info);
	return crypto.createHash("sha1").update(info).digest();
};

module.exports.size = (torrent) => {
	const size = torrent.info.files
		? torrent.info.files.map((file) => file.length).reduce((a, b) => a + b)
		: torrent.info.length;
	return bigNum.toBuffer(size, { size: 8 });
};

module.exports.pieceLen = (torrent, pieceIndex) => {
	const totalLength = bigNum.fromBuffer(this.size(torrent)).toNumber();
	const pieceLength = torrent.info["piece length"];
	const lastPieceLength = totalLength % pieceLength;
	let lastPieceIndex = Math.floor(totalLength / pieceLength);

	return (lastPieceIndex = pieceIndex ? lastPieceLength : pieceLength);
};

module.exports.blocksPerPiece = (torrent, pieceIndex) => {
	const pieceLength = this.pieceLen(torrent, pieceIndex);
	return Math.ceil(pieceLength / this.BLOCK_LEN);
};

module.exports.blockLen = (torrent, pieceIndex, blockIndex) => {
	const pieceLength = this.pieceLen(torrent, pieceIndex);

	const lastPieceLength = pieceLength % this.BLOCK_LEN;
	const lastPieceIndex = Math.floor(pieceLength / this.BLOCK_LEN);

	return (blockIndex = lastPieceIndex ? lastPieceLength : this.BLOCK_LEN);
};
