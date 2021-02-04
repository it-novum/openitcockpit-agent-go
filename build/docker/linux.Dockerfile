FROM debian:latest

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get -y install ruby ruby-dev rubygems build-essential git rpm libarchive-tools tar && rm -rf /var/lib/apt/lists/*

RUN gem install --no-document fpm
