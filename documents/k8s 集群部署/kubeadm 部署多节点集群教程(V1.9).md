# kubeadm 部署多节点 k8s 集群教程（1.9.1）

kubeadm 是一个用于快速创建与扩展 k8s 集群的工具包。本教程主要讲述如何使用 kubeadm 构建一个双节点的 k8s 集群（版本为 1.9.1），构建集群中的 pod 通信网络，最后安装可视化的管理控制台。

本教程的集群节点均是由同一台 PC 所虚拟出来的两台虚机。

## 一. 前提条件

配置要求如下：
- 两台安装有 CentOS 7 操作系统的机器（命名为 kube-1 与 kube-2）
- 每台机器拥有 2G 以上内存以及 2核 以上处理器
- 每台机器关闭防火墙与 SELinux
- 每台机器彼此之间都可以通过网络联通
- 拥有一个可以“科学上网”的 ShadowSocksS 服务器

禁用 swap，以保证 kubelet 正确运行：每台机器执行`swapoff -a`。（注意：机器重启后可能需要再次禁用 swap）

确认每台机器的 MAC 地址与 product_uuid 都是独有的。查询 MAC 地址：`ifconfig -a`，查询 product_uuid：`cat /sys/class/dmi/id/product_uuid`。

拥有一个可以科学上网的 VPS 服务器：[VPS 配置 Shadowsocks 教程](https://github.com/Zouzhp3/Learn/blob/master/kubernetes/%E9%85%8D%E7%BD%AE%20VPS%20%E8%BF%9B%E8%A1%8C%E7%A7%91%E5%AD%A6%E4%B8%8A%E7%BD%91.md)。

每台机器都可以科学上网：[Linux 配置 Shadowsocks 客户端](https://github.com/Zouzhp3/Learn/blob/master/kubernetes/%28%E7%A7%91%E5%AD%A6%E4%B8%8A%E7%BD%91%29Linux%20%E9%85%8D%E7%BD%AE%20Shadowsocks%20%E5%AE%A2%E6%88%B7%E7%AB%AF.md)，以及 [为 Docker 配置网络代理](https://github.com/Zouzhp3/Learn/blob/master/kubernetes/%E4%B8%BA%20Docker%20%E9%85%8D%E7%BD%AE%E7%BD%91%E7%BB%9C%E4%BB%A3%E7%90%86.md)。如果不能科学上网的话，就会导致很多镜像无法正常下载。

## 二. 安装依赖环境

### 安装 Docker

为所有节点（逻辑上的机器）安装 Docker，官方推荐安装 v1.12 版本（过高版本将不兼容 k8s）。可参考 [Docker CE 安装](https://github.com/Zouzhp3/Learn/blob/master/kubernetes/Docker%20CE%20%E5%AE%89%E8%A3%85.md)。
```
$ yum-config-manager \
    --add-repo \
    https://mirrors.ustc.edu.cn/docker-ce/linux/centos/docker-ce.repo
$ yum makecache fast
$ yum install -y docker-ce
$ systemctl enable docker && systemctl start docker
``` ocker stmeablekertesocker
```
 ucgr    r 
    /tc/docker/eoso
  ec atuiersystemdocker  et r ocker``

### 安装 kubeadm, kubelet 与 kubectl

需要在所有节点上安装：
- kubeadm：引导集群的命令工具
- kubelet：运行在集群中所有节点上的组件，负责处理 Pods 与容器。
- kubectl：与集群交互的命令工具

若可以“科学上网”（否则需要手动下载 rpm 包），则安装命令如下：
```
$ cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF

$ yum install -y kubelet-1.9.1 kubeadm-1.9.1 kubectl-1.9.1
$ systemctl enable kubelet && systemctl start kubelet
```
> 选择 1.9.1 版本进行下载，但请注意必须确保 kubeadm， kubelet 的版本都一致，且与 kubectl 不低于 kubeadm 的版本。

为防止“科学上网”，则可直接进行下一步。否则就需要手动 kubeadm 初始化 k8s 时 RHEL/CentOS 7 的用户可能会报错配置失败：`You should ensure net.bridge.bridge-nf-call-iptables is set to 1 in your sysctl config`。需要执行如下命令：
```
# k8s.conf是k8s的配置文件
$ cat <<EOF >  /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
$ sysctl --system
```

若 Docker 也配置了代理“科学上网”，则可直接进行下一步，否则就需要手动下载如下镜像到本地：
```
REPOSITORY                                               TAG                 IMAGE ID            CREATED             SIZE
gcr.io/google_containers/kube-apiserver-amd64            v1.9.1              e313a3e9d78d        7 weeks ago         210.4 MB
gcr.io/google_containers/kube-scheduler-amd64            v1.9.1              677911f7ae8f        7 weeks ago         62.7 MB
gcr.io/google_containers/kube-proxy-amd64                v1.9.1              e470f20528f9        7 weeks ago         109.1 MB
gcr.io/google_containers/kube-controller-manager-amd64   v1.9.1              4978f9a64966        7 weeks ago         137.8 MB
quay.io/coreos/flannel                                   v0.9.1-amd64        2b736d06ca4c        3 months ago        51.31 MB
gcr.io/google_containers/k8s-dns-sidecar-amd64           1.14.7              db76ee297b85        4 months ago        42.03 MB
gcr.io/google_containers/k8s-dns-kube-dns-amd64          1.14.7              5d049a8c4eec        4 months ago        50.27 MB
gcr.io/google_containers/k8s-dns-dnsmasq-nanny-amd64     1.14.7              5feec37454f4        4 months ago        40.95 MB
gcr.io/google_containers/etcd-amd64                      3.1.10              1406502a6459        5 months ago        192.7 MB
gcr.io/google_containers/pause-amd64                     3.0                 99e59f495ffa        22 months ago       746.9 kB
```
可通过 [官网](https://kubernetes.io/docs/reference/setup-tools/kubeadm/kubeadm-init/) 来查看所需手动下载的依赖镜像的版本。

## 三. 使用 kubeadm 初始化集群

在一个节点（该节点将会成为集群的 master ）上使用`kubeadm init --kubernetes-version 1.9.1 --pod-network-cidr=10.244.0.0/16`来初始化一个集群（`--pod-network-cidr` 在下一节介绍）。

> 注意：若主机开启了“科学上网”的网络访问代理的话，需要先关掉主机的代理，否则初始化集群时访问内部 IP 也会经过代理，从而导致报错。

若运行成功，则输出将如下所示：
```
[kubeadm] WARNING: kubeadm is in beta, please do not use it for production clusters.
[init] Using Kubernetes version: v1.8.0
[init] Using Authorization modes: [Node RBAC]
[preflight] Running pre-flight checks
[kubeadm] WARNING: starting in 1.8, tokens expire after 24 hours by default (if you require a non-expiring token use --token-ttl 0)
[certificates] Generated ca certificate and key.
[certificates] Generated apiserver certificate and key.
[certificates] apiserver serving cert is signed for DNS names [kubeadm-master kubernetes kubernetes.default kubernetes.default.svc kubernetes.default.svc.cluster.local] and IPs [10.96.0.1 10.138.0.4]
[certificates] Generated apiserver-kubelet-client certificate and key.
[certificates] Generated sa key and public key.
[certificates] Generated front-proxy-ca certificate and key.
[certificates] Generated front-proxy-client certificate and key.
[certificates] Valid certificates and keys now exist in "/etc/kubernetes/pki"
[kubeconfig] Wrote KubeConfig file to disk: "admin.conf"
[kubeconfig] Wrote KubeConfig file to disk: "kubelet.conf"
[kubeconfig] Wrote KubeConfig file to disk: "controller-manager.conf"
[kubeconfig] Wrote KubeConfig file to disk: "scheduler.conf"
[controlplane] Wrote Static Pod manifest for component kube-apiserver to "/etc/kubernetes/manifests/kube-apiserver.yaml"
[controlplane] Wrote Static Pod manifest for component kube-controller-manager to "/etc/kubernetes/manifests/kube-controller-manager.yaml"
[controlplane] Wrote Static Pod manifest for component kube-scheduler to "/etc/kubernetes/manifests/kube-scheduler.yaml"
[etcd] Wrote Static Pod manifest for a local etcd instance to "/etc/kubernetes/manifests/etcd.yaml"
[init] Waiting for the kubelet to boot up the control plane as Static Pods from directory "/etc/kubernetes/manifests"
[init] This often takes around a minute; or longer if the control plane images have to be pulled.
[apiclient] All control plane components are healthy after 39.511972 seconds
[uploadconfig] Storing the configuration used in ConfigMap "kubeadm-config" in the "kube-system" Namespace
[markmaster] Will mark node master as master by adding a label and a taint
[markmaster] Master master tainted and labelled with key/value: node-role.kubernetes.io/master=""
[bootstraptoken] Using token: <token>
[bootstraptoken] Configured RBAC rules to allow Node Bootstrap tokens to post CSRs in order for nodes to get long term certificate credentials
[bootstraptoken] Configured RBAC rules to allow the csrapprover controller automatically approve CSRs from a Node Bootstrap Token
[bootstraptoken] Creating the "cluster-info" ConfigMap in the "kube-public" namespace
[addons] Applied essential addon: kube-dns
[addons] Applied essential addon: kube-proxy

Your Kubernetes master has initialized successfully!

To start using your cluster, you need to run (as a regular user):

  mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config

You should now deploy a pod network to the cluster.
Run "kubectl apply -f [podnetwork].yaml" with one of the options listed at:
  http://kubernetes.io/docs/admin/addons/

You can now join any number of machines by running the following on each node
as root:

  kubeadm join --token <token> <master-ip>:<master-port> --discovery-token-ca-cert-hash sha256:<hash>
```
初始化完毕后，运行以下命令给予用户权限来使用集群：
```
$ mkdir -p $HOME/.kube
$ sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
$ sudo chown $(id -u):$(id -g) $HOME/.kube/config
```
执行 `kubectl get nodes`，发现得到了一个状态为`NotReady`的 Node。

查看一下集群状态`kubectl get cs`，确认各个组件都处于 healthy 状态。：
```
NAME                 STATUS    MESSAGE              ERROR
scheduler            Healthy   ok
controller-manager   Healthy   ok
etcd-0               Healthy   {"health": "true"}
```

查看集群组件 pod 运行情况：`kubectl get pods --all-namespaces`，正常情况下的输出如下：
```
NAMESPACE     NAME                             READY     STATUS    RESTARTS   AGE
kube-system   etcd-kube-1                      1/1       Running   0          1h
kube-system   kube-apiserver-kube-1            1/1       Running   0          1h
kube-system   kube-controller-manager-kube-1   1/1       Running   0          1h
kube-system   kube-dns-6f4fd4bdf-jthnq         0/3       Pending   0          1h
kube-system   kube-proxy-2r2m4                 1/1       Running   0          1h
kube-system   kube-scheduler-kube-1            1/1       Running   0          1h
```

> 上面输出的 kube-dns 的状态是正常的，因为集群还没有配置网络。

此外，集群初始化如果遇到问题，可以使用下面的命令进行清理：
```
$ kubeadm reset
$ ifconfig cni0 down
$ ip link delete cni0
$ ifconfig flannel.1 down
$ ip link delete flannel.1
$ rm -rf /var/lib/cni/
```

## 四. 配置集群网络（Flannel）

集群必须安装一个 pod 网络插件以便于 pods 能够互相通信，本教程中使用 Flannel 作为集群配置网络。

> 为使 flannel 运行成功，`kubeadm init` 运行时必须加上参数`--pod-network-cidr=10.244.0.0/16`。

运行：
```
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/v0.9.1/Documentation/kube-flannel.yml
```
正常输出如下：
```
clusterrole "flannel" created
clusterrolebinding "flannel" created
serviceaccount "flannel" created
configmap "kube-flannel-cfg" created
daemonset "kube-flannel-ds" created
```
等待一段时间后，查看组件 pod 的运行情况：`kubectl get pods --all-namespaces`，若所有pods 都处于运行成功的状态，则说明网络部署成功。然后执行 `kubectl get nodes` 也可以发现 Node 已经处于 Ready 状态了。

> 注意：若主机有多个网卡，则可能会遭遇错误如右：[flannel issues 3970](https://github.com/kubernetes/kubernetes/issues/39701)。
> 解决该问题（尚未测试）：目前需要在 kube-flannel.yml 中使用`--iface`参数指定集群主机内网网卡的名称，否则可能会导致 dns 无法解析。因此需要将 kube-flannel.yml 下载到本地，flanneld 启动参数加上 `--iface=<iface-name>`。

### 配置 master 节点是否调度 pod（可选）

出于安全性考虑，在默认情况下 pod 不会被调度到 master 节点上，也就是说它不参与工作负载。但如果需要 master 也能调度 pod，以便于构造一个单节点集群用于开发用，则可以执行以下命令：
```
$ kubectl taint nodes --all node-role.kubernetes.io/master-
```
输出如下：
```
node "test-01" untainted
taint key="dedicated" and effect="" not found.
taint key="dedicated" and effect="" not found.
```

## 五. 向集群中添加节点

Node 是集群中负责运行容器与 Pod 的节点，当需要添加一个主机到集群中成为一个新 Node 时，ssh 连接到该主机，切换到 root 用户权限，运行 master 节点 `kubeadm init` 时的输出中的参考命令，如下：
```
$ kubeadm join --token <token> <master-ip>:<master-port> --discovery-token-ca-cert-hash sha256:<hash>
```
若运行成功则输出如下：
```
[preflight] Running pre-flight checks.
	[WARNING FileExisting-crictl]: crictl not found in system path
[discovery] Trying to connect to API Server "192.168.80.128:6443"
[discovery] Created cluster-info discovery client, requesting info from "https://192.168.80.128:6443"
[discovery] Requesting info from "https://192.168.80.128:6443" again to validate TLS against the pinned public key
[discovery] Cluster info signature and contents are valid and TLS certificate validates against pinned roots, will use API Server "192.168.80.128:6443"
[discovery] Successfully established connection with API Server "192.168.80.128:6443"

This node has joined the cluster:
* Certificate signing request was sent to master and a response
  was received.
* The Kubelet was informed of the new secure connection details.

Run 'kubectl get nodes' on the master to see this node join the cluster.
```
过一段时间后查询 `kubectl get nodes` 即可得到处于 Ready 状态的新 Node。

### 尝试运行一个应用（可选&重要）
```
$ kubectl run curl --image=radial/busyboxplus:curl -i --tty
If you don't see a command prompt, try pressing enter.
[ root@curl-2716574283-xr8zd:/ ]$
```
进入后执行 `nslookup kubernetes.default` 确认是否解析正常:
```
$ nslookup kubernetes.default
Server:    10.96.0.10
Address 1: 10.96.0.10 kube-dns.kube-system.svc.cluster.local

Name:      kubernetes.default
Address 1: 10.96.0.1 kubernetes.default.svc.cluster.local
```

### 从 master 以外的节点来控制集群（可选&重要）

为了让其他节点（或者集群外部的节点）上的 kubectl 可以与集群通信，你需要从 master 节点复制管理员集群配置文件 `admin.conf` 到目标节点，命令如下：
```
$ scp root@<master ip>:/etc/kubernetes/admin.conf .
$ kubectl --kubeconfig ./admin.conf get nodes
```
> `admin.conf` 给予用户控制集群的超级权限，必须谨慎使用。
> 对于普通用户而言，建议生成一个独一凭证，使得放置该用户于白名单中：`$ kubeadm alpha phase kubeconfig user --client-name <CN>`，该命令将会输出一个 KubeConfig 文件，你可以保存该文件并发给该普通用户。然后使用`$ kubectl create (cluster)rolebinding` 启动白名单。

### 配置 API Server 的代理到本地 localhost（可选&重要）

若需要从**集群外部**连接到集群的 API Server，则可以使用`kubectl proxy`：
```
$ scp root@<master ip>:/etc/kubernetes/admin.conf .
$ kubectl --kubeconfig ./admin.conf proxy
```
这样就可以在本地通过 `http://localhost:8001/api/v1` 来访问集群的 API Server 了。

## 六. 从集群中删除节点

当需要从一个集群中删除节点时，首先需要停止该节点以确保该节点在关闭前是空的。

在 master 上运行以下命令：
```
$ kubectl drain <node name> --delete-local-data --force --ignore-daemonsets
$ kubectl delete node <node name>
```
然后在被删除的节点上重置 kubeadm 的状态即可：
```
$ kubeadm reset
```

## 七. 安装 Dashboard

[Kubernetes Dashboard](https://github.com/kubernetes/dashboard) 是一个基于 web 的 k8s 集群控制台，它允许用户管理和调试已经运行在集群上的应用，甚至可以管理集群本身。

下载 Dashboard 的 yaml 配置文件：
```
$ wget https://raw.githubusercontent.com/kubernetes/dashboard/master/src/deploy/recommended/kubernetes-dashboard.yaml
```
由于Dashboard 的 service 配置模式是 ClusterIP 而不能被集群外访问，因此我们需要配置成 NodePort 模式。编辑 `kubernetes-dashboard.yaml` 文件，在 `Dashboard Service` 中添加`type: NodePort`，暴露 Dashboard 服务：
```
# ------------------- Dashboard Service ------------------- #

kind: Service
apiVersion: v1
metadata:
  labels:
    k8s-app: kubernetes-dashboard
  name: kubernetes-dashboard
  namespace: kube-system
spec:
  type: NodePort
  ports:
    - port: 443
      targetPort: 8443
  selector:
    k8s-app: kubernetes-dashboard
```
根据配置文件安装 Dashboard（若不能“科学上网”则可能下载镜像失败）：
```
$ kubectl create -f kubernetes-dashboard.yaml
```
执行 `kubectl proxy` 启动代理后，可以在**代理机本地**通过以下 URL 访问 Dashboard 网站：
```
http://localhost:8001/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/
```
也可以在外部通过`https://[NodeIP]:[NodePort]`来访问。

查看网站页面的登录 token：
```
$ kubectl -n kube-system get secret | grep kubernetes-dashboard
kubernetes-dashboard-token-jxq7l kubernetes.io/service-account-token 3  22h
$ kubectl describe -n kube-system secret/kubernetes-dashboard-token-jxq7l
```
当需要卸载 Dashboard ：
```
$ kubectl delete -f kubernetes-dashboard.yaml
```

### 如何使用管理员权限？

如果我们直接使用上面获取的 token 登录 Dashboard 的网站后，发现几乎大部分的权限都不可以使用。 这是因为默认的 `kubernetes-dashboard.yaml` 文件中的 ServiceAccount `kubernetes-dashboard` 只有相对较小的权限。

因此我们需要创建一个 `kubernetes-dashboard-admin` 的 ServiceAccount 并授予其集群 admin 的权限。创建 `kubernetes-dashboard-admin.rbac.yaml`：
```
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    k8s-app: kubernetes-dashboard
  name: kubernetes-dashboard-admin
  namespace: kube-system
  
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: kubernetes-dashboard-admin
  labels:
    k8s-app: kubernetes-dashboard
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: kubernetes-dashboard-admin
  namespace: kube-system
```
执行该文件：
```
$ kubectl create -f kubernetes-dashboard-admin.rbac.yaml
```
再按照同样的方法查询到 `kubernetes-dashboard-admin` 的 token ，登录 Dashborad 网站后便可以使用所有功能了。

## 八. Heapster 插件部署

安装 [Heapster](https://github.com/kubernetes/heapster/) 可以为集群添加使用统计和监控功能，为 Dashboard 添加仪表盘。使用 InfluxDB 做为 Heapster 的后端存储，开始部署：

```
mkdir -p ~/k8s/heapster
cd ~/k8s/heapster
wget https://raw.githubusercontent.com/kubernetes/heapster/master/deploy/kube-config/influxdb/grafana.yaml
wget https://raw.githubusercontent.com/kubernetes/heapster/master/deploy/kube-config/rbac/heapster-rbac.yaml
wget https://raw.githubusercontent.com/kubernetes/heapster/master/deploy/kube-config/influxdb/heapster.yaml
wget https://raw.githubusercontent.com/kubernetes/heapster/master/deploy/kube-config/influxdb/influxdb.yaml

kubectl create -f ./
```
最后确认所有的 pod 都处于 running 状态，打开 Dashboard，集群的使用统计会以仪表盘的形式显示出来。
## 四. 配置集群网络（Flannel）

集群必须安装一个 pod 网络插件以便于 pods 能够互相通信，本教程中使用 Flannel 作为集群配置网络。

> 为使 flannel 运行成功，`kubeadm init` 运行时必须加上参数`--pod-network-cidr=10.244.0.0/16`。

运行：
```
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/v0.9.1/Documentation/kube-flannel.yml
```

## 注意事项

### 1. 为何配置了科学上网也无法 pull gcr.io 的镜像？

主机网络代理与 docker 的网络代理设置是不同的，你还需要设置 docker 的网络代理：[为 Docker 配置网络代理](https://github.com/Zouzhp3/Learn/blob/master/kubernetes/%E4%B8%BA%20Docker%20%E9%85%8D%E7%BD%AE%E7%BD%91%E7%BB%9C%E4%BB%A3%E7%90%86.md)

### 2. kubelet 服务不正常运行

启动启动报错 

单独使用 `kubelet` 服务后命令时 ，发现没有正常运行，日志报错如下：
```
error: failed to run Kubelet: failed to create kubelet: misconfiguration: kubelet cgroup driver: "cgroupfs" is different from docker cgroup driver: "systemd"
``` 
kubelet 的配置文件是`/etc/systemd/system/kubelet.service.d/10-kubeadm.conf`，请确保里面`KUBELET_CGROUP_ARGS=--cgroup-driver=systemd`的 `--cgroup-driver`驱动与 `docker info` 里的 `--cgroup-driver` 驱动相同。但即使确保了相同，独自执行`kubelet`命令时也会继续报同样的错。

后来发现，该配置文件是在集群初始化时才会成功读取，而单独执行`kubelet`时不会读取配置文件。

因此无视此问题即可，不影响集群初始化。在集群初始化成功后，可以发现`kubelet`服务是正常运行的。后来发现不影响集群初始化，且成功初始化

可参考：[1.6.0 kubelet fails with error "misconfiguration: kubelet cgroup driver: "cgroupfs" is different from docker cgroup driver: "systemd"](https://github.com/kubernetes/kubernetes/issues/43805)

### 3. 运行 kubeadm init 时弹出警告

警告如下：
```
[preflight] Running pre-flight checks.
	[WARNING FileExisting-crictl]: crictl not found in system path
```
经过 github 上的开发者确认，这个 warning 可以无视。

### 4. 运行 kubeadm init 时报错

报错如下所示：
```
[kubelet-check] It seems like the kubelet isn't running or healthy.  
[kubelet-check] The HTTP call equal to 'curl -sSL [http://localhost:10255/healthz](http://localhost:10255/healthz)' failed with error: Get [http://localhost:10255/healthz](http://localhost:10255/healthz): dial tcp [::1]:10255: getsockopt: connection refused.
```

经查资料发现是因为没有禁用 swap（每次机器重启会重置 swap），但是经过重置虚拟机网络以及恢复快照后得到的系统仍然会出现此错误。后来经过重新安装 kubeadm 与 kubelet 成功解决。

### 5. 忘记了集群初始化成功时输出的参考命令

如果不小心忘记了集群初始化成功时输出的参考命令，可使用以下命令来查询。

查看 master 的 token：
```
$ kubeadm token list | grep authentication,signing | awk '{print $1}'
```
查看 master 的 discovery-token-ca-cert-hash：
```
$ openssl x509 -pubkey -in  /etc/kubernetes/pki/ca.crt | openssl rsa -pubin -outform der 2>/dev/null  | openssl dgst -sha256 -hex | sed 's/^.* //'
```

### 6. 为何设置了 NodePort 模式也无法从外部访问 Dashboard 网站？

这是因为 Dashboard 配置文件默认采取了 HTTPS 协议，因此需要以 `https://[NodeIP]:[NodePort]` 的方式来访问。同时又因为 Chrome 浏览器不能支持访问未认证的 https 网站，所以建议使用其他浏览器（如 Firefox）。


## 部署总结

从零开始学习使用 kubeadm 部署一个 k8s 集群总计花了我五天多时间，一开始大部分时间花在了如何下载合适版本的 kubeadm 和 kubelet，以及通过各种手段下载国内获取不到的镜像，但效果仍然不好。之后通过搭建 ShadowSocks 客户端使得可以成功下载合适版本的 kubeadm 和 kubelet，但发现还是 pull 不了镜像，最后发现是因为 Docker 的代理配置与主机的代理配置是不共用的。配置了 Docker 代理后便可以在初始化集群的过程中自动 pull 需要的镜像。通过使用翻墙代理，确实可以少花很多不必要的时间。

部署期间也遇到一些问题，有的是因为过于钻牛角尖，陷入到单独启动 `kubelet` 失败的问题中了，误以为该问题必须在集群初始化之前解决；有的是因为理论知识不扎实，对 Docker 的不了解；还有的是就是虚拟机网络配置突然变化导致的玄学和 Chrome 不支持不安全证书的 HTTPS 网站而导致的误判。

总而言之，以官方文档作为主线，配以他人博客上的成功部署经验，再加上使用翻墙代理，才能够又快又好地部署成功。此外，学会使用虚拟机的快照功能也是很重要的，方便进行重要操作前的备份。


## 参考资料

[Installing kubeadm](https://kubernetes.io/docs/setup/independent/install-kubeadm/) （官网）

[Using kubeadm to Create a Cluster](https://kubernetes.io/docs/setup/independent/create-cluster-kubeadm/)（官网）

[Dashboard: Creating sample user](https://github.com/kubernetes/dashboard/wiki/Creating-sample-user) （Dashboard 的 github）

[使用kubeadm安装kubernetes1.7/1.8/1.9](http://blog.csdn.net/zhuchuangang/article/details/76572157#2-%E9%85%8D%E7%BD%AEkubelet)

[使用kubeadm安装Kubernetes 1.9](https://blog.frognew.com/2017/12/kubeadm-install-kubernetes-1.9.html#1%E5%87%86%E5%A4%87)（Dashboard）

[使用kubeadm安装Kubernetes 1.8版本](https://www.kubernetes.org.cn/2906.html)

[使用 kubeadm 创建 kubernetes 1.9 集群](https://www.kubernetes.org.cn/3357.html)

[使用kubeadm在CentOS 7上安装Kubernetes 1.8](https://www.zybuluo.com/ncepuwanghui/note/953929)（Dashboard） 参考资料

[Installing kubeadm](https://kubernetes.io/docs/setup/independent/install-kubeadm/) （来自官网）

[使用kubeadm安装kubernetes1.7/1.8/1.9](http://blog.csdn.net/zhuchuangang/article/details/76572157#2-%E9%85%8D%E7%BD%AEkubelet)


<!--stackedit_data:
eyJoaXN0b3J5IjpbNjc2MDkxNDMxLDc3Nzg2ODE4OF19
-->