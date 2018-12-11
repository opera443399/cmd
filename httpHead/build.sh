#!/bin/sh
# 2018/12/11

bin_name='httpHead'
image_name='opera443399/httphead'
image_ver=$(grep 'version' app.go |grep -Eo '[0-9].[0-9]')
GOARCH="amd64" GOOS="linux" CGO_ENABLED=0 go build -a --installsuffix cgo --ldflags="-s" -o ${bin_name}
docker build -t ${image_name}:${image_ver} .
docker images |grep "${image_name}"
rm -f ${bin_name}

