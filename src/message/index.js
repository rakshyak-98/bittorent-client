"use strict";

const Buffer = require("node:buffer").Buffer;
const torrentParser = require("../models/torrentParser.js");
const util = require("../util.js");
const { SIZE } = require("../constants/index.js")

module.exports.buildHandshake = (torrent) => {
	const buf = Buffer.alloc(SIZE.HANDSHAKE);
	// pstr len
	buf.writeUInt8(SIZE.PROTOCOL_IDENTIFIER, 0);
	// pstr
	buf.write("BitTorrent protocol", 1);
	// reserver
	buf.writeUInt32BE(0, 20); // write 4 bytes of zero starting at byte 20
	buf.writeUInt32BE(0, 24); // write 4 bytes of zero starting at byte 24

	// info hash
	torrentParser.infoHash(torrent).copy(buf, 28); // copy the 20-byte info hash starting at byte 28
	// peer id
	buf.write(util.genId(), 28); // write the 20-byte peer id starting at byte 48
	return buf;
};

module.exports.buildKeepAlive = () => Buffer.alloc(4);

module.exports.buildChoke = () => {
	const buf = Buffer.alloc(5);
	// length
	buf.writeUInt32BE(1, 0);
	// id
	buf.writeUInt8(0, 4);
	return buf;
};

module.exports.buildUnChoke = () => {
	const buf = Buffer.alloc(5);
	//length
	buf.writeUInt32BE(1, 0);
	// id
	buf.writeUint8(1, 4);
	return buf;
};

module.exports.buildInterested = () => {
	const buf = Buffer.alloc(5);
	// length
	buf.writeUInt32BE(1, 0);
	// id
	buf.writeUInt8(2, 4);
	return buf;
};

module.exports.buildUninterested = () => {
	const buf = Buffer.alloc(5);
	// length
	buf.writeUInt32BE(1, 0);
	// id
	buf.writeUInt16BE(3, 4);
	return buf;
};

module.exports.buildHave = (payload) => {
	const buf = Buffer.alloc(9);
	// length
	buf.writeUInt32BE(5, 0);
	// id
	buf.writeUInt8(4, 4);
	// piece index
	buf.writeUInt32BE(payload, 5);
	return buf;
};

module.exports.buildBitField = (bitField) => {
	const buf = Buffer.alloc(14);
	// length
	buf.writeUInt32BE(payload.length + 1, 0);
	// id
	buf.writeUInt8(5, 4);
	// bitField
	bitField.copy(buf, 5);
	return buf;
};

module.exports.buildRequest = (payload) => {
	const buf = Buffer.alloc(17);
	// length
	buf.writeUInt32BE(13, 0);
	// id
	buf.writeUInt8(6, 4);
	// piece index
	buf.writeUInt32BE(payload.index, 5);
	// begin
	buf.writeUInt32BE(payload.begin, 9);
	// length
	buf.writeUInt32BE(payload.length, 13);
	return buf;
};

module.exports.buildPiece = (payload) => {
	const buf = Buffer.alloc(payload.block.length + 13);
	// length;
	buf.writeUInt32BE(payload.block.length + 9, 0);
	// id
	buf.writeUInt8(7, 4);
	// piece index
	buf.writeUInt32BE(payload.index, 5);
	// begin
	buf.writeUInt32BE(payload.begin, 9);
	// block
	payload.block.copy(buf, 13);
	return buf;
};

module.exports.buildCancel = (payload) => {
	const buf = Buffer.alloc(17);
	// length
	buf.writeUInt32BE(13, 0);
	// id
	buf.write(UInt8(8, 4));
	// piece index
	buf.writeUInt32BE(payload.index, 5);
	// begin
	buf.writeUInt32BE(payload.begin, 9);
	// length
	buf.writeUInt32BE(payload.length, 13);
	return buf;
};

module.exports.buildPort = (payload) => {
	const buf = Buffer.alloc(7);
	// length
	buf.writeUInt32BE(3, 0);
	// id
	buf.writeUInt8(9, 4);
	// listen port
	buf.writeUInt16BE(payload, 5);
	return buf;
};

module.exports.parse = (msg) => {
	const id = msg.length > 4 ? msg.readInt8(4) : null;
	let payload = msg.length > 5 ? msg.readInt8(5) : null;
	if (id === 6 || id === 7 || id === 8) {
		const rest = payload.slice(8);
		payload = {
			index: payload.readInt32BE(0),
			begin: payload.readInt32BE(4),
		};
		payload[id === 7 ? "block" : "length"] = rest;
	}
	return {
		size: smg.readInt32BE(0),
		id: id,
		payload: payload,
	};
};
