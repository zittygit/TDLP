#!/bin/bash
rm -f /tmp/start-master-hadoop.sh
rm -f /tmp/registerServer
service ssh start

sed -i "s/hadoop-master/$1/" $HADOOP_HOME/etc/hadoop/core-site.xml
sed -i "s/hadoop-master/$1/" $HADOOP_HOME/etc/hadoop/yarn-site.xml

/usr/local/hadoop/sbin/hadoop-daemon.sh start datanode & 
/usr/local/hadoop/sbin/yarn-daemon.sh start nodemanager &

/tmp/registerClient $1
rm -f /tmp/registerClient
 
tail -f /dev/null


