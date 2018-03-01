# 为 Docker 配置网络代理

## 背景

在一些实验环境中服务器没有直接连接外网的权限，需要通过网络代理。我们通常会将网络代理直接配置在`/etc/environment`、`/etc/profile`之类的配置文件中，这对于大部分操作都是可行的。然而，docker 命令却使用不了这些代理。比如 `docker pull` 时需要从外网下载镜像，就会出现错误。

## 步骤

首先为 docker 服务创建一个内嵌的 systemd 目录：
```
$ mkdir -p /etc/systemd/system/docker.service.d
```
创建`/etc/systemd/system/docker.service.d/http-proxy.conf`文件，并添加 HTTP_PROXY 环境变量（其中`[proxy-addr]`和`[proxy-port]`分别改成实际情况的代理地址和端口）：
```
[Service]
Environment="HTTP_PROXY=http://[proxy-addr]:[proxy-port]/" "HTTPS_PROXY=https://[proxy-addr]:[proxy-port]/"
```
如果还有内部的不需要使用代理来访问的 Docker registries，那么还需要制定 NO_PROXY 环境变量：
```
[Service]
Environment="HTTP_PROXY=http://[proxy-addr]:[proxy-port]/" "HTTPS_PROXY=https://[proxy-addr]:[proxy-port]/" "NO_PROXY=localhost,127.0.0.1,docker-registry.somecorporation.com"
```
更新配置与重启 Docker 服务：
```
$ systemctl daemon-reload
$ systemctl restart docker
```
验证配置是否已经加载：
```
$ systemctl show --property=Environment docker
Environment=HTTP_PROXY=http://proxy.example.com:80/
```

## 参考资料

[Docker 网络代理设置](http://blog.csdn.net/styshoo/article/details/55657714)

[Docker 官网教程](https://docs.docker.com/config/daemon/systemd/)
<!--stackedit_data:
eyJoaXN0b3J5IjpbMjcyNTAzOTQ4XX0=
-->