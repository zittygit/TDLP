#!/bin/bash

cd /opt
/usr/sbin/nslcd
export TERM=xterm
export ZEPPELIN_HOME=/opt/zeppelin
export ZEPPELIN_CONF_DIR=$ZEPPELIN_HOME/conf

sed -i "s/SPARK_MASTER/$1/" /opt/spark/conf/spark-defaults.conf
sed -i "s/SPARK_MASTER/$1/" /opt/zeppelin/conf/zeppelin-env.sh
/opt/zeppelin/bin/zeppelin-daemon.sh --config $ZEPPELIN_CONF_DIR start
gotty --port 8000 --permit-write --reconnect /bin/bash
