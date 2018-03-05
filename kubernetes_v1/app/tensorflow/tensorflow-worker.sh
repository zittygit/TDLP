#! /bin/sh

/usr/sbin/nslcd

cd /tensorflow
client $1 worker
