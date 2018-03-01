# （科学上网）Linux 配置 Shadowsocks 客户端

## 背景

我们可以通过在境外配置一台安装 Shadowsocks 应用的 VPS 服务器作为代理，从而可以通过 Window 系统上安装 Shadowsocks 客户端而实现对境外网站的访问。但是 Linux 系统下如何配置 Shadowsocks 客户端呢？

由于安装 k8s 使用官方源需要使用代理，故需要在 CentOS7 配置代理进行科学上网。

## 安装 Shadowsocks 客户端

安装epel源、安装pip包管理以及安装Shadowsocks客户端：
```
sudo yum -y install epel-release
sudo yum -y install python-pip
sudo pip install shadowsocks
```

## 配置 Shadowsocks 连接

新建配置文件（默认不存在）：
```
sudo mkdir /etc/shadowsocks
sudo vi /etc/shadowsocks/shadowsocks.json
```
在配置文件中添加配置信息（需要有 Shadowsocks 服务器的地址、端口等信息）：
```
{
    "server":"x.x.x.x",  # Shadowsocks服务器地址
    "server_port":1035,  # Shadowsocks服务器端口
    "local_address": "127.0.0.1", # 本地IP
    "local_port":1080,  # 本地端口
    "password":"password", # Shadowsocks连接密码
    "timeout":300,  # 等待超时时间
    "method":"aes-256-cfb",  # 加密方式
    "fast_open": false,  # true或false。开启fast_open以降低延迟，但要求Linux内核在3.7+
    "workers": 1  #工作线程数 
}
```
为了配置自启动，新建启动脚本文件`/etc/systemd/system/shadowsocks.service`，内容如下：
```
[Unit]
Description=Shadowsocks
[Service]
TimeoutStartSec=0
ExecStart=/usr/bin/sslocal -c /etc/shadowsocks/shadowsocks.json
[Install]
WantedBy=multi-user.target
```
启动 Shadowsocks 服务：
```
systemctl enable shadowsocks.service
systemctl start shadowsocks.service
systemctl status shadowsocks.service
```
验证 Shadowsocks 客户端服务是否正常运行：
```
curl --socks5 127.0.0.1:1080 http://httpbin.org/ip
```
若 Shadowsock 客户端服务已正常运行，则结果如下：
```
{
  "origin": "x.x.x.x"       #你的Shadowsock服务器IP
}
```

## 安装配置 privoxy

以上步骤安装好了 Shadowsocks，但它是 socks5 代理，我们在 shell 里执行的命令发起的网络请求现在还不支持 socks5 代理，只支持 http／https 代理。为了我门需要安装 privoxy 代理，它能把电脑上所有 http 请求转发给 Shadowsocks。 

安装 privoxy：
```
yum install privoxy -y
systemctl enable privoxy
systemctl start privoxy
systemctl status privoxy
```
配置 privoxy（修改配置文件 /etc/privoxy/config）：
```
listen-address 127.0.0.1:8118 # 8118 是默认端口，不用改
forward-socks5t / 127.0.0.1:1080 . #转发到本地端口，注意最后有个点
```

## 设置代理

设置 http、https 代理：
```
# vi /etc/profile 在最后添加如下信息
PROXY_HOST=127.0.0.1
export all_proxy=http://$PROXY_HOST:8118
export ftp_proxy=http://$PROXY_HOST:8118
export http_proxy=http://$PROXY_HOST:8118
export https_proxy=http://$PROXY_HOST:8118
export no_proxy=localhost,172.16.0.0/16,192.168.0.0/16,127.0.0.1,10.10.0.0/16,mirrors.aliyun.com # 此处是不使用代理的访问网址，加入了阿里云镜像

# 重载环境变量
source /etc/profile
```

## 测试代理
```
[root@aniu-k8s ~]# curl -I www.google.com 

HTTP/1.1 200 OK
Date: Fri, 26 Jan 2018 05:32:37 GMT
Expires: -1
Cache-Control: private, max-age=0
Content-Type: text/html; charset=ISO-8859-1
P3P: CP="This is not a P3P policy! See g.co/p3phelp for more info."
Server: gws
X-XSS-Protection: 1; mode=block
X-Frame-Options: SAMEORIGIN
Set-Cookie: 1P_JAR=2018-01-26-05; expires=Sun, 25-Feb-2018 05:32:37 GMT; path=/; domain=.google.com
Set-Cookie: NID=122=PIiGck3gwvrrJSaiwkSKJ5UrfO4WtAO80T4yipOx4R4O0zcgOEdvsKRePWN1DFM66g8PPF4aouhY4JIs7tENdRm7H9hkq5xm4y1yNJ-sZzwVJCLY_OK37sfI5LnSBtb7; expires=Sat, 28-Jul-2018 05:32:37 GMT; path=/; domain=.google.com; HttpOnly
Transfer-Encoding: chunked
Accept-Ranges: none
Vary: Accept-Encoding
Proxy-Connection: keep-alive
```

## 取消使用代理
```
while read var; do unset $var; done < <(env | grep -i proxy | awk -F= '{print $1}')
```


<!--stackedit_data:
eyJoaXN0b3J5IjpbMTQyNDM1MDgyOF19
-->