#!/bin/sh
# 2018/12/12

bin_name='httpHead'
image_name='opera443399/httphead'
image_ver=$(grep 'ENV APP_VERSION' Dockerfile |awk '{print $NF}')
GOARCH="amd64" GOOS="linux" CGO_ENABLED=0 go build -a --installsuffix cgo --ldflags="-s" -o ${bin_name}
docker build -t ${image_name}:${image_ver} .
docker tag ${image_name}:${image_ver} ${image_name}
docker images |grep "${image_name}"
rm -f ${bin_name}

