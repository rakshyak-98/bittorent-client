"use strict";
const crypto = require("crypto");
let id = null;

/**
 * Generate peer_id that can be any random 20 byte value got to http://www.bittorrent.org/beps/bep_0020.html
 * @returns
 */
module.exports.genId = () => {
	if (!id) {
		id = crypto.randomBytes(20);
		// ID abbreviation 0001 is version number
		Buffer.from("-ID0001-").copy(id, 0);
	}
	return id;
};

