#!/bin/bash
export JAVA_HOME=/usr/lib/jdk/jdk-10.0.2
export PATH=$PATH:$JAVA_HOME/bin

echo "init Plum Peer!"
java -jar plum-network-1.0-SNAPSHOT.jar

echo "sleep for 3 seconds for stabilize"
sleep 3s