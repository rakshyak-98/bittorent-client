const message = require('../message');

class HandShakeHandler {
    handle(socket, torrent) {
        socket.write(message.buildHandshake(torrent))
    }
}

class InterestedHandler {
    handle(socket) {
        socket.write(message.buildInterested())
    }
}

class ChokeHandler {
    handle(socket) {
        socket.write(message.buildChoke())
    }
}

class UnChokeHandler {
    handle(socket) {
        socket.write(message.buildUnchoke())
    }
}

class HaveHandler {
    handle(socket, payload) {
        socket.write(message.buildHave(payload))
    }
}

class BitFieldHandler {
    handle(socket, payload) {
        socket.write(message.buildBitfield(payload))
    }
}

class PickPieceHandler {
    handle(socket, pieces, queue) {
        if (queue.choked) return null;
        while (queue.length()) {
            const pieceBlock = queue.deque();
            if (pieces.needed(pieceBlock)) {
                socket.write(message.buildRequest(pieceBlock));
                pieces.addRequested(pieceBlock);
                break;
            }
        }
    }
}

module.exports = {
    HandShakeHandler,
    InterestedHandler,
    ChokeHandler,
    UnChokeHandler,
    HaveHandler,
    BitFieldHandler,
    PickPieceHandler
}