package operator

import (
	"fmt"

	opv1 "github.com/openshift/api/operator/v1"
	infralisterv1 "github.com/openshift/client-go/config/listers/config/v1"
	"github.com/openshift/library-go/pkg/operator/csi/csistorageclasscontroller"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/klog/v2"
)

func getResourceGroupHook(infraLister infralisterv1.InfrastructureLister) csistorageclasscontroller.StorageClassHookFunc {
	return func(_ *opv1.OperatorSpec, class *storagev1.StorageClass) error {
		infra, err := infraLister.Get(infrastructureName)
		if err != nil {
			return err
		}
		if infra.Status.PlatformStatus == nil {
			return fmt.Errorf("error parsing infrastructure.status: platformStatus is nil")
		}
		if infra.Status.PlatformStatus.AlibabaCloud == nil {
			return fmt.Errorf("error parsing infrastructure.status: platformStatus.alibabaCloud is nil")
		}
		if infra.Status.PlatformStatus.AlibabaCloud.ResourceGroupID != "" {
			resourceGroupID := infra.Status.PlatformStatus.AlibabaCloud.ResourceGroupID
			klog.V(4).Infof("Using resourceGroupID %q", resourceGroupID)
			if class.Parameters == nil {
				class.Parameters = map[string]string{}
			}
			class.Parameters[resourceGroupIDParam] = resourceGroupID
			return nil
		}
		klog.V(4).Infof("Using no resourceGroupID")
		return nil
	}
}
