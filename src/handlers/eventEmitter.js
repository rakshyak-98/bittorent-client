const EventEmitter = require("node:events")

class TorrentEmitter extends EventEmitter {}

module.exports = new TorrentEmitter();