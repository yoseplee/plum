FROM ubuntu:xenial

MAINTAINER yoseplee

# update repository source list
RUN DEBIAN_FRONTEND=noninteractive && apt-get -qq update && apt-get -y -qq dist-upgrade

# install openjdk and other programs
RUN DEBIAN_FRONTEND=noninteractive && apt-get -qq install -y vim net-tools screen wget git curl openjdk-8-jdk

# copy script file to run jar file
ADD init.sh /plum/
# copy PlumPeer jar file
ADD plum-network-1.0-SNAPSHOT.jar /plum/

# run PlumPeer
WORKDIR /plum
ENTRYPOINT [ "/plum/init.sh" ]

EXPOSE 50051