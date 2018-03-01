# Docker CE 安装

首先安装 Docker 的依赖包如下：
```
$ yum install yum-utils device-mapper-persistent-data lvm2
```
添加 Docker 的软件源（这里使用的是国内源，也可以挂代理来使用官方的源）：
```
$ sudo yum-config-manager \
    --add-repo \
    https://mirrors.ustc.edu.cn/docker-ce/linux/centos/docker-ce.repo
```
更新 yum 软件源缓存，并安装 docker-ce ：
```
$ sudo yum makecache fast
$ sudo yum install docker-ce
```
启动 Docker CE：
```
$ sudo systemctl enable docker
$ sudo systemctl start docker
```
> PS：默认情况下，docker 命令会使用 Unix socket 与 Docker 引擎通讯。而只有 root 用户和 docker 组的用户才可以访问 Docker 引擎的 Unix socket。出于安全考虑，一般 Linux 系统上不会直接使用 root 用户。因此，更好的做法是将需要使用 docker 的用户加入 docker 用户组。

国内从 Docker Hub 拉取镜像有时会遇到困难，此时可以配置镜像加速器。对于使用 CentOS 系统，请在 /etc/docker/daemon.json 中写入如下内容（如果文件不存在请新建该文件）
```
{
  "registry-mirrors": [
    "https://registry.docker-cn.com"
  ]
}
```
之后重新启动服务：
```
$ sudo systemctl daemon-reload
$ sudo systemctl restart docker
```
测试 Docker 是否安装正确：
```
$ docker run hello-world
```
若输出如下则说明安装成功：
```
Hello from Docker!
This message shows that your installation appears to be working correctly.

To generate this message, Docker took the following steps:
 1. The Docker client contacted the Docker daemon.
 2. The Docker daemon pulled the "hello-world" image from the Docker Hub.
    (amd64)
 3. The Docker daemon created a new container from that image which runs the
    executable that produces the output you are currently reading.
 4. The Docker daemon streamed that output to the Docker client, which sent it
    to your terminal.

To try something more ambitious, you can run an Ubuntu container with:
 $ docker run -it ubuntu bash

Share images, automate workflows, and more with a free Docker ID:
 https://cloud.docker.com/

For more examples and ideas, visit:
 https://docs.docker.com/engine/userguide/
```
若出现以下报错：
```
docker: Error response from daemon: OCI runtime create failed: unable to retrieve OCI runtime error (open /run/docker/containerd/daemon/io.containerd.runtime.v1.linux/moby/225edd3d808116d3cc5992849e60bf5369ace67c291a066ebae4ca5784bcce7a/log.json: no such file or directory): docker-runc did not terminate sucessfully: unknown.
```
则是因为 libseccomp 没有更新到最新版本，安装最新版本即可：
```
yum install http://mirror.centos.org/centos/7/os/x86_64/Packages/libseccomp-2.3.1-3.el7.x86_64.rpm
```

默认配置下，如果在 CentOS 使用 Docker CE 看到下面的这些警告信息：
```
WARNING: bridge-nf-call-iptables is disabled
WARNING: bridge-nf-call-ip6tables is disabled
```
请添加内核配置参数以启用这些功能：
```
$ sudo tee -a /etc/sysctl.conf <<-EOF
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
```
然后重新加载 sysctl.conf 即可：
```
$ sudo sysctl -p
```
## 参考资料
- [CentOS 安装 Docker CE](https://yeasy.gitbooks.io/docker_practice/content/install/centos.html)
- [Docker 官方 CentOS 安装文档](https://docs.docker.com/install/linux/docker-ce/centos/#set-up-the-repository)
- [docker-runc did not terminate sucessfully: unknown](https://github.com/moby/moby/issues/35906)
<!--stackedit_data:
eyJoaXN0b3J5IjpbLTQ5MzE3MDg5M119
-->