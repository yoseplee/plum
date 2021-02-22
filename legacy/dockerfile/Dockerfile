FROM ubuntu:xenial

MAINTAINER yoseplee

# update repository source list
# RUN DEBIAN_FRONTEND=noninteractive && apt-get -qq update && apt-get -y -qq dist-upgrade
RUN apt-get -qq update && apt-get -y -qq dist-upgrade

# install essential tools
# RUN DEBIAN_FRONTEND=noninteractive && apt-get -qq install -y vim net-tools screen wget git curl
RUN apt-get -qq install -y vim net-tools screen wget git curl

# install openjdk
RUN wget https://download.java.net/java/GA/jdk10/10.0.2/19aef61b38124481863b1413dce1855f/13/openjdk-10.0.2_linux-x64_bin.tar.gz
RUN tar -xvf openjdk-10.0.2_linux-x64_bin.tar.gz
RUN mkdir -p /usr/lib/jdk
RUN mv jdk-10.0.2 /usr/lib/jdk

# copy script file to run jar file
ADD init.sh /plum/
# copy PlumPeer jar file
ADD plum-network-1.0-SNAPSHOT.jar /plum/

# run PlumPeer
WORKDIR /plum
ENTRYPOINT [ "/plum/init.sh" ]

EXPOSE 50051