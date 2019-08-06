#!/bin/bash

which=$1
mainClass=""

if [ x$1 == x ]; then
	echo "invalid argumeht: see guide below"
	echo "./exec.sh [argument]"
	echo "peer | run plum peer"
	echo "client | run plum client"
	exit
fi

if [ ${which} == "peer" ]; then
	mainClass="plum.PlumPeer"
elif [ ${which} == "client" ]; then
	mainClass="plum.PlumClient"
fi

echo "exec via mvn... mainClass=${mainClass}"

mvn exec:java -Dexec.mainClass=${mainClass}
