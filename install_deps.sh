#!/bin/bash

sudo docker image pull postgres:latest 
sudo docker image pull adminer:latest

go get github.com/bmizerany/pq
go get github.com/pkg/errors
go get golang.org/x/net/websocket

