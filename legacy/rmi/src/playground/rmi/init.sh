#!/bin/bash
echo "init peer"
rmiregistry &
java playground.rmi.Server