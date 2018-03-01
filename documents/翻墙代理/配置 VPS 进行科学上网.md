# VPS 配置 Shadowsocks 教程

## 一. 原理介绍

虚拟专用服务器（英语：Virtual private server，缩写为 VPS），是指通过虚拟化技术在独立服务器中运行的专用服务器。每个使用 VPS 技术的虚拟独立服务器拥有各自独立的公网IP地址、操作系统、硬盘空间、内存空间、CPU资源等，还可以进行安装程序、重启服务器等操作，与运行一台独立服务器基本相同。

VPS 一定是要可被访问到，并且可以访问境外网站的，所以 VPS 一般是选择境外的，常见的地区主要是香港、日本、新加坡、美国等。

[Shadowsocks](https://shadowsocks.org/en/index.html) 是一个基于 socks5 协议的代理服务器软件，原理如下所示。可以通过在 VPS 上部署 Shadowsocks 来实现“科学上网”。

![Shadowsocks 原理图](http://blog.021xt.cc/wp-content/uploads/2017/01/Shadowsocks%E5%8E%9F%E7%90%86tu.jpg)

1. 我们首先通过 SS Local 和 VPS 进行通信，通过 Socks5 协议进行通信。
2. SS Local 连接到 VPS，并对 Socks5 中传输的数据进行对称加密传输，传输的数据格式是 SS 的协议。
3. SS Server 收到请求后，对数据解密，按照 SS 协议来解析数据。
4. SS Server 根据协议内容转发请求。
5. SS Server 获取请求结果后回传给 SS Local。
6. SS Local 获取结果回传给应用程序。

## 二. 配置 Shadowsocks 服务端

Shadowsocks （以下简称“SS”）有多个语言的版本，包括 Python、Go 和 C。本教程介绍 Python 版的部署过程：
```
# 确保python是2.7版本
$ python --version 
# 使用pip进行安装
$ yum -y install epel-release
$ yum -y install python-pip
$ pip install shadowsocks
```
安装完成后需要进行配置，SS 使用 JSON 格式文件进行配置，配置文件路径和名字可以自己决定，一般放在`/etc/shadowsocks.json`，格式如下：
```
{
    "server":"my_server_ip",
    "server_port":8388,
    "local_address": "127.0.0.1",
    "local_port":1080,
    "password":"mypassword",
    "timeout":300,
    "method":"aes-256-cfb",
    "fast_open": false
}
```
各个参数的解释如下表所示：
| Name | Explanation |
|--|--|
| server | VPS的IP地址，IPV4与IPV6都可以 |
|server_port |	提供SS服务的端口号，写自己想用的端口号 |
|local_address|	the address your local listens|
|local_port	|local port|
|password	|传输数据时用来加密的密钥，和Client相同|
|timeout	|连接超时时间|
|method	|加密方法 推荐使用 “aes-256-cfb”|
|fast_open	|use TCP_FASTOPEN, true / false|
|workers	|number of workers, available on Unix/Linux|

### 多用户配置（可选）
若需要多用户配置，则可使用 port_password，每个用户对应一个端口，后面是密码：
```
{
    "server": "0.0.0.0",
    "port_password": {
        "8381": "foobar1",
        "8382": "foobar2",
        "8383": "foobar3",
        "8384": "foobar4"
    },
    "timeout": 300,
    "method": "aes-256-cfb"
}
```

### 启动与停止服务
前台运行命令：
```
$ ssserver -c /etc/shadowsocks.json
```
后台运行命令：
```
$ ssserver -c /etc/shadowsocks.json -d start
$ ssserver -c /etc/shadowsocks.json -d stop
```
启动后可以在系统中查看对应进程：
```
ps -aux | grep ssserver
```
也可以查看 log，检查 SS 的状态：
```
tail -f /var/log/shadowsocks.log
```

### 设置开机启动
在`/etc/rc.local`中加入：
```
# start the shadowsocks server
sudo ssserver -c /etc/shadowsocks.json -d start
```

## 三. 配置 Shadowsocks 客户端

Server 服务启动之后，就可以在客户端安装 SS Client 使用了。SS 目前支持 Windows、MAC、 Linux、Android、IOS 等众多平台。请参考：[Shadowsocks Clients](https://shadowsocks.org/en/download/clients.html)。

## 参考资料

[Shadowsocks 原理与搭建](http://blog.021xt.cc/archives/98)

[Shadowsocks 官网](https://shadowsocks.org/en/index.html)

[上网限制和翻墙基本原理](http://blog.021xt.cc/archives/85)

[搭建自己的 VPS ](http://blog.021xt.cc/archives/54)


<!--stackedit_data:
eyJoaXN0b3J5IjpbMTg2MjMyMDk2Ml19
-->