#!/bin/bash
rm -f /tmp/registerClient
rm -f /tmp/start-worker-hadoop.sh
service ssh start
ip=`ifconfig eth0 | grep 'inet addr' | cut -d : -f 2 | cut -d ' ' -f 1`
sed -i "s/hadoop-master/$ip/" $HADOOP_HOME/etc/hadoop/core-site.xml
sed -i "s/hadoop-master/$ip/" $HADOOP_HOME/etc/hadoop/yarn-site.xml

$HADOOP_HOME/sbin/start-dfs.sh &

$HADOOP_HOME/sbin/start-yarn.sh &

/tmp/registerServer &

/bin/upload/upload &

/bin/gotty --port 8000 --permit-write --reconnect /bin/bash  


