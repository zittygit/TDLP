mysql -uroot -p"AI&BIGDATA" -e "drop user kubernetes"
mysql -uroot -p"AI&BIGDATA" -e "drop database kubernetes"
mysql -uroot -p"AI&BIGDATA" -e "create user 'kubernetes'@'%' identified by 'kubernetes'"
mysql -uroot -p"AI&BIGDATA" -e "create database kubernetes"
mysql -uroot -p"AI&BIGDATA" -e "grant all on kubernetes.* to 'kubernetes'@'%'"
mysql -ukubernetes -pkubernetes kubernetes < createTable.sql
mysql -ukubernetes -pkubernetes kubernetes -e "insert into groups value(1001, 'guoguixin')"
mysql -ukubernetes -pkubernetes kubernetes -e "insert into users(uid, userName, email, active, createTime, lastLogin, gid) value(1001, 'guoguixin', 'guixin.guo@nscc-gz.cn', 1, now(), now(), 1001)"
mysql -ukubernetes -pkubernetes kubernetes -e "insert into templates(tid, templateName, path, info, param) value(1, 'spark', 'bin/spark', 'Spark (tools for bigdata processing)', '{[{\"name\":\"name\",\"type\":\"text\",\"regex\":\"^[a-z0-9][-a-z0-9]*[a-z0-9]$\"}, {\"name\":\"cpu\",\"type\":\"int\",\"min\":1,\"max\":240},{\"name\":\"memory\",\"type\":\"int\",\"min\":100,\"max\":64000},{\"name\":\"nodes\",\"type\":\"int\",\"min\":1,\"max\":64}]}')"
mysql -ukubernetes -pkubernetes kubernetes -e "insert into templates(tid, templateName, path, info, param) value(2, 'tensorflow_cpu', 'bin/tensorflow-cpu', 'cpu version of tensorflow', '{[{\"name\":\"name\",\"type\":\"text\",\"regex\":\"^[a-z0-9][-a-z0-9]*[a-z0-9]$\"}, {\"name\":\"cpu\",\"type\":\"int\",\"min\":1,\"max\":240},{\"name\":\"memory\",\"type\":\"int\",\"min\":100,\"max\":64000}]}')"
mysql -ukubernetes -pkubernetes kubernetes -e "insert into templates(tid, templateName, path, info, param) value(3, 'tensorflow_cpu_cluster', 'bin/tensorflow-cpu-cluster', 'cpu cluster version of tensorflow', '{[{\"name\":\"name\",\"type\":\"text\",\"regex\":\"^[a-z0-9][-a-z0-9]*[a-z0-9]$\"}, {\"name\":\"cpu\",\"type\":\"int\",\"min\":1,\"max\":240},{\"name\":\"memory\",\"type\":\"int\",\"min\":100,\"max\":64000},{\"name\":\"nodes\",\"type\":\"int\",\"min\":1,\"max\":64}]}')"
mysql -ukubernetes -pkubernetes kubernetes -e "insert into templates(tid, templateName, path, info, param) value(4, 'rstudio', 'bin/rstudio', 'Rstudio (web frontend for R)', '{[{\"name\":\"name\",\"type\":\"text\",\"regex\":\"^[a-z0-9][-a-z0-9]*[a-z0-9]$\"}, {\"name\":\"cpu\",\"type\":\"int\",\"min\":1,\"max\":240},{\"name\":\"memory\",\"type\":\"int\",\"min\":100,\"max\":64000}]}')"
mysql -ukubernetes -pkubernetes kubernetes -e "insert into templates(tid, templateName, path, info, param) value(5, 'slurm', 'bin/slurm', 'SLURM (Resurce Managerment for linux cluster)', '{[{\"name\":\"name\",\"type\":\"text\",\"regex\":\"^[a-z0-9][-a-z0-9]*[a-z0-9]$\"}, {\"name\":\"cpu\",\"type\":\"int\",\"min\":1,\"max\":240},{\"name\":\"memory\",\"type\":\"int\",\"min\":100,\"max\":64000},{\"name\":\"nodes\",\"type\":\"int\",\"min\":1,\"max\":64}]}')"