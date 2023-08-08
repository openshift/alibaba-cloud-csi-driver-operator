package operator

import (
	"context"
	"fmt"
	"time"

	"github.com/openshift/alibaba-disk-csi-driver-operator/assets"
	"github.com/openshift/alibaba-disk-csi-driver-operator/pkg/operator/volumesnapshotclasscontroller"
	"k8s.io/client-go/dynamic"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	opv1 "github.com/openshift/api/operator/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	opclient "github.com/openshift/client-go/operator/clientset/versioned"
	opinformers "github.com/openshift/client-go/operator/informers/externalversions"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/csi/csicontrollerset"
	"github.com/openshift/library-go/pkg/operator/csi/csidrivercontrollerservicecontroller"
	"github.com/openshift/library-go/pkg/operator/csi/csidrivernodeservicecontroller"
	goc "github.com/openshift/library-go/pkg/operator/genericoperatorclient"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
	apiextclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

const (
	defaultNamespace      = "openshift-cluster-csi-drivers"
	operatorName          = "alibaba-cloud-csi-driver-operator"
	operandName           = "alibaba-cloud-csi-driver"
	instanceName          = "diskplugin.csi.alibabacloud.com"
	cloudCredSecretName   = "alibaba-cloud-credentials"
	metricsCertSecretName = "alibaba-disk-csi-driver-controller-metrics-serving-cert"
	trustedCAConfigMap    = "alibaba-disk-csi-driver-trusted-ca-bundle"
	resourceGroupIDParam  = "resourceGroupId"
	resync                = 20 * time.Minute
)

func RunOperator(ctx context.Context, controllerConfig *controllercmd.ControllerContext) error {
	kubeClient := kubeclient.NewForConfigOrDie(rest.AddUserAgent(controllerConfig.KubeConfig, operatorName))
	kubeInformersForNamespaces := v1helpers.NewKubeInformersForNamespaces(kubeClient, defaultNamespace, "")
	secretInformer := kubeInformersForNamespaces.InformersFor(defaultNamespace).Core().V1().Secrets()
	configMapInformer := kubeInformersForNamespaces.InformersFor(defaultNamespace).Core().V1().ConfigMaps()
	nodeInformer := kubeInformersForNamespaces.InformersFor("").Core().V1().Nodes()
	configClient := configclient.NewForConfigOrDie(rest.AddUserAgent(controllerConfig.KubeConfig, operatorName))
	configInformers := configinformers.NewSharedInformerFactory(configClient, resync)

	// operator.openshift.io client, used for ClusterCSIDriver
	operatorClientSet := opclient.NewForConfigOrDie(rest.AddUserAgent(controllerConfig.KubeConfig, operatorName))
	operatorInformers := opinformers.NewSharedInformerFactory(operatorClientSet, resync)

	infraInformer := configInformers.Config().V1().Infrastructures()
	gvr := opv1.SchemeGroupVersion.WithResource("clustercsidrivers")
	operatorClient, dynamicInformers, err := goc.NewClusterScopedOperatorClientWithConfigName(controllerConfig.KubeConfig, gvr, instanceName)
	if err != nil {
		return err
	}

	dynamicClient, err := dynamic.NewForConfig(controllerConfig.KubeConfig)
	if err != nil {
		return err
	}

	// Create apiextension client. This is used to verify is a VolumeSnapshotClass CRD exists.
	apiExtClient, err := apiextclient.NewForConfig(rest.AddUserAgent(controllerConfig.KubeConfig, operatorName))
	if err != nil {
		return err
	}

	csiControllerSet := csicontrollerset.NewCSIControllerSet(
		operatorClient,
		controllerConfig.EventRecorder,
	).WithLogLevelController().WithManagementStateController(
		operandName,
		false,
	).WithStaticResourcesController(
		"AlibabaCloudDriverStaticResourcesController",
		kubeClient,
		dynamicClient,
		kubeInformersForNamespaces,
		assets.ReadFile,
		[]string{
			"csidriver.yaml",
			"controller_sa.yaml",
			"controller_pdb.yaml",
			"node_sa.yaml",
			"service.yaml",
			"cabundle_cm.yaml",
			"rbac/attacher_role.yaml",
			"rbac/attacher_binding.yaml",
			"rbac/privileged_role.yaml",
			"rbac/controller_privileged_binding.yaml",
			"rbac/node_privileged_binding.yaml",
			"rbac/provisioner_role.yaml",
			"rbac/provisioner_binding.yaml",
			"rbac/resizer_role.yaml",
			"rbac/resizer_binding.yaml",
			"rbac/snapshotter_role.yaml",
			"rbac/snapshotter_binding.yaml",
			"rbac/prometheus_role.yaml",
			"rbac/prometheus_rolebinding.yaml",
			"rbac/kube_rbac_proxy_role.yaml",
			"rbac/kube_rbac_proxy_binding.yaml",
		},
	).WithCSIConfigObserverController(
		"AlibabaCloudDriverCSIConfigObserverController",
		configInformers,
	).WithCSIDriverControllerService(
		"AlibabaCloudDriverControllerServiceController",
		assets.ReadFile,
		"controller.yaml",
		kubeClient,
		kubeInformersForNamespaces.InformersFor(defaultNamespace),
		configInformers,
		[]factory.Informer{
			secretInformer.Informer(),
			nodeInformer.Informer(),
			infraInformer.Informer(),
			configMapInformer.Informer(),
		},
		csidrivercontrollerservicecontroller.WithSecretHashAnnotationHook(defaultNamespace, cloudCredSecretName, secretInformer),
		csidrivercontrollerservicecontroller.WithSecretHashAnnotationHook(defaultNamespace, metricsCertSecretName, secretInformer),
		csidrivercontrollerservicecontroller.WithObservedProxyDeploymentHook(),
		csidrivercontrollerservicecontroller.WithCABundleDeploymentHook(
			defaultNamespace,
			trustedCAConfigMap,
			configMapInformer,
		),
		csidrivercontrollerservicecontroller.WithReplicasHook(nodeInformer.Lister()),
	).WithCSIDriverNodeService(
		"AlibabaCloudDriverNodeServiceController",
		assets.ReadFile,
		"node.yaml",
		kubeClient,
		kubeInformersForNamespaces.InformersFor(defaultNamespace),
		[]factory.Informer{configMapInformer.Informer()},
		csidrivernodeservicecontroller.WithObservedProxyDaemonSetHook(),
		csidrivernodeservicecontroller.WithCABundleDaemonSetHook(
			defaultNamespace,
			trustedCAConfigMap,
			configMapInformer,
		),
	).WithStorageClassController(
		"AlibabaCloudStorageClassController",
		assets.ReadFile,
		[]string{
			"storageclass.yaml",
		},
		kubeClient,
		kubeInformersForNamespaces.InformersFor(""),
		operatorInformers,
		getResourceGroupHook(infraInformer.Lister()),
	)
	if err != nil {
		return err
	}

	volumeSnapshotClassAsset, err := assets.ReadFile("volumesnapshotclass.yaml")
	if err != nil {
		return err
	}
	volumeSnapshotCtrl := volumesnapshotclasscontroller.NewVolumeSnapshotClassController(
		"VolumeSnapshotController",
		volumeSnapshotClassAsset,
		infraInformer,
		apiExtClient,
		dynamicClient,
		operatorClient,
		time.Minute,
		controllerConfig.EventRecorder,
	)

	klog.Info("Starting the informers")
	go kubeInformersForNamespaces.Start(ctx.Done())
	go dynamicInformers.Start(ctx.Done())
	go configInformers.Start(ctx.Done())
	go operatorInformers.Start(ctx.Done())

	klog.Info("Starting VolumeSnapshotController")
	go volumeSnapshotCtrl.Run(ctx, 1)

	klog.Info("Starting controllerset")
	go csiControllerSet.Run(ctx, 1)
	<-ctx.Done()

	return fmt.Errorf("stopped")
}
