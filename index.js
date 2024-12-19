"use strict";

const download = require("./src/controller/download");
const torrentParser = require("./src/models/torrentParser");

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
