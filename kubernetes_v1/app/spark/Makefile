.PHONY: all nss-pam-ldapd spark-base spark zeppelin

all: spark zeppelin

nss-pam-ldapd:
	docker build -t nscc/nss-pam-ldapd:0.8.13 -f nss-pam-ldapd:0.8.13 .

spark-base: nss-pam-ldapd
	docker build -t nscc/spark:2.1.0-base -f spark:2.1.0-base .

spark: spark-base
	docker build -t nscc/spark:2.1.0 -f spark:2.1.0 .

zeppelin: spark-base
	docker build -t nscc/zeppelin:0.7.2 -f zeppelin:0.7.2 .
