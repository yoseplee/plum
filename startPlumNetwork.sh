#!/bin/bash

function startNetwork() {
  docker-compose -f compose/"${PEERS}"-peers.yml up &
}

function stopNetwork() {
  docker-compose -f compose/"${PEERS}"-peers.yml down
}

#parse commands
if [[ $# -lt 1 ]] ; then
  echo "no commands"
  exit 0
else
  CMD=$1
fi

# parse
if [[ $# -ge 2 ]] ; then
  PEERS=$2
else
  echo "default number of peers is set: 4"
  PEERS=4
fi

if [ "$CMD" == "up" ]; then
  echo "start the plum network for ${PEERS}-peers.yml"
  echo
elif [ "$CMD" == "down" ]; then
  echo "stop the plum network for ${PEERS}-peers.yml"
  echo
fi

if [ "$CMD" == "up" ]; then
  startNetwork
  echo
elif [ "$CMD" == "down" ]; then
  stopNetwork
  echo
fi