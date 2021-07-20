FROM registry.cn-hangzhou.aliyuncs.com/plugins/centos:7.9.2009
LABEL maintainers="Alibaba Cloud Authors"
LABEL description="Alibaba Cloud Csi Driver Operator"

COPY alibaba-cloud-csi-driver-operator /bin/alibaba-cloud-csi-driver-operator

COPY /assets/plugin /assets/
COPY /assets/rbac /assets/
COPY /assets/storageclass /assets/
COPY /assets/driver /assets/

RUN chmod +x /bin/alibaba-cloud-csi-driver-operator

ENTRYPOINT ["/bin/alibaba-cloud-csi-driver-operator"]
