openjdk8 hadoop2.7.2 vim
bug:经常出现unknownhost异常
原因：hadoop执行job时是随机分配在节点上，所以每个节点需要知道其他节点的hostname
解决方案：重写registerMaster.go 将各个节点的hostname ip 同步到各个节点上

bug 上传下载找不到页面
解决方案：采用绝对路径