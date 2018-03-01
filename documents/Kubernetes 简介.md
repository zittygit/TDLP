# Kubernetes 简介

## 什么是 Kubernetes

Kubernetes 是 Google 开源的容器集群管理系统，它构建在 Docker 技术之上，为容器化的应用提供资源调度、部署运行、服务发现、扩容缩容等功能。利用 Kubernetes 可以方便管理跨机器运行的容器化应用，它的主要功能如下：

- 使用 Docker 对应用程序包装(package)、实例化(instantiate)、运行(run)
- 以集群的方式运行、管理跨机器的容器
- 解决 Docker 跨机器容器之间的通讯问题
- Kubernetes 的自我修复机制使得容器集群总是运行在用户期望的状态

> 当前 Kubernetes 支持 GCE、vShpere、CoreOS、OpenShift、Azure 等平台，除此之外也可以直接运行在物理机上。

## 操作对象

### Pod
Pod 是 Kubernetes 最基本的部署调度单元，可以包含 container，逻辑上表示某种应用的一个实例。比如一个 web 站点应用由前端、后端及数据库构建而成，这三个组件将运行在各自的容器中，那么我们可以创建包含三个 container 的 pod。

### Service
Service 是 pod 的路由代理抽象，用于解决 pod 之间的服务发现问题。因为 pod 的运行状态可动态变化(比如切换机器了、缩容过程中被终止了等)，所以访问端不能以写死 IP 的方式去访问该 pod 提供的服务。Service 的引入旨在保证 pod 的动态变化对访问端透明，访问端只需要知道 service 的地址，由 service 来提供代理。

### Replication Controller
Replication Controller 是 pod 的复制抽象，用于解决 pod 的扩容缩容问题。通常情况下，分布式应用为了性能或高可用性的考虑，需要复制多份资源，并且根据负载情况动态伸缩。通过 Replication Controller，我们可以指定一个应用需要几份复制，Kubernetes 将为每份复制创建一个 pod，并且保证实际运行 pod 数量总是与该复制数量相等（例如，当前某个 pod 宕机时，自动创建新的 pod 来替换）。

### 关于 Label
Label 是用于区分 Pod、Service、Replication Controller 的 key/value 键值对，Pod、Service、 Replication Controller 可以有多个 label，但是每个 label 的 key 只能对应一个 value。Label 是 Service 和 Replication Controller 运行的基础，为了将访问 Service 的请求转发给后端提供服务的多个容器，正是通过标识容器的 label 来选择正确的容器。同样，Replication Controller 也使用 label 来管理通过 pod 模板创建的一组容器，这样无论有多少容器， Replication Controller 都可以更加容易、方便地管理它们。


## 功能组件

Kubernetes 的集群架构是一个典型的 master/slave 模型。

![enter image description here](http://img.blog.csdn.net/20141030000449252?watermark/2/text/aHR0cDovL2Jsb2cuY3Nkbi5uZXQvemhhbmdqdW4yOTE1/font/5a6L5L2T/fontsize/400/fill/I0JBQkFCMA==/dissolve/70/gravity/Center)

master 运行三个组件：

- apiserver：作为 kubernetes 系统的入口，封装了核心对象的增删改查操作，以 RESTFul 接口方式提供给外部客户和内部组件调用。它维护的 REST 对象将持久化到 etcd（一个分布式强一致性的 key/value 存储）。
- scheduler：负责集群的资源调度，为新建的 pod 分配机器。这部分工作分出来变成一个组件，意味着可以很方便地替换成其他的调度器。
- controller-manager：负责执行各种控制器，目前有两类：
	- endpoint-controller：定期关联 service 和 pod (关联信息由 endpoint 对象维护)，保证 service 到 pod 的映射总是最新的。
	- replication-controller：定期关联 replicationController 和 pod，保证 replicationController 定义的复制数量与实际运行 pod 的数量总是一致的。

slave（也称作minion）运行两个组件：

- kubelet：负责管控 Docker 容器，如启动/停止、监控运行状态等。它会定期从 etcd 获取分配到本机的 pod，并根据 pod 信息启动或停止相应的容器。同时，它也会接收 apiserver 的 HTTP 请求，汇报 pod 的运行状态。
- proxy：负责为 pod 提供代理。它会定期从 etcd 获取所有的 service，并根据 service 信息创建代理。当某个客户 pod 要访问其他 pod 时，访问请求会经过本机 proxy 做转发。

## 工作流

上文已经提到了 Kubernetes 中最基本的三个操作对象：pod，replicationController 及 service。下面分别从它们的对象创建出发，通过时序图来描述 Kubernetes 各个组件之间的交互及其工作流。

### 创建 pod

![enter image description here](http://img.blog.csdn.net/20141030000701946?watermark/2/text/aHR0cDovL2Jsb2cuY3Nkbi5uZXQvemhhbmdqdW4yOTE1/font/5a6L5L2T/fontsize/400/fill/I0JBQkFCMA==/dissolve/70/gravity/Center)

### 创建 service

![enter image description here](http://img.blog.csdn.net/20141030000741960?watermark/2/text/aHR0cDovL2Jsb2cuY3Nkbi5uZXQvemhhbmdqdW4yOTE1/font/5a6L5L2T/fontsize/400/fill/I0JBQkFCMA==/dissolve/70/gravity/Center)

### 创建 controller 

![enter image description here](http://img.blog.csdn.net/20141030000812671?watermark/2/text/aHR0cDovL2Jsb2cuY3Nkbi5uZXQvemhhbmdqdW4yOTE1/font/5a6L5L2T/fontsize/400/fill/I0JBQkFCMA==/dissolve/70/gravity/Center)
<!--stackedit_data:
eyJoaXN0b3J5IjpbMTEzOTkyMzk4NV19
-->