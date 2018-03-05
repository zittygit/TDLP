#!/bin/sh

openssl genrsa -out kubernetes.rsa 4096
openssl rsa -in kubernetes.rsa -pubout > kubernetes.rsa.pub
