FROM ubuntu:xenial

MAINTAINER yoseplee

# update repository source list
RUN DEBIAN_FRONTEND=noninteractive && apt-get -qq update && apt-get -y -qq dist-upgrade

# install openjdk and other programs
RUN DEBIAN_FRONTEND=noninteractive && apt-get -qq install -y vim net-tools screen wget git curl openjdk-8-jdk

# copy necessary files
ADD init.sh /plum/
ADD Peer.java /plum/peer/rmi/
ADD MessageIF.java /plum/peer/rmi/
ADD Client.java /plum/peer/rmi/

# compile Peer program
WORKDIR /plum/peer/rmi/
RUN javac *.java

# run rmiregistry and Peer program
WORKDIR /plum
ENTRYPOINT [ "/plum/init.sh" ]

EXPOSE 1099