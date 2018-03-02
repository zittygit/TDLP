#!/bin/bash
# push 镜像到私有库
# param ip ,param2 port
####example####
#sh push-images.sh 192.168.31.85 5523

set -e
for file in ./images/*
do
    if  test -f $file
    then
        fullname=${file##*/}
		name=${fullname%.*}
		echo push $name to private registry...
		imageid=`docker load < $file`
		if [[ $imageid =~ "sha256:" ]]
        then
            docker tag ${imageid#*sha256:} $1:$2/${name%%_*}:${name#*_}	
        else
  		    docker tag ${imageid#*image:} $1:$2/${name%%_*}:${name#*_}
		fi
		docker push $1:$2/${name%%_*}:${name#*_}
    fi
done
echo push complete!		