#!/usr/bin/env bash
set -e

cd ${GOPATH}/src/github.com/JiaoDean/alibaba-cloud-csi-driver-operator
GIT_SHA=`git rev-parse --short HEAD || echo "HEAD"`

rm -rf build/alibaba-cloud-csi-driver-operator

export GOARCH="amd64"
export GOOS="linux"

branch="v1.0.0"
version="v1.20.0"
commitId=${GIT_SHA}
buildTime=`date "+%Y-%m-%d-%H:%M:%S"`

CGO_ENABLED=0 go build -o ./build/alibaba-cloud-csi-driver-operator

cd ${GOPATH}/src/github.com/JiaoDean/alibaba-cloud-csi-driver-operator/build/

if [[ "$1" == "" ]]; then
  docker build -t=registry.cn-beijing.aliyuncs.com/acs1/alibaba-cloud-csi-driver-operator:${version}-${GIT_SHA} ./
  docker push registry.cn-beijing.aliyuncs.com/acs1/alibaba-cloud-csi-driver-operator:${version}-${GIT_SHA}
fi
