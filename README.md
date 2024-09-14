# Seedstore

#### Inspired by [Queue4Download](https://github.com/weaselBuddha/Queue4Download)

## Overview
Take a look at the inspiration above to learn more about how this is setup
## Setup
You need to have a `config.json` located in `$HOME/.seedstore`
and here is the schema for it:
```json5
{
  "mqtt":{
    "user": "freerealesstate", // MQTT user account name
    "password": "foobar", // MQTT user account password 
    "host": "192.168.3.2" // MQTT server for events
  },
  "torrent": {
    "client": "RTORRENT", // Unused property right now
    "source": "192.168.1.2", // seedbox server address
    "labelling":false // Unused property right now
  },
  "server":{
    "defaultCode":"V", // If the processing of rules fails, this is the default code that is assigned
    "codeConditions":[
      {
        "value":"loremipsum",
        "operator":"in",
        "entity":"name",
        "code":"A"
      }
    ]
  },
  "client": {
    "codeDestinations":{
      "A":"/local/path/on/client/for/code",
      // make sure your have the default code specified here
    },
    "lftp": {
      "threads": 5, // the amount of threads to use on LFTP transfer
      "segments": 4 // the amount of segments to use when mirroring directories on LFTP transfer
    },
    "credentials": {
      "username":"iliketurtles", // the username for the seedbox
      "password":"batquot" // the password for the seedbox
    }
  }
}
```