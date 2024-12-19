const fs = require("fs");
const net = require("node:net");
const Buffer = require("node:buffer").Buffer;
const tracker = require("../services/tracker");
const Pieces = require("../models/pieces").default;
const message = require("../message");
const Queue = require("../models/Queue");
const eventEmitter = require("../handlers/eventEmitter")
const handler = require("../handlers/message-handler")

module.exports = (torrent, path) => {
	tracker.getPeers(torrent, (peers) => {
		const pieces = new Pieces(torrent);
		const file = fs.openSync(path, "w");
		peers.forEach((peer) => download(peer, torrent, pieces, file));
	});
};

function download(peer, torrent, pieces, file) {
	/* 
		logic for downloading the file from the peer.
	 */
	const socket = new net.Socket();
	socket.on("error", console.error);
	socket.connect(peer.port, peer.id, () => {
		socket.write(message.buildHandshake(torrent));
	});
	const queue = new Queue(torrent);
	onWholeMsg(socket, (msg) => msgHandler(msg, socket, pieces, queue, file));
}

function onWholeMsg(socket, callback) {
	let saveBuf = Buffer.alloc(0);
	let handshake = true;
	socket.on("data", (recvBuf) => {
		const msgLen = () =>
			handshake ? saveBuf.readUInt8(0) + 49 : saveBuf.readInt32BE(0) + 4;
		saveBuf = Buffer.concat([saveBuf, recvBuf]);
		while (saveBuf.length >= 4 && saveBuf.length >= msgLen()) {
			callback(saveBuf.subarray(0, msgLen()));
			saveBuf = saveBuf.subarray(msgLen());
			handshake = false;
		}
	});
}

function isHandshake(msg) {
	return (
		msg.length === msg.readUInt8(0) + 49 &&
		msg.toString("utf-8", 1) === "BitTorrent protocol"
	);
}

/**
 *
 * @param {Buffer} payload
 * @param {net.Socket} socket
 * @param {Pieces} requested
 * @param {Queue} queue
 */

function haveHandler(payload, socket, requested, queue) {
	// if the peer has the index-0 piece, first bit is 1. if not it will be 0;
	const pieceIndex = payload.readUInt32BE(0);
	const queueEmpty = queue.length === 0;
	queue.push(pieceIndex);
	if (queueEmpty) requestPiece(socket, requested, queue);
}

function chokeHandler(socket, pieces) {
	socket.end();
}

/**
 * @param {net.Socket} socket;
 * @param {Pieces} pieces;
 * @param {Queue} queue;
 */
function unChokeHandler(socket, pieces, queue) {
	queue.choked = false;
	requestPiece(socket, pieces, queue);
}

/**
 * @param {net.Socket} socket
 * @param {Pieces} piece;
 * @param {Queue} queue;
 */
function requestPiece(socket, piece, queue) {
	if (queue.choked) return null;
	while (queue.queue.length) {
		const pieceIndex = queue.shift();
		if (piece.needed(pieceIndex)) {
			socket.write(message.buildRequest(pieceIndex));
			piece.addRequest(pieceIndex);
			break;
		}
	}
}

/**
 * @param {net.Socket} socket;
 * @param {Pieces} pieces;
 * @param {Queue} queue;
 * @param {Buffer} payload;
 */
function bitFieldHandler(socket, pieces, queue, payload) {
	const queueEmpty = queue.length === 0;
	payload.forEach((byte, i) => {
		for (let j = 0; j < 8; j++) {
			if (byte % 2) queue.queue(i * 8 + 7 - j);
			// dividing by 2 and taking the remainder will convert a base-10 to binary RHS to LSH.
			byte = Math.floor(byte / 2);
		}
	});
	if (queueEmpty) requestPiece(socket, pieces, queue);
}

/**
 *
 * @param {Buffer} payload
 * @param {net.Socket} socket
 * @param {Pieces} pieces
 * @param {Queue} queue
 * @param {torrent} torrent
 * @param {String} file
 * @returns
 */
function pieceHandler(
	payload,
	socket,
	pieces,
	queue,
	torrent,
	file,
	pieceResp
) {
	pieces.printPercentDone();
	pieces.addReceived(pieceResp);
	const offset =
		pieceResp.index * torrent.info["piece length"] + pieceResp.begin;
	fs.write(file, pieceResp.block, 0, pieceResp.block.length, offset, () => {});
	if (pieces.isDone()) {
		console.log("Connection end");
		socket.end();
		try {
			fs.closeSync();
		} catch (e) {}
	} else {
		requestPiece(socket, pieces, queue);
	}
}

function requestPiece(socket, pieces, queue) {
	if (queue.choked) return null;

	while (queue.length()) {
		const pieceBlock = queue.deque();
		if (pieces.needed(pieceBlock)) {
			socket.write(message.buildRequest(pieceBlock));
			pieces.addRequest(pieceBlock);
			break;
		}
	}
}

function msgHandler(msg, socket, pieces, queue, file) {
	if (isHandshake(msg)) {
		socket.write(message.buildInterested());
	} else {
		const m = message.parse(msg);
		switch(m.id){
		case 0:
			eventEmitter.emit('choke', socket, pieces, queue);
			break;
		case 1:
			eventEmitter.emit('unchoke', socket);
			break;
		case 4:
			eventEmitter.emit('have', m.payload, socket, pieces, queue);
			break;
		case 5:
			eventEmitter.emit('bitfield', socket, pieces, queue, m.payload);	
			break;
		case 7:
			eventEmitter.emit('piece', m.payload, socket, pieces, queue, torrent, file);	
			break;
		}
	}
}


// Register event handlers
eventEmitter.on('handshake', (socket, torrent) => new handler.HandShakeHandler().handle(socket, torrent));
eventEmitter.on('interested', (socket) => new handlers.InterestedHandler().handle(socket));
eventEmitter.on('choke', (socket, pieces, queue) => new handlers.ChokeHandler().handle(socket, pieces, queue));
eventEmitter.on('unchoke', (socket) => new handlers.UnChokeHandler().handle(socket));
eventEmitter.on('have', (payload, socket, pieces, queue) => new handlers.HaveHandler().handle(payload, socket, pieces, queue));
eventEmitter.on('bitfield', (socket, pieces, queue, payload) => new handlers.BitFieldHandler().handle(socket, pieces, queue, payload));
eventEmitter.on('piece', (payload, socket, pieces, queue, torrent, file) => new handlers.PieceHandler().handle(payload, socket, pieces, queue, torrent, file));