FROM  192.168.31.85:5523/ububut-jdk8:14.04
MAINTAINER zhangzy65@mail2.sysu.edu.cn
WORKDIR /root

# install hadoop 2.7.2
COPY hadoop-2.7.2.tar.gz /root 
RUN tar -xzvf hadoop-2.7.2.tar.gz && \
    mv hadoop-2.7.2 /usr/local/hadoop && \
    rm hadoop-2.7.2.tar.gz

# set environment variable
ENV JAVA_HOME=/usr/lib/jvm/java-8-openjdk-amd64 
ENV HADOOP_HOME=/usr/local/hadoop 
ENV PATH=$PATH:/usr/local/hadoop/bin:/usr/local/hadoop/sbin 

# ssh without key
RUN ssh-keygen -t rsa -f ~/.ssh/id_rsa -P '' && \
    cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys

RUN mkdir -p ~/hdfs/namenode && \ 
    mkdir -p ~/hdfs/datanode && \
    mkdir $HADOOP_HOME/logs  && \
	mkdir /bin/upload

COPY config/* /tmp/
COPY gotty /bin/
COPY upload/* /bin/upload/

RUN mv /tmp/ssh_config ~/.ssh/config && \
    mv /tmp/hadoop-env.sh /usr/local/hadoop/etc/hadoop/hadoop-env.sh && \
    mv /tmp/hdfs-site.xml $HADOOP_HOME/etc/hadoop/hdfs-site.xml && \ 
    mv /tmp/core-site.xml $HADOOP_HOME/etc/hadoop/core-site.xml && \
    mv /tmp/mapred-site.xml $HADOOP_HOME/etc/hadoop/mapred-site.xml && \
    mv /tmp/yarn-site.xml $HADOOP_HOME/etc/hadoop/yarn-site.xml && \
    mv /tmp/slaves $HADOOP_HOME/etc/hadoop/slaves && \
    mv /tmp/run-wordcount.sh ~/run-wordcount.sh

RUN chmod +x /tmp/start-master-hadoop.sh && \
    chmod +x /tmp/start-worker-hadoop.sh && \
    chmod +x /tmp/registerServer &&\
    chmod +x /tmp/registerClient &&\
    chmod +x ~/run-wordcount.sh && \
    chmod +x $HADOOP_HOME/sbin/start-dfs.sh && \
    chmod +x $HADOOP_HOME/sbin/start-yarn.sh && \
    chmod +x /bin/gotty

# format namenode
RUN /usr/local/hadoop/bin/hdfs namenode -format


