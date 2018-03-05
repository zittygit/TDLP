.PHONY: all test clean

all: bin/proxy bin/server bin/rstudio bin/slurm bin/spark bin/tensorflow-cpu-cluster bin/tensorflow-cpu

bin/proxy: proxy/proxy.go
	go build -o bin/proxy proxy/proxy.go

bin/server: server/server.go
	go build -o bin/server server/server.go

bin/rstudio: templates/rstudio.go
	go build -o bin/rstudio templates/rstudio.go

bin/slurm: templates/slurm.go
	go build -o bin/slurm templates/slurm.go

bin/spark: templates/spark.go
	go build -o bin/spark templates/spark.go

bin/tensorflow-cpu-cluster: templates/tensorflow-cpu-cluster.go
	go build -o bin/tensorflow-cpu-cluster templates/tensorflow-cpu-cluster.go

bin/tensorflow-cpu: templates/tensorflow-cpu.go
	go build -o bin/tensorflow-cpu templates/tensorflow-cpu.go

clean:
	rm -f bin/proxy bin/server bin/rstudio bin/slurm bin/spark bin/tensorflow-cpu-cluster bin/tensorflow-cpu proxy/proxy.log server/server.log