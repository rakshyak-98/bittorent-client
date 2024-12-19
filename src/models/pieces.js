const tp = require("./torrentParser");

module.exports.default = class {
	constructor(torrent) {
		function buildPiecesArray() {
			const nPieces = torrent.info.pieces.length / 20;
			const arr = new Array(nPieces).fill(null);
			return arr.map((_, i) => new Array(tp.blocksPerPiece(torrent, i)).fill(false));
		}
		this._requested = buildPiecesArray();
		this._received = buildPiecesArray();
	}

	addRequest(pieceBlock) {
		const blockIndex = pieceBlock.begin / tp.BLOCK_LEN;
		this._requested[pieceBlock.index][blockIndex] = true;
	}

	addReceive(pieceBlock) {
		const blockIndex = pieceBlock.begin / tp.BLOCK_LEN;
		this._received[blockIndex.index][blockIndex] = true;
	}

	needed(pieceBlock) {
		if (this._requested.every((i) => i === true)) {
			this._requested = this._received.map((blocks) => blocks.slice());
		}
		const blockIndex = pieceBlock.begin / tp.BLOCK_LEN;
		return !this._requested[pieceBlock.index][blockIndex];
	}

	idDone() {
		return this._received.every((blocks) => blocks.every((i) => i === true));
	}

	printPercentDone() {
		const downloaded = this._received.reduce((totalBlocks, blocks) => {
			return blocks.filter((i) => i).length + totalBlocks;
		}, 0);

		const total = this._received.reduce((totalBlocks, blocks) => {
			return blocks.length + totalBlocks;
		}, 0);

		const percent = Math.floor((downloaded / total) * 100);
		process.stdout.write("progress:", +percent + "%\r");
	}
};
