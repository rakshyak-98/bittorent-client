Bittorrent solve the issue of the bandwidth. when one host (peer) is broadcasting and rest are watching if they require any piece one can request for that piece. The big issue with the Bittorrent is latency.

>[!NOTE] an initial seeder is necessary to start the process of sharing a file using bittorrent protocol.

## Overview of bittorrent
1. we need to send a request to **tracker**, then tracker will respond with a list of peers.
2. we tell the tracker which file we want to download and tracker will respond with the ip address of the user that can share the file.

>[INFO] Making request to tracker will add your ip address to the list that can share the file.

### What is Bencode?
Bencode is a data serialization format

>[!INFO] When we use data in the networking, it's best to keep them as buffers.

### What does a torrent file contains?
A torrent do not contain actual data but contains _metadata_ about the files to be shared, including:
- Announce URL: The URL of the tracker.
- Info Hash: A unique identifier for the torrent. You get it from creating hash of the `torrent.info`.
- File list: Information about the files being shared.

### How does a client find peers from torrent file?
1. Retrieve Tracker URL
2. Announce to Tracker
3. Receive Peer list

### What is a tracker?
is a special type of server that assists in the communication between peers using the BitTorrent protocol.
- keep track of which peers have which pieces of files.
In a peer-to-peer file sharing, a software client on an end PC requests a file, and portions of the requested file residing on peer machines are sent to the client, and then reassembled into a fully copy of the requested file.

### What is peer?
- Participants in the bittorrent network who shares part of the file with other peers.
1. seeders: which have complete files and share with other.
2. leechers: which are downloading the file and may not share parts they have already downloaded.

### Problems with this approach.
- Initial Seeding: The initial seed must be available to start the sharing process.
- Peer Behavior: If too many users act as leechers and do not upload, the download speed can be affected.

### what is announce-list?
an 'announce-list' is a key in a torrent file that specifies multiple tracker URLs. This allows clients to have a list of trackers they can use to find peers for a given torrent.

>[!INFO] using multiple trackers can improve the availability and reliability of peer discovery.

>[!NOTE] modern BitTorrent clients may implement a distributed hash table and peer exchange protocol to discover peers without trackers; however, trackers are still often included with torrents to improve the speed of peer discovery.

### Why torrent switch from HTTP to UDP protocol and what is the difference
An additional advantage is that a UDP based binary protocol doesn't require a complex parser and no connection handling, reducing the complexity of tracker code and increasing it's performance.
HTTP is built on top of another protocol TCP, which must create a persistance connection before sending data from one user to another use which makes it slower then UDP. In case of UDP the data sent is small enough (like 512 bytes), you don't to worry about receiving only part of the data or out of order. It is possible sometime that the sent data will never reach its destination so data need to be re-send or re-requested. For this UDP is good choice for tracker because they send small messages. And we use TCP when we need to send file between peers because those files tend to larger and need to be intact.

### UDP is unreliable protocol, how to re-transmit lost packets?
The application retransmit the request

### Socket
A socket is an Object through which network communication can happen. In order to send a message through socket it must be buffer not string, or number.
- announce request: is generated to tell the tracker which file intrusted in.
- connect request:

### Why to generate a connect request?
To discover other peers in a swarm a client announces it's existence to a tracker. The HTTP protocol is used and a typical request contains the following parameter:
- info_hash
- key
- peer_id
- port
- downloaded
- left
- uploaded
- compact

