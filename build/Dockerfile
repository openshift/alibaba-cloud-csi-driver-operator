FROM registry.cn-hangzhou.aliyuncs.com/plugins/centos:7.9.2009
LABEL maintainers="Alibaba Cloud Authors"
LABEL description="Alibaba Cloud Csi Driver Operator"

COPY alibaba-cloud-csi-driver-operator /bin/alibaba-cloud-csi-driver-operator

COPY /deploy/plugin /deploy/
COPY /deploy/rbac /deploy/
COPY /deploy/storageclass /deploy/
COPY /deploy/driver /deploy/

RUN chmod +x /bin/alibaba-cloud-csi-driver-operator

ENTRYPOINT ["/bin/alibaba-cloud-csi-driver-operator"]
