"use strict";

const download = require("./src/download");
const torrentParser = require("./src/torrent-parser");

class Download{
  path
  constructor(filePath){
    this.path = filePath;
  }

  download(){
    download(torrentParser.open(filepath))
  }
}

module.exports =  new Download();
