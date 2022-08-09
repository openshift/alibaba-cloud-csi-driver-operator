package operator

import (
	"github.com/openshift/alibaba-disk-csi-driver-operator/pkg/alibaba"
	opv1 "github.com/openshift/api/operator/v1"
	infralisterv1 "github.com/openshift/client-go/config/listers/config/v1"
	"github.com/openshift/library-go/pkg/operator/csi/csistorageclasscontroller"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/klog/v2"
)

func getResourceGroupHook(infraLister infralisterv1.InfrastructureLister) csistorageclasscontroller.StorageClassHookFunc {
	return func(_ *opv1.OperatorSpec, class *storagev1.StorageClass) error {
		resourceGroupID, err := alibaba.GetResourceGroupID(infraLister)
		if err != nil {
			return err
		}
		if resourceGroupID != "" {
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
