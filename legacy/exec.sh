#!/bin/bash

which=$1
mainClass=""

if [ x$1 == x ]; then
	echo "invalid argumeht: see guide below"
	echo "./exec.sh [argument]"
	echo "peer | run plum peer"
	echo "client | run plum client"
	echo "conductor | run plum conductor"
	echo "cli | run plum cli"
	exit
fi

if [ ${which} == "peer" ]; then
	mainClass="plum.PlumPeer"
elif [ ${which} == "client" ]; then
	mainClass="plum.PlumClient"
elif [ ${which} == "conductor" ]; then
	mainClass="plum.PlumConductor"
elif [ ${which} == "cli" ]; then
	mainClass="plum.cli.PlumCli"
fi

echo "exec via mvn... mainClass=${mainClass}"

mvn exec:java -Dexec.mainClass=${mainClass}
