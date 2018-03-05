#!/bin/bash

export MASTER="spark://SPARK_MASTER:7077"
export SPARK_HOME=/opt/spark
export ZEPPELIN_NOTEBOOK_DIR="${ZEPPELIN_HOME}/notebook"
export ZEPPELIN_MEM=-Xmx1024m
export ZEPPELIN_PORT=8080
export PYTHONPATH="${SPARK_HOME}/python:${SPARK_HOME}/python/lib/py4j-0.8.2.1-src.zip"
