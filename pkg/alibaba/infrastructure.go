package alibaba

import (
	"fmt"

	infralisterv1 "github.com/openshift/client-go/config/listers/config/v1"
)

const (
	infrastructureName = "cluster"
)

func GetResourceGroupID(infraLister infralisterv1.InfrastructureLister) (string, error) {
	infra, err := infraLister.Get(infrastructureName)
	if err != nil {
		return "", err
	}
	if infra.Status.PlatformStatus == nil {
		return "", fmt.Errorf("error parsing infrastructure.status: platformStatus is nil")
	}
	if infra.Status.PlatformStatus.AlibabaCloud == nil {
		return "", fmt.Errorf("error parsing infrastructure.status: platformStatus.alibabaCloud is nil")
	}
	return infra.Status.PlatformStatus.AlibabaCloud.ResourceGroupID, nil
}
