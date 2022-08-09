package volumesnapshotclasscontroller

import (
	"context"
	"reflect"
	"testing"
	"time"

	snap "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	configv1 "github.com/openshift/api/config/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	fakeconfig "github.com/openshift/client-go/config/clientset/versioned/fake"
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resource/resourceread"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
)

func infra(resourceGroupID string) *configv1.Infrastructure {
	return &configv1.Infrastructure{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Status: configv1.InfrastructureStatus{
			PlatformStatus: &configv1.PlatformStatus{
				AlibabaCloud: &configv1.AlibabaCloudPlatformStatus{
					ResourceGroupID: resourceGroupID,
				},
			},
		},
	}
}

func TestSync(t *testing.T) {
	volumeSnapshotClassHeader := `
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotClass
metadata:
  name: alicloud-disk
driver: diskplugin.csi.alibabacloud.com
deletionPolicy: Delete
parameters:
`

	tests := []struct {
		name             string
		infra            *configv1.Infrastructure
		inputManifest    string
		expectedManifest string
		expectError      bool
	}{
		{
			name:             "no resource ID",
			infra:            infra(""),
			inputManifest:    volumeSnapshotClassHeader + "  resourceGroupID: ${RESOURCE_GROUP_ID}\n",
			expectedManifest: volumeSnapshotClassHeader + "  resourceGroupID: \"\"\n",
			expectError:      false,
		},
		{
			name:             "resource ID",
			infra:            infra("MyID"),
			inputManifest:    volumeSnapshotClassHeader + "  resourceGroupID: ${RESOURCE_GROUP_ID}\n",
			expectedManifest: volumeSnapshotClassHeader + "  resourceGroupID: MyID\n",
			expectError:      false,
		},
		{
			name: "invalid infra",
			infra: &configv1.Infrastructure{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Status: configv1.InfrastructureStatus{
					PlatformStatus: &configv1.PlatformStatus{
						AlibabaCloud: nil,
					},
				},
			},
			inputManifest:    volumeSnapshotClassHeader + "  resourceGroupID: ${RESOURCE_GROUP_ID}\n",
			expectedManifest: "",
			expectError:      true,
		},
	}
	snap.AddToScheme(scheme.Scheme)

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			dynamicClient := fake.NewSimpleDynamicClient(scheme.Scheme)
			fakeOperatorClient := v1helpers.NewFakeOperatorClientWithObjectMeta(
				&metav1.ObjectMeta{
					Name: "cluster",
				},
				&operatorv1.OperatorSpec{
					ManagementState: operatorv1.Managed,
				},
				&operatorv1.OperatorStatus{},
				nil, /*triggerErr func*/
			)

			initialInfras := []runtime.Object{test.infra}
			configClient := fakeconfig.NewSimpleClientset(initialInfras...)
			configInformerFactory := configinformers.NewSharedInformerFactory(configClient, 0)
			configInformerFactory.Config().V1().Infrastructures().Informer().GetIndexer().Add(test.infra)

			ctrl := NewVolumeSnapshotClassController(
				"test",
				[]byte(test.inputManifest),
				configInformerFactory.Config().V1().Infrastructures(),
				dynamicClient,
				fakeOperatorClient,
				time.Minute*1,
				events.NewInMemoryRecorder("test"),
			)

			err := ctrl.Sync(context.TODO(), factory.NewSyncContext("test", events.NewInMemoryRecorder("test")))
			if err != nil && !test.expectError {
				t.Errorf("Expected no sync error, got %s", err)
			}
			if err == nil && test.expectError {
				t.Errorf("Expected sync error, got none")
			}

			classClient := dynamicClient.Resource(schema.GroupVersionResource{
				Group:    snap.SchemeGroupVersion.Group,
				Version:  snap.SchemeGroupVersion.Version,
				Resource: "volumesnapshotclasses",
			})
			if test.expectedManifest != "" {
				expectedClass := resourceread.ReadUnstructuredOrDie([]byte(test.expectedManifest))
				actualClass, err := classClient.Get(context.TODO(), expectedClass.GetName(), metav1.GetOptions{})
				if err != nil {
					t.Fatalf("Failed to get VolumeSnapshotClass %s: %s", expectedClass.GetName(), err)
				}
				if !reflect.DeepEqual(expectedClass, actualClass) {
					t.Errorf("Expected VolumeSnapshotClass:\n%+v\ngot:\n%+v", expectedClass, actualClass)
				}
			}
		})
	}
}
