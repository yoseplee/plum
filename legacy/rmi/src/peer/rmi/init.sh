#!/bin/bash
echo "init peer"
rmiregistry &
echo "sleep for 3 seconds"
sleep 3s #waits 3 seconds
java peer.rmi.Peer