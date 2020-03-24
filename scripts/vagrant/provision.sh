#!/usr/bin/env bash

add-apt-repository ppa:longsleep/golang-backports
apt update -y
apt install -y zfsutils-linux golang-go
