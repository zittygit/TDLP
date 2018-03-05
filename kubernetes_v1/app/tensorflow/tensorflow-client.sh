#! /bin/sh

/usr/sbin/nslcd
export HOME=/tensorflow
export TERM=xterm

tensorboard --port 8888 --logdir /tensorflow/logs &
jupyter notebook --notebook-dir /notebooks --config /config.json &

cd /tensorflow
gotty --port 8000 --permit-write --reconnect /bin/bash &
server $1 $2 
