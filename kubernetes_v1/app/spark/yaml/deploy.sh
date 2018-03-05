#! /bin/sh

kubectl create -f namespace-test.yaml
kubectl create -f spark-master-controller.yaml
kubectl create -f spark-master-service.yaml
kubectl create -f spark-worker-controller.yaml
kubectl create -f zeppelin-controller.yaml
kubectl create -f zeppelin-service.yaml
