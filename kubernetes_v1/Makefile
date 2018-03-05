.PHONY: all test clean

all: bin/server bin/serverd bin/firstproxy bin/firstproxyd bin/secondproxy bin/secondproxyd bin/spark bin/zeppelin bin/tensorflow-cpu bin/tensorflow-cpu-cluster test/test_app test/test_group test/test_instance test/test_jwt test/test_k8s test/test_ldapAuth test/test_ssha test/test_user

bin/server: server/server.go
	go build -o bin/server server/server.go

bin/serverd: server/serverd.go
	go build -o bin/serverd server/serverd.go

bin/firstproxy: proxy/first/firstproxy.go
	go build -o bin/firstproxy proxy/first/firstproxy.go

bin/firstproxyd: proxy/first/firstproxyd.go
	go build -o bin/firstproxyd proxy/first/firstproxyd.go

bin/secondproxy: proxy/second/secondproxy.go
	go build -o bin/secondproxy proxy/second/secondproxy.go

bin/secondproxyd: proxy/second/secondproxyd.go
	go build -o bin/secondproxyd proxy/second/secondproxyd.go

bin/spark: application/spark.go
	go build -o bin/spark application/spark.go

bin/zeppelin: application/zeppelin.go
	go build -o bin/zeppelin application/zeppelin.go

bin/tensorflow-cpu: application/tensorflow-cpu.go
	go build -o bin/tensorflow-cpu application/tensorflow-cpu.go

bin/tensorflow-cpu-cluster: application/tensorflow-cpu-cluster.go
	go build -o bin/tensorflow-cpu-cluster application/tensorflow-cpu-cluster.go

test/test_app: test/test_app.go
	go build -o test/test_app test/test_app.go

test/test_group: test/test_group.go
	go build -o test/test_group test/test_group.go

test/test_instance: test/test_instance.go
	go build -o test/test_instance test/test_instance.go

test/test_jwt: test/test_jwt.go
	go build -o test/test_jwt test/test_jwt.go

test/test_k8s: test/test_k8s.go
	go build -o test/test_k8s test/test_k8s.go

test/test_ldapAuth: test/test_ldapAuth.go
	go build -o test/test_ldapAuth test/test_ldapAuth.go

test/test_ssha: test/test_ssha.go
	go build -o test/test_ssha test/test_ssha.go

test/test_user: test/test_user.go
	go build -o test/test_user test/test_user.go

test:
	bin/server restart
	bin/firstproxy restart
	bin/secondproxy restart
	test/test_app
	test/test_group
	test/test_instance
	test/test_jwt
	test/test_k8s
	test/test_ldapAuth
	test/test_ssha
	test/test_user
	bin/server stop
	bin/firstproxy stop
	bin/secondproxy stop

clean:
	rm -f bin/server bin/serverd bin/firstproxy bin/firstproxyd bin/secondproxy bin/secondproxyd bin/spark bin/zeppelin bin/tensorflow-cpu bin/tensorflow-cpu-cluster test/test_app test/test_group test/test_instance test/test_jwt test/test_k8s test/test_ldapAuth test/test_ssha test/test_user server/server.log server/server.pid proxy/first/firstproxy.log proxy/first/firstproxy.pid proxy/second/secondproxy.log proxy/second/secondproxy.pid
