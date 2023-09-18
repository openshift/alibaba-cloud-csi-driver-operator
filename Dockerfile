FROM registry.ci.openshift.org/ocp/builder:rhel-8-golang-1.20-openshift-4.15 AS builder
WORKDIR /go/src/github.com/openshift/alibaba-disk-csi-driver-operator
COPY . .
RUN make

FROM registry.ci.openshift.org/ocp/4.15:base
COPY --from=builder /go/src/github.com/openshift/alibaba-disk-csi-driver-operator/alibaba-disk-csi-driver-operator /usr/bin/
ENTRYPOINT ["/usr/bin/alibaba-disk-csi-driver-operator"]
LABEL io.k8s.display-name="OpenShift Alibaba Disk CSI Driver Operator" \
	io.k8s.description="The Alibala Disk CSI Driver Operator installs and maintains the Alibaba Disk CSI Driver on a cluster."