>[!INFO Each message is a buffer with a specific format described in the [BEP](http://www.bittorrent.org/beps/bep_0015.html).]

### Connect message
Initial step for initiating.
The BEP describe the connect request as follows:
| Offset | Size           | Name           | Value         |
|--------|----------------|----------------|---------------|
|0       | 64-bit-integer | connection_id  | 0x41727101980 |
|8       | 32-bit-integer | action         | 0 // connect  |
|12      | 32-bit-integer | transaction_id | ? // random   |
|16      |                |                |               |
- The `Ox` indicates that the number is a hexadecimal number, which can be more convenient representation when working with bytes.
- this tells us that our message should start out with 64-bit integer at index 0, and that the value should be 0x417271019180.
Since we just write 8 bytes, the index of the next part is 8. Now wre write 32-bit integer with the value 0. This moves us up to an offset of 12bytes, and we write a random 32-bit integer. So that total message is 8byte+4byte+4byte=16byte long.
- connection id always should be `0x41727101980` when writing the connection request.

### What is the role of connection_id?
It is possible to spoof the source address of a UDP packet. The tracker has to ensure this doesn't occur, so it calculates a value (connection_id) and sends it to the client. If the client spoofed it's source address, it won't receive this value (unless it's sniffing the network). The connection_id will then be send to the tracker again in packet 3. The tracker verifies the connection_id and ignores the request if it doesn't match.
- A connection_id can be used for multiple requests.
- A client can use a connection id until one minute after it has received it.
- Tracker should accept the connection id until two minutes after it has been send.
> [!NOTE] Connection IDs should not be guessable by the client.

### why to split the connection id into two writes in `buildConnReq` function?
The reason we have to write 4byte chunks, is that there is no method to read or write 64-bit integer. Actually node.js [doesn't support precise 64-bit integers](https://stackoverflow.com/questions/307179/what-is-javascripts-highest-integer-value-that-a-number-can-go-to-without-losin).
- `writeUInt32BE` writes an unsigned 32-bit integer in big-endian format.
- `big-endian format` this means the most significant byte (the **big end**) is stored first.

### How does the `readUInt16BE` function?
Internal Steps of `readUInt16BE`:
1. Extract Bytes:
    - The function reads the byte at the specified offset.
    - It then reads the next byte at the offset plus one.
2. Combine Bytes:
    - It shifts the first byte (most significant byte) to the left by 8 bits (one byte).
    - It then performs a bitwise `OR` with the second byte to combine them into a 16-bit integer.

### response message format
Parsing the response is much simpler. Here’s how the response is formatted:
| Offset | Size            | Name            | Value                |
|--------|-----------------|-----------------|----------------------|
| 0      | 32-bit integer  | action          | 0 // connect         |
| 4      | 32-bit integer  | transaction_id  |                      |
| 8      | 64-bit integer  | connection_id   |                      |
| 16     |                 |                 |                      |

### Announce message format
| Offset | Size             | Name         | Value                   |
|--------|------------------|--------------|-------------------------|
| 0      | 64-bit integer   | connection_id|                         |
| 8      | 32-bit integer   | action       | 1 // announce           |
| 12     | 32-bit integer   | transaction_id |                       |
| 16     | 20-byte string   | info_hash    |                         |
| 36     | 20-byte string   | peer_id      |                         |
| 56     | 64-bit integer   | downloaded   |                         |
| 64     | 64-bit integer   | left         |                         |
| 72     | 64-bit integer   | uploaded     |                         |
| 80     | 32-bit integer   | event        | 0 // 0: none; 1: completed; 2: started; 3: stopped |
| 84     | 32-bit integer   | IP address   | 0 // default            |
| 88     | 32-bit integer   | key          | ? // random             |
| 92     | 32-bit integer   | num_want     | -1 // default           |
| 96     | 16-bit integer   | port         | ? // should be betwee  |
| 98     |                  |              |                         |

### Parsing response message format
| Offset     | Size             | Name           | Value                             |
|------------|------------------|----------------|-----------------------------------|
| 0          | 32-bit integer   | action         | 1 // announce                     |
| 4          | 32-bit integer   | transaction_id |                                   |
| 8          | 32-bit integer   | interval       |                                   |
| 12         | 32-bit integer   | leechers       |                                   |
| 16         | 32-bit integer   | seeders        |                                   |
| 20 + 6 * n | 32-bit integer   | IP address     |                                   |
| 24 + 6 * n | 16-bit integer   | TCP port       |                                   |
| 20 + 6 * N |                  |                |                                   |

It's a bit tricky because the number of addresses that come back isn't fixed. The addresses come in group of 6 bytes, the first 4 represent the IP address and the next 2 represent the port. So our code will need to correctly break up the address part of the response.

### Why used SHA1 hashing function?
SHA1 is the one used by bittorrent so in our case no other hashing function will do.
**why use hash ?** it is a compact way to uniqely identify the torrent. A hashing function return a fixed length buffer. And the info property contains information about every piece of the torrent's files, it's a good way to uniquely identify a torrent.

> [!INFO] `torrent.info.pieces` is a buffer that contains 20-byte SHA-1 hash of each piece, and the length gives you the total number of bytes in the buffer.

## Why `bignum` package is used?
There's one problem to consider with the file, size, which is that it may be larger than a 32-bit integer. The easiest way to deal with this is to install a module to handle larger number. The option `{size: 8}` tells the function you want to write the number to a buffer of size 8 bytes. This is also the buffer size required by the [announce request](https://allenkim67.github.io/programming/2016/05/04/how-to-make-your-own-bittorrent-client.html#announce-messaging)

## Handle UDP messages dropped in transit
retry function should wait for 2^n*15 seconds (ideal time) between each request up to 8 requests total according to the [BEP](http://www.bittorrent.org/beps/bep_0015.html).
>[!NOTE] it is necessary to re-request a connection ID when re-transmission has expired.

# Downloading from peers
- first create a tcp connection with all the peers in your list. (The more peers you can get connected to the faster you can download your files.)
- After exchanging some messages with the peers as setup,
    - start requesting pieces of the files you want.
>[!INFO] A torrent's shared files are broken up into pieces so that you can download different parts of the files from different peers simultaneously.
- When we're done receiving a piece from a peer request the next piece.
> [!NOTE] Ideally you want all the connections to be requesting different and new pieces so you'll need to keep track of which pieces you already have and which onces you still need.
- finally, when you receive the pieces they'll be store in memory so you'll need to write the data to your hard disk.

## Protocol overview
Once a tcp connection is established the messages you send and receive have to follow the following protocol.
- first let your peer know which files you are interested in downloading from them, as well as some identifying info.
- next peer will let you know what pieces they have. this happens through the *have* and *bitfield* message contains a piece index as its payload. this means you will receive multiple have messages one for each piece that your peer has.

> [!INFO] the `bitfield` message serves a similar purpose, but does it in a different way the `bitfield` message can tell you all the pieces that the peer has in just one message. it does this by sending a string of bits, one for each piece in the file.

> [!NOTE] The index of each bit is the same as the pieces index, and if they have the pieces it will be set to 1, if not it will be set to 0. for example if you receive a bitfield that starts with 011001... that means they have the pieces at index 1,2, and 5 but not the pieces at index 0,3 and 4.

it's possible to receive both *have* messages and a bitfield message, if which case you should combine them to get the full list of pieces.

### What is a **handshake**
If the peer doesn't have the files you want they will close the connection, but if they do have the files they should send back a similar message as confirmation. This is called the *handshake*.

### message spec
Once the handshake has been established there are 10 different types of message that can be exchanged, [specs](https://wiki.theory.org/BitTorrentSpecification#Messages).

### Pieces
After you establish the handshake your peers should tell you which pieces they have.
If you open up a torrent file, we saw that it contains data with various properties like the *announce* and *info* properties. Another property is *piece length* property. This tells you how long a piece is in bytes. Let's say hypothetically that you have a piece length of 1000 bytes. Then if the total size of the file's is 12000bytes, that means the file should have 12 pieces. Note that the last piece might not be the full 1000bytes. If the file were 120001bytes larger, then it would be a total of 13 pieces, where the last piece is just 1 byte large.
- these pieces are indexed starting at 0, this is how we know which piece it is that we are sending or receiving. For example, you might request the piece at index 0, that means from our previous example we want the first 1000 bytes of the file. If we ask for the piece at index 1, we want the second 1000bytes and so on.

>[!INFO] piece length can be found in torrent file, and block length is 2^14 bytes by convention.

## Job queue
This list will contain the pieces that a single peer has. Why do we have to maintain this list? The problem is that we would probably end udp requesting all the pieces from the very first peer we connect to and then since we don't want to double request the same piece, none of the other peers would have pieces left to request.

- some peers will inevitably upload faster than others.
- Ideally we want to fastest peers to get more requests, rather than have multiple requests bottle-necked by the slowest peer.

A natural solution is to request just one or a few pieces from a peer at a time, and only make the next request after receiving a response. This way that faster peers will send their responses faster, *coming back* for more requests more frequently.

### Request Failures
it's possible for us to request a piece but never receive it. This is because a connection can drop at any time for whatever reason. Since we avoid requesting pieces that have been added to the `requested` array, these pieces will never be received.
What solves is maintain two lists, one for requested pieces and one for received pieces.
- update the `requested` list at request time
- update the `received` list at receive time.
- Then whenever we have requested all pieces but there are still pieces that we haven't received, we copy the `received` list into the `requested` list, and that will allow us to re-request those missing pieces.

### Other kind of messages Choke, Un-choke, interested, not interested?
it's possible to receive another kind of message, the peer might decide they don't want to share with you! That's what the *choke*, *un-choke*, *interested*, and *not interested* message are for.
- choked means the peer does not want to share with you.
- un-choked peer is willing to share.
- interested means you want what your peer has.
- not interested means you don't want what they have.
> [!INFO] a request start with choked and not interested. So the first message you send should be the interested message. Then hopefully the peers will send you an un-choke message and you can move to the next step. If you receive a choke message instead you can just let the connection drop.

### approach taken for implementation of Choke and Un-choke
> [!NOTE] we don't want to request any pieces until we've been un-choked.
The simplest way to enforce this to create a new object that hold both our `queue` array as well as `choked` property.

### why dividing `torrent.info.pieces` by 20?
the `torrent.info.pieces` is a buffer that contain 20-byte SHA-1 hash of each piece, and the length gives you the total number of bytes in the buffer. That's why we divide by 20 to get the total number of pieces.

### the `requested` and `received` arrays
holds an array of arrays, where the inner arrays hold the status of a block at a given piece index. So if you wanted to find out the status of a block at index 1 for a piece at index 7, you could look up `_requested[7][1]` and check it's set to true. As before all values are initially set to false.

### the `begin`and `length` property. Why are present?
These two properties are necessary because sometimes a pieces is oo big for one message. Although there is some dispute about the best size, it is typically considered to be around 2^14 (16384)bytes. This is called a *block*, where a piece consists of one or more blocks. If the piece length is greater than the length of single block, then it should be broken up into blocks with one message sent for each block.
So the "begin" field is the zero-based byte offset starting from the beginning of the piece, and the "length" of the requested block. This is always to be 2^14, except possibly the last block which might be less.

### what is a payload?
Payload is a long string of bits.

### `pieceHandler` function
`pieceHandler` which is for when we receive the actual bytes of the piece we requested.

