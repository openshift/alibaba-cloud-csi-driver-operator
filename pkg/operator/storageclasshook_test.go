package operator

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	configv1 "github.com/openshift/api/config/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func withParameters(sc *storagev1.StorageClass, keysAndValues ...string) *storagev1.StorageClass {
	for i := 0; i < len(keysAndValues); i += 2 {
		sc.Parameters[keysAndValues[i]] = keysAndValues[i+1]
	}
	return sc
}

func sc() *storagev1.StorageClass {
	return &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: nil,
		},
		Parameters: map[string]string{
			"type":                    "available",
			"volumeSizeAutoAvailable": "true",
		},
		Provisioner: "diskplugin.csi.alibabacloud.com",
	}
}

func TestStorageClassHook(t *testing.T) {
	tests := []struct {
		name        string
		infra       *configv1.Infrastructure
		expectedSC  *storagev1.StorageClass
		expectError bool
	}{
		{
			name: "no resourceGroupID",
			infra: &configv1.Infrastructure{
				Status: configv1.InfrastructureStatus{
					PlatformStatus: &configv1.PlatformStatus{
						AlibabaCloud: &configv1.AlibabaCloudPlatformStatus{
							ResourceGroupID: "",
						},
					},
				},
			},
			expectedSC: sc(),
		},
		{
			name: "unsupported cloud",
			infra: &configv1.Infrastructure{
				Status: configv1.InfrastructureStatus{
					PlatformStatus: &configv1.PlatformStatus{
						AWS: &configv1.AWSPlatformStatus{},
					},
				},
			},
			expectError: true,
			expectedSC:  sc(),
		},
		{
			name: "resourceGroupID set",
			infra: &configv1.Infrastructure{
				Status: configv1.InfrastructureStatus{
					PlatformStatus: &configv1.PlatformStatus{
						AlibabaCloud: &configv1.AlibabaCloudPlatformStatus{
							ResourceGroupID: "myID",
						},
					},
				},
			},
			expectError: false,
			expectedSC:  withParameters(sc(), "resourceGroupID", "myID"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			infraLister := &fakeInfraLister{test.infra}
			hook := getResourceGroupHook(infraLister)
			testSC := sc()

			err := hook(nil, testSC)

			if err != nil && !test.expectError {
				t.Errorf("got unexpected error: %s", err)
			}
			if err == nil && test.expectError {
				t.Errorf("expected error, got none")
			}
			if !equality.Semantic.DeepEqual(test.expectedSC, testSC) {
				t.Errorf("Unexpected StorageClass content:\n%s", cmp.Diff(test.expectedSC, testSC))
			}
		})
	}
}

type fakeInfraLister struct {
	infra *configv1.Infrastructure
}

func (f fakeInfraLister) List(selector labels.Selector) (ret []*configv1.Infrastructure, err error) {
	return nil, fmt.Errorf("not implemented")
}

func (f fakeInfraLister) Get(name string) (*configv1.Infrastructure, error) {
	if name != "cluster" {
		return nil, fmt.Errorf("Infrastructure %q not found", name)
	}
	return f.infra, nil
}
