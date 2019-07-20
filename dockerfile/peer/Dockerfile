FROM ubuntu:xenial

MAINTAINER yoseplee

# update repository source list
RUN DEBIAN_FRONTEND=noninteractive && apt-get -qq update && apt-get -y -qq dist-upgrade

# install openjdk and other programs
RUN DEBIAN_FRONTEND=noninteractive && apt-get -qq install -y vim screen wget git curl openjdk-8-jdk

# copy necessary files
ADD init.sh /peer/
ADD Server.java /peer/playground/rmi/
ADD Hello.java /peer/playground/rmi/
ADD Client.java /peer/playground/rmi/

#ADD /Users/yosep/Documents/github/peer-network/playground/rmi/Hello.java /playground/rmi/

WORKDIR /peer/playground/rmi/
RUN pwd

# compile server program
RUN javac Server.java Hello.java Client.java

WORKDIR /peer
# run rmiregistry and server program
ENTRYPOINT [ "/peer/init.sh" ]

EXPOSE 1099