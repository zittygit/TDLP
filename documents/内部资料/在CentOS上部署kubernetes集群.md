### 部署环境
  - OS环境：centos 7
  - kubernetes版本：v1.7.2
  - docker版本
    - version:1.12.6
    - go version: go1.6.4
    - API version: 1.24
##### **注意：docker 版本不能用最新版本，否则不兼容1.7的k8s.**

  - K8S的安装包
 
     ![image](http://img.blog.csdn.net/20171213152813702?watermark/2/text/aHR0cDovL2Jsb2cuY3Nkbi5uZXQvejI5NDE1NTY3Mw==/font/5a6L5L2T/fontsize/400/fill/I0JBQkFCMA==/dissolve/70/gravity/SouthEast)

  - K8S组件docker镜像
  
    ![image](http://img.blog.csdn.net/20171213162004437?watermark/2/text/aHR0cDovL2Jsb2cuY3Nkbi5uZXQvejI5NDE1NTY3Mw==/font/5a6L5L2T/fontsize/400/fill/I0JBQkFCMA==/dissolve/70/gravity/SouthEast)

 - 服务环境安装
   - 服务磁盘阵列使用RAID 5模式，既保证了IO性能，同时满足数据可恢复。
   - 配置服务器RAID
      - 在BIOS自检时，根据提示按crtl+i进入RAID管理界面。
      - 创建VD。
      - 选择RAID模式，这里为了提高硬盘的IO性能使用的是RAID-5模式。
      - 将srcipt size设置为最大值。
      - 将所有硬盘添加到VD中，退出，进入centos7安装中就可以识别硬盘了。
### kubernetes 部署
- kubeadm 介绍

Kubernetes 是 Google 开源的基于 Docker 的容器集群管理系统，通过 yaml 语言写的配置文件，简单快速的就能自动部署好应用环境，支持应用横向扩展，并且可以组织、编排、管理和迁移这些容器化的应用。Kubeadm 是一个可以快速帮助我们创建稳定集群服务的工具，通过它，我们可以在虚拟机、实体机或者云端快速部署一个高可用的集群服务。

- 集群服务器信息

HostName | IP|CPU|MEM|GPU
---|---|---|---|---
k8s-service | 192.168.1.200|6 core|64G|0
k8s-master | 192.168.1.201|6 core|64G|0
k8s-node1 | 192.168.1.101|54 core|166G|0
k8s-node2 | 192.168.1.102|54 core|166G|0
k8s-node3 | 192.168.1.103|6 core|64G|0
k8s-node4 | 192.168.1.104|6 core|64G|0

- 安装Docker
   
Docker 本地安装，已经打包成压缩文件，需要解压后本地安装。
```
tar zxf /tmp/docker.tar.gz -C /tmp
yum localinstall -y /tmp/docker/*.rpm
```
Docker配置

```
setenforce 0
sed -i -e 's/SELINUX=enforcing/SELINUX=disabled/g' /etc/selinux/config
#关闭防火墙
systemctl disable firewalld
systemctl stop firewalld
echo DOCKER_STORAGE_OPTIONS=\" -s overlay --selinux-enabled=false\" > /etc/sysconfig/docker-storage
systemctl daemon-reload && systemctl restart docker.service
```
查看Docker 版本

```
docker version
```
![image](C:/Users/ziye/Desktop/1.png)

- 安装Docker私有库
私有库可以使用最新版本的Docker,安装教程Docker官网

```
$ sudo yum-config-manager \
    --add-repo \
    https://download.docker.com/linux/centos/docker-ce.repo
$ sudo yum install docker-ce
# 启动docekr
$ sudo systemctl start docker
```
![image](C:/Users/ziye/Desktop/2.png)

Docker 私有库配置
启动docker私有库镜像服务

```
docker run --name=registry --restart=unless-stopped --privileged=true -d -p 5523:5000 -v /opt/registry:/var/lib/registry registry.docker-cn.com/library/registry
```
命令解读：--name 服务名，--restart 重启策略，--privileged特权，-p端口映射，-v挂载
![image](C:/Users/ziye/Desktop/5.png)

为了方便配置，私有库Docker采用http模式访问，故需要配置insecure-registry参数
在 /etc/sysconfig/docker 文件的 OPTIONS参数中添加–-insecure-registry=192.168.1.200:5523
![image](C:/Users/ziye/Desktop/3.png)
K8S集群和Docker 私有库Docker版本不同，需要修改的文件不同，在K8S集群Docker需要修改/usr/lib/systemd/system/docker.service文件配置insecure-registry参数，命令入下：

```
$ sudo sed -i -e 's/dockerd/dockerd --insecure-registry=192.168.1.200:5523/g' /usr/lib/systemd/system/docker.service
```
- kubernetes 安装

master 安装
k8s版本V1.7.2
```
$ sudo tar zxf /tmp/k8s.tar.gz -C /tmp
$ sudo yum localinstall -y  /tmp/k8s/*.rpm
$ sudo sed -i -e 's/cgroup-driver=systemd/cgroup-driver=cgroupfs/g' /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
$ sudo systemctl enable kubelet.service && systemctl start kubelet.servic
$ sudo export KUBE_ETCD_IMAGE=gcr.io/google_containers/etcd-amd64:3.0.17
$ sudo kubeadm init --kubernetes-version=v1.7.2 --pod-network-cidr=10.96.0.0/12
$ sudo export KUBECONFIG=/etc/kubernetes/admin.conf

# install flannel network
$ sudo kubectl apply -f http://$HTTP_SERVER/network/kube-flannel-rbac.yml
$ sudo kubectl apply -f http://$HTTP_SERVER/network/kube-flannel.yml --namespace=kube-system

#install dashboard
$ sudo kubectl create -f http://$HTTP_SERVER/network/kubernetes-dashboard.yml
```
配置环境变量

```
$ sudo echo "export KUBECONFIG=/etc/kubernetes/admin.conf" >> ~/.bashrc
#reload ~/.bashrc
$ sudo source ~/.bashrc
```
获取token
```
$ sudo kubeadm token list
```
minion 安装 k8s版本V1.7.2

```
$ sudo tar zxf /tmp/k8s.tar.gz -C /tmp
$ sudo yum localinstall -y  /tmp/k8s/*.rpm
$ sudo sed -i -e 's/cgroup-driver=systemd/cgroup-driver=cgroupfs/g' /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
$ sudo systemctl enable kubelet.service && systemctl start kubelet.service

# join master gluster
kubeadm join --skip-preflight-checks join --token=6669b1.81f129bc847154f9 192.168.1.201:6443
```
安装完成后查看集群信息

```
$ sudo kubectl get nodes
```

![image](C:/Users/ziye/Desktop/6.png)

- kubernetes 认证配置

  Kubernetes 系统提供了三种认证方式：CA 认证、Token 认证 和 Base 认证。kube-apiserver同时监听两个端口：insecure-port（8000）和secure-port（6443）。通过secure-port的流量将经过k8s的安全机制。insecure-port的存在一般是为了集群bootstrap或集群开发调试使用的。
  ![image](C:/Users/ziye/Desktop/6.svg)
  
API Serve Token 配置
1. 生成Static Token File，token文件是至少包含3列的csv格式文件： token, user name, user uid，第四列为可选group 名项。注意：如果有多个group名，列必须用””双引号包含其中。
2. 在/etc/kubernetes/manifests/kube-apiserver.yaml 文件中添加参数--token-auth-file
 ![image](C:/Users/ziye/Desktop/7.png)
3. 创建user和binding role

```
$ sudo kubectl create clusterrolebinding apiadmin --clusterrole=cluster-admin --user=cluster-admin
```
测试token是否有效

```
curl -k -H "Authorization: Bearer password" https://192.168.1.201:6443/
```
如果返回是kubectl api则说明token通过

##### **注意：每次修改token需要重启apiserver，目前找到重启apiserver的方式就是修改apiserver.yaml文件后apiserver会自动重启，重启后需要需要重启docker和kubelet。**

- 大规模集群自动部署

   当面对大规模集群安装时，如果一台台人工安装是不可行的，需要通过脚本实现自动安装部署。

设计思路
1. 传入master和minion的ip,user,password参数。
2. 通说ip,user password可以SSH到对应服务器上，执行安装脚本。
3. 安装脚本和安装包做成服务提供http下载。

代码如下：

```
#------------------------------------安装Master------------------------------------
systemctl disable firewalld && systemctl stop firewallddcoer

#安装docker
curl -L http://$HTTP_SERVER/install_docker.sh | bash -s $HTTP_SERVER $PRIVATE_REGISTRY

#安装kubernetes
curl -L http://$HTTP_SERVER/install_k8s.sh > /tmp/k8s/install_k8s.sh
chmod +x /tmp/k8s/install_k8s.sh
/tmp/k8s/install_k8s.sh

# Change cgroup-driver for kubelet
sed -i -e 's/cgroup-driver=systemd/cgroup-driver=cgroupfs/g' /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
systemctl enable kubelet.service && systemctl start kubelet.service &&

#设置hosts 欺骗kubeadm
echo $HTTP_SERVER storage.googleapis.com >> /etc/hosts
# 这里一定要带上--pod-network-cidr参数，不然后面的flannel网络会出问题
export KUBE_ETCD_IMAGE=gcr.io/google_containers/etcd-amd64:3.0.17
kubeadm init --kubernetes-version=v1.7.2 --pod-network-cidr=10.96.0.0/12
export KUBECONFIG=/etc/kubernetes/admin.conf
# install flannel network
kubectl apply -f http://$HTTP_SERVER/yaml/kube-flannel-rbac.yml
kubectl apply -f http://$HTTP_SERVER/yaml/kube-flannel.yml --namespace=kube-system
#install dashboard
kubectl create -f http://$HTTP_SERVER/yaml/kubernetes-dashboard.yml
# show pods
kubectl get po --all-namespaces
# show tokens
result=`kubeadm token list`
temp=${result##*DESCRIPTION}
token=${temp%%<forever*}
echo token:$token
echo "export KUBECONFIG=/etc/kubernetes/admin.conf" >> ~/.bashrc
source ~/.bashrc

#-------------------------------------安装minion-----------------------------------------
curl -L http://$HTTP_SERVER/login.sh > /tmp/k8s/login.sh
chmod +x /tmp/k8s/login.sh
for((i=2;i<=$#;i++));
  do
      param=${!i}
      host=${param%%/*}
      user=${host%%/*}
      psw=${user%%/*}
      /tmp/k8s/login.sh $host $user $psw $token $PRIVATE_REGISTRY $1:6443
  done
```


```
#!/usr/bin/expect -f
set hostname [lindex $argv 0]
set user [lindex $argv 1]
set passwd [lindex $argv 2]
set server [lindex $argv 3]
set token [lindex $argv 4]
set master [lindex $argv 5]
set registry [lindex $argv 6]
set timeout 20
spawn ssh $user@$hostname
expect {
 "*continue connecting (yes/no)?" { send "yes\r" }
 "*password*" { send "$passwd\r" }
}
expect {
 "*continue connecting (yes/no)?" { send "yes\r" }
 "*password*" { send "$passwd\r" }
}
expect "#"
send "curl -L http://$server/install.sh | bash -s $server $registry $token $master\r"
expect "#"
send "exit\r"
```





