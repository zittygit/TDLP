#!/bin/bash
#docker 私有镜像库的安装配置脚本
#environment centos7
#param IP port  registry
######example##########
#docker-registry.sh 192.168.88.155 5050 /opt/registry

set -e #出现异常则终止退出

if [ $(id -u) -ne 0 ] ;then
    echo must run as root
    exit 1
fi
if [ $# != 3 ];then
    echo must input [IP] [port] [registry]
	exit 1
fi

if [ -z "`ps -e|grep docker`" ] ;then
    echo install docker...
	yum install -y docker
fi

echo open docker insercure registry option
if [ -z "`cat /etc/sysconfig/docker|grep insecure`" ] ;then
   sed -i -e "s/OPTIONS='/OPTIONS='--add-registry=$1:$2 --insecure-registry=$1:$2 /g" /etc/sysconfig/docker
   echo restart docker
   systemctl restart docker
fi

echo install private docker registry...
docker run --name=registry  --privileged=true -d -p $2:5000 -v $3:/var/lib/registry registry.docker-cn.com/library/registry

if [ -z "`docker ps|grep registry`" ] ;then
   echo install failed
   echo progarm exit!
fi

docker ps|grep registry

echo docker registry install complete!