FROM ubuntu:xenial

MAINTAINER yoseplee

RUN DEBIAN_FRONTEND=noninteractive && apt-get -qq update && apt-get -y -qq dist-upgrade

RUN DEBIAN_FRONTEND=noninteractive && apt-get -qq install -y vim screen wget git curl openjdk-8-jdk

# ENV JAVA_HOME /usr/lib/jvm/java-8-openjdk-amd64