#! /bin/sh

cp kubernetes.repo /etc/yum.repos.d
yum install -y kubeadm docker
systemctl start docker
./load_image.sh
join=$(kubeadm init --pod-network-cidr=10.244.0.0/16 2>&1 | tail -n 2 | head -n 1)
echo $join > join.sh
mkdir -p $HOME/.kube
cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
chown $(id -u):$(id -g) $HOME/.kube/config
kubectl apply -f kube-flannel.yml
kubectl apply -f kube-flannel-rbac.yml
kubectl apply -f kubernetes-dashboard.yaml
