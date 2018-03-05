.PHONY: all tensorflow-cpu tensorflow-gpu tensorflow-cpu-cluster tensorflow-gpu-cluster

all: server client tensorflow-cpu tensorflow-gpu tensorflow-cpu-cluster tensorflow-gpu-cluster

server: server.go
	go build -o server server.go
client: client.go
	go build -o client client.go

tensorflow-cpu:
	docker build -t nscc/tensorflow:1.3.0-cpu -f tensorflow:1.3.0-cpu .

tensorflow-gpu:
	docker build -t nscc/tensorflow:1.3.0-gpu -f tensorflow:1.3.0-gpu .

tensorflow-cpu-cluster:
	docker build -t nscc/tensorflow:1.3.0-cpu-cluster -f tensorflow:1.3.0-cpu-cluster .

tensorflow-gpu-cluster:
	docker build -t nscc/tensorflow:1.3.0-gpu-cluster -f tensorflow:1.3.0-gpu-cluster .
