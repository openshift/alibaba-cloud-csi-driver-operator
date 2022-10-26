package volumesnapshotclasscontroller

import (
	"context"
	"strings"
	"time"

	"github.com/openshift/alibaba-disk-csi-driver-operator/pkg/alibaba"
	operatorv1 "github.com/openshift/api/operator/v1"
	configinformer "github.com/openshift/client-go/config/informers/externalversions/config/v1"
	infralisterv1 "github.com/openshift/client-go/config/listers/config/v1"
	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/library-go/pkg/operator/resource/resourceread"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
	apiextclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
)

type VolumeSnapshotClassController struct {
	manifest       []byte
	infraLister    infralisterv1.InfrastructureLister
	apiExtClient   apiextclient.Interface
	dynamicClient  dynamic.Interface
	operatorClient v1helpers.OperatorClient
	recorder       events.Recorder
}

func NewVolumeSnapshotClassController(
	name string,
	manifest []byte,
	infraInformer configinformer.InfrastructureInformer,
	apiExtClient apiextclient.Interface,
	dynamicClient dynamic.Interface,
	operatorClient v1helpers.OperatorClient,
	resyncInterval time.Duration,
	recorder events.Recorder) factory.Controller {

	c := &VolumeSnapshotClassController{
		manifest:       manifest,
		infraLister:    infraInformer.Lister(),
		apiExtClient:   apiExtClient,
		dynamicClient:  dynamicClient,
		operatorClient: operatorClient,
		recorder:       recorder.WithComponentSuffix(name),
	}

	return factory.New().WithSync(c.sync).ResyncEvery(resyncInterval).WithSyncDegradedOnError(operatorClient).WithInformers(
		operatorClient.Informer(),
		infraInformer.Informer(),
	).ToController(name, recorder)
}

func (c *VolumeSnapshotClassController) sync(ctx context.Context, syncCtx factory.SyncContext) error {
	opSpec, _, _, err := c.operatorClient.GetOperatorState()
	if err != nil {
		return err
	}
	if opSpec.ManagementState != operatorv1.Managed {
		return nil
	}

	crdName := "volumesnapshotclasses.snapshot.storage.k8s.io"
	_, err = c.apiExtClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crdName, metav1.GetOptions{})
	if err != nil {
		// Ignore the error: we don't want to either set the Degraded condition nor log too often
		return nil
	}

	vsc, err := c.getVolumeSnapshotClass()
	if err != nil {
		return err
	}

	_, _, err = resourceapply.ApplyVolumeSnapshotClass(ctx, c.dynamicClient, c.recorder, vsc)
	return err
}

func (c *VolumeSnapshotClassController) getVolumeSnapshotClass() (*unstructured.Unstructured, error) {
	resourceGroupID, err := alibaba.GetResourceGroupID(c.infraLister)
	if err != nil {
		return nil, err
	}

	// Add double quotes to make empty resourceGroupID a valid yaml string "".
	resourceGroupID = `"` + resourceGroupID + `"`
	klog.V(4).Infof("Using resourceGroupID: %s", resourceGroupID)

	pairs := []string{
		"${RESOURCE_GROUP_ID}", resourceGroupID,
	}
	replaced := strings.NewReplacer(pairs...).Replace(string(c.manifest))

	vsc := resourceread.ReadUnstructuredOrDie([]byte(replaced))
	return vsc, nil
}
