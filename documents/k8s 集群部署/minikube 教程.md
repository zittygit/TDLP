# Minikube 教程（v0.21.4）

本教程将介绍如何通过 minikube 构建一个 k8s 集群，并在上部署一个可访问的 NodeJS 应用。

## 一. 配置 kubectl 与 Minikube

本教程开始前，你需要拥有一台内存较高（推荐为 4G）、操作系统为 **CentOS 7** 以上的机器（虚拟机或物理机）。

下载 kubectl 并配置到系统路径（注意 kubectl 的最新版本）：
```
$ curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.9.1/bin/linux/amd64/kubectl
$ chmod +x ./kubectl
$ sudo mv ./kubectl /usr/local/bin/
```
本教程中 minikube 的虚拟机驱动采用 VirtualBox ，安装步骤如下：
1. 在官网下载对应系统的 VirtualBox rpm 包
2. 安装 VirtualBox 的依赖包：`yum install qt qt-x11 gcc gcc-c++ kernel-devel perl SDL`
3. 安装 VirtualBox：`rpm -i VirtualBox-5.2-5.2.6_120293_el7-1.x86_64.rpm`
4. 添加当前用户到 VirtualBox 创建的用户组 "vboxusers"：`usermod -a -G vboxusers {{用户名}}`

下载 Minikube 并配置到系统路径：
```
$ curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 
$ chmod +x ./minikube
$ sudo mv ./minikube /usr/local/bin/
```
完成以上步骤后，执行 `minikube start` 启动本地 k8s 集群。当正常启动成功时，出现类似以下输出：
```
Starting local Kubernetes v1.8.0 cluster...
Starting VM...
Downloading Minikube ISO
 140.01 MB / 140.01 MB [============================================] 100.00% 0s
Getting VM IP address...
Moving files into cluster...
Downloading localkube binary
 148.25 MB / 148.25 MB [============================================] 100.00% 0s
Connecting to cluster...
Setting up kubeconfig...
Starting cluster components...
Kubectl is now configured to use the cluster.
Loading cached images from config file.
```
执行 `minikube status` 可查看当前集群状态。

## 二. 安装 Docker 与 NodeJS

### 1. Docker 的安装
请参见 [Docker CE 安装教程](https://github.com/Zouzhp3/Learn/blob/master/kubernetes/Docker%20CE%20%E5%AE%89%E8%A3%85.md)

### 2. NodeJS 的安装
访问 [NodeJS 官网](https://nodejs.org/en/download/)，选择需要的版本进行下载。由于下载的文件类型是 xz，因此可能需要安装能够解压 xz 格式的工具。解压成功后获得一个后缀为 tar 的文件。再使用 `tar -xf node-v8.9.4-linux-x64.tar` 解压该文件。为了验证是否能够使用 node，我们可以输入`./node-v8.9.4/bin/node -v`查看 node 版本，安装成功则能够成功显示 node 版本。最后配置到全局路径即可：
```
$ ln -s /root/node-v8.9.4/bin/node /usr/local/bin/node 
$ ln -s /root/node-v8.9.4/bin/npm /usr/local/bin/npm
```

## 三. 创建 JS 应用

创建一个名为`hellonode`的目录，以及在目录中创建一个名为`server.js`的文件，内容如下：
```
var http = require('http');

var handleRequest = function(request, response) {
  console.log('Received request for URL: ' + request.url);
  response.writeHead(200);
  response.end('Hello World!');
};
var www = http.createServer(handleRequest);
www.listen(8080);
```
可使用 `node server.js` 运行应用，则可通过`http://localhost:8080/`进行网页访问。

## 四. 构建Docker容器镜像

在`hellonode`目录中创建一个名为`Dockerfile`的文件，Dockerfile 描述了你所要创建的镜像。以下的 Dockerfile 则是由 NodeJS 的镜像扩展而来：
```
FROM node:6.9.2
EXPOSE 8080
COPY server.js .
CMD node server.js
```
这个镜像从 Node.js 的 Docker 仓库扩展而来，开放了 8080 端口，复制了 server.js 文件到了该镜像中，并运行了该应用。

由于本教程使用了 Minikube 主机而不是把镜像推到宿主机的仓库，为此需要把 Docker Deadom 切换到 Minikube 主机中的环境。：`eval $(minikube docker-env)`。这样在宿主机的终端所执行的 docker 命令都是在 Minikube 主机中执行的。

> 当不需要再使用 Minikube Host 后，可以使用以下命令来撤销该操作：`eval $(minikube docker-env -u)`

使用 Minikube Docker 来把`hellonode`目录构建为镜像：`docker build -t hello-node:v1 .`。这样 Minikube 就可以运行所创建的镜像了。

## 五. 创建 Deployment

使用`kubectl run`命令创建 Deployment 来管理 Pod。Pod 将根据所创建的镜像 hello-node:v1 来运行容器镜像：
```
$ kubectl run hello-node --image=hello-node:v1 --port=8080
```
查看Deployment：`kubectl get deployments`，输出：
```
NAME         DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
hello-node   1         1         1            1           3m
```
查看Pod：`kubectl get pods`，输出：
```
NAME                         READY     STATUS    RESTARTS   AGE
hello-node-714049816-ztzrb   1/1       Running   0          6m
```
查看群集 events：`kubectl get events`，以及查看 kubectl 配置：`kubectl config view`。

### 如何解决 pod 创建失败的问题（在国内 gcr.io 被 GFW 屏蔽）

在有一台 VPS 服务器的情况下，当然最好的办法莫过于设置“科学上网”：[(科学上网)Linux 配置 Shadowsocks 客户端](https://github.com/Zouzhp3/Learn/blob/master/kubernetes/%28%E7%A7%91%E5%AD%A6%E4%B8%8A%E7%BD%91%29Linux%20%E9%85%8D%E7%BD%AE%20Shadowsocks%20%E5%AE%A2%E6%88%B7%E7%AB%AF.md)，可以一劳永逸解决问题。如果没有 VPS，就需要按照下面的办法解决了。

运行镜像后如果服务一直是 containerCreating 状态且没有变化，则是创建实例出现问题。如下方法查看日志：
```
$ sudo minikube logs
```
若日志中出现 `failed pulling image…` 则是说明镜像拉取失败导致服务创建失败，这是因为国内 GFW 的境外访问拦截。从日志可见，创建 pod 时在拉取自身需要的 `gcr.io/google_containers/pause-amd64:3.0`镜像时失败了，报错如下：
```
Jan 05 03:52:58 minikube localkube[3624]: E0105 03:52:58.952990    3624 kuberuntime_manager.go:632] createPodSandbox for pod "nginx666-864b85987c-kvdpb_default(b0cc687d-f1cb-11e7-ba05-080027e170dd)" failed: rpc error: code = Unknown desc = failed pulling image "gcr.io/google_containers/pause-amd64:3.0": Error response from daemon: Get https://gcr.io/v2/: net/http: request canceled while waiting for connection (Client.Timeout exceeded while awaiting headers)
```
**解决方法**：可以用本地镜像替代，具体方法就是把阿里云的镜像下载到本地，然后命名为 minikube 所使用 gcr.io 的同名镜像，替代远端镜像即可。以下是按照此法拉取一个容器镜像的实例。

下载阿里云镜像：
```
$ docker pull registry.cn-hangzhou.aliyuncs.com/google-containers/pause-amd64:3.0
```
用 tag 命令映射为 `gcr.io/google_containers/pause-amd64:3.0`：
```
docker tag registry.cn-hangzhou.aliyuncs.com/google-containers/pause-amd64:3.0 gcr.io/google_containers/pause-amd64:3.0
```

> 最推荐的方法是登录[阿里云容器镜像控制台](https://cr.console.aliyun.com)来搜索所缺少的镜像，再把它们 pull 下来。

依此法继续拉取所需要但又获取不到的镜像（具体命令见文末注意事项）。

## 六. 创建 Service

默认情况，Pod 只能通过集群内部 IP 访问。要使 hello-node 容器从 k8s 虚拟网络外部访问，须要部署 Service 来暴露 Pod 。

我们可以使用`kubectl expose`命令将 Pod 暴露到外部环境：
```
$ kubectl expose deployment hello-node --type=LoadBalancer
```
查看刚创建的 Service：
```
$ kubectl get services
```
输出：
```
NAME         CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
hello-node   10.0.0.71    <pending>     8080/TCP   6m
kubernetes   10.0.0.1     <none>        443/TCP    14d
```
通过`--type=LoadBalancer`在集群外暴露 Service，支持负载均衡的云提供商可为 Service 配置外部 IP 地址。在 Minikube 上，该 LoadBalancer 类型可以通过 Minikube Service 命令来访问服务。
```
$ minikube service hello-node
```

## 七. 更新应用程序

编辑 server.js 文件以返回新消息：
```
response.end('Hello World Again!');
```
构建新版本镜像：
```
$ docker build -t hello-node:v2 .
```
Deployment 更新镜像：
```
$ kubectl set image deployment/hello-node hello-node=hello-node:v2
```
再次运行应用以查看新消息：
```
$ minikube service hello-node
```

## 八. 清理删除

现在可以删除在群集中创建的资源：
```
$ kubectl delete service hello-node
$ kubectl delete deployment hello-node
```
或者停止Minikube：
```
$ minikube stop
```

## 注意事项

### 1. minikube host 里拉取官方镜像时速度太慢

这是因为虽然本机里设置了 Docker 国内镜像加速器，但是 minikube host 里是没有设置的。只需按照一样的方法在 minikube host 中设置镜像加速器即可。参见 [Docker CE 安装教程](https://github.com/Zouzhp3/Learn/blob/master/kubernetes/Docker%20CE%20%E5%AE%89%E8%A3%85.md)。

### 2. 安装 gnome 桌面后导致 virtualbox 启动失败

重装 gnome 桌面后，启动 virtualbox 时报错如下：
```
VirtualBox: supR3HardenedMainGetTrustedMain: dlopen("/usr/lib/virtualbox/VirtualBox.so",) failed: /lib64/libGL.so.1: undefined symbol: drmFreeDevice
```
这是因为安装 gnome 时自动更新了依赖软件包`mesa-libGL`的版本，导致 virtualbox 启动出错。

花费一天多的时间也找不到方法解决。启动 minikube 时可以用 `--vm-driver=none` 代替，或者用 KVM2 代替 virtualbox。

### 3. 使用 kvm2 作为 minikube 虚拟机驱动中遇到的问题

[minikube 使用 kvm2 的官方教程](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#kvm2-driver)。

按照教程完成所有步骤后，启动`minikube start --vm-driver kvm2`时会报错找不到 libvirtd 的相关文件，这是因为 libvirtd 服务没有启动，启动它即可。

之后又会报错`Domain not found: no domain with matching name 'minikube'`，这是因为没有安装新版的 kernel 与 kernel-devel。

但最后还是会报错：`Error starting host: Temporary Error: Error configuring auth on host: OS type not recognized.`

简直变态，没法解决。

> 因此 gnome 最好不要与 virtualbox 同时使用，这些虚拟机驱动都太麻烦了。

### 4. 通过阿里云获取 minikube 所需镜像的命令脚本

注意请查看 minikube 的 log 来确定镜像的版本。
```
$ eval $(minikube docker-env)

# 拉取 gcr.io/google_containers/pause-amd64:3.0
$ docker pull registry.cn-hangzhou.aliyuncs.com/google-containers/pause-amd64:3.0
$ docker tag registry.cn-hangzhou.aliyuncs.com/google-containers/pause-amd64:3.0 gcr.io/google_containers/pause-amd64:3.0

# 拉取 gcr.io/google-containers/kube-addon-manager:v6.4-beta.2
$ docker pull registry.cn-hangzhou.aliyuncs.com/google_containers/kube-addon-manager:v6.4-beta.2
$ docker tag registry.cn-hangzhou.aliyuncs.com/google_containers/kube-addon-manager:v6.4-beta.2 gcr.io/google-containers/kube-addon-manager:v6.4-beta.2

# 拉取 gcr.io/k8s-minikube/storage-provisioner:v1.8.1
$ docker pull registry.cn-hangzhou.aliyuncs.com/google_containers/storage-provisioner:v1.8.1
$ docker tag registry.cn-hangzhou.aliyuncs.com/google_containers/storage-provisioner:v1.8.1 gcr.io/k8s-minikube/storage-provisioner:v1.8.1

# 拉取 gcr.io/google_containers/k8s-dns-dnsmasq-nanny-amd64:1.14.5
$ docker pull registry.cn-hangzhou.aliyuncs.com/google_containers/k8s-dns-dnsmasq-nanny-amd64:1.14.5
$ docker tag registry.cn-hangzhou.aliyuncs.com/google_containers/k8s-dns-dnsmasq-nanny-amd64:1.14.5 gcr.io/google_containers/k8s-dns-dnsmasq-nanny-amd64:1.14.5

# 拉取 gcr.io/google_containers/k8s-dns-sidecar-amd64:1.14.5
$ docker pull registry.cn-hangzhou.aliyuncs.com/google_containers/k8s-dns-sidecar-amd64:1.14.5
$ docker tag registry.cn-hangzhou.aliyuncs.com/google_containers/k8s-dns-sidecar-amd64:1.14.5 gcr.io/google_containers/k8s-dns-sidecar-amd64:1.14.5

# 拉取 gcr.io/google_containers/kubernetes-dashboard-amd64:v1.8.0
$ docker pull registry.cn-hangzhou.aliyuncs.com/gcr_k8s/kubernetes-dashboard-amd64:v1.8.0
$ docker tag registry.cn-hangzhou.aliyuncs.com/gcr_k8s/kubernetes-dashboard-amd64:v1.8.0 gcr.io/google_containers/kubernetes-dashboard-amd64:v1.8.0

# 拉取 gcr.io/google_containers/k8s-dns-kube-dns-amd64:1.14.5
docker pull registry.cn-hangzhou.aliyuncs.com/outman_google_containers/k8s-dns-kube-dns-amd64:1.14.5
docker tag registry.cn-hangzhou.aliyuncs.com/outman_google_containers/k8s-dns-kube-dns-amd64:1.14.5 gcr.io/google_containers/k8s-dns-kube-dns-amd64:1.14.5
```

## 参考资料
[CentOS 下安装 nodejs 并配置环境](http://blog.csdn.net/qq_21794603/article/details/68067821)

[官方minikube教程](https://kubernetes.io/docs/tutorials/stateless-application/hello-minikube/)

[Minikube 本地集群](https://zhulg.github.io/2017/11/08/mac-minikube%E6%9C%AC%E5%9C%B0%E9%9B%86%E7%BE%A4/)

[使用minikube在本机搭建kubernetes集群](https://www.centos.bz/2018/01/%E4%BD%BF%E7%94%A8minikube%E5%9C%A8%E6%9C%AC%E6%9C%BA%E6%90%AD%E5%BB%BAkubernetes%E9%9B%86%E7%BE%A4/)

[利用Minikube来部署一个nodejs应用](https://www.jianshu.com/p/c8bb49edf466)
<!--stackedit_data:
eyJoaXN0b3J5IjpbMjAzMDE2MzY5NF19
-->