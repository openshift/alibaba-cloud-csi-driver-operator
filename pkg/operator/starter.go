package operator

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/openshift/library-go/pkg/controller/factory"
	"k8s.io/client-go/dynamic"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	opv1 "github.com/openshift/api/operator/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/openshift/library-go/pkg/operator/csi/csicontrollerset"
	"github.com/openshift/library-go/pkg/operator/csi/csidrivercontrollerservicecontroller"
	"github.com/openshift/library-go/pkg/operator/csi/csidrivernodeservicecontroller"
	goc "github.com/openshift/library-go/pkg/operator/genericoperatorclient"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
)

const (
	cloudConfigNamespace = "openshift-config-managed"
	defaultNamespace     = "openshift-cluster-csi-drivers"
	operatorName         = "alibaba-cloud-csi-driver-operator"
	operandName          = "alibaba-cloud-csi-driver"
	instanceName         = "csi.alibabacloud.com"
	cloudConfigName      = "kube-cloud-config"
	secretName           = "alibaba-cloud-credentials"
)

// ReadFile reads and returns the content of the named file.
func ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func RunOperator(ctx context.Context, controllerConfig *controllercmd.ControllerContext) error {
	kubeClient := kubeclient.NewForConfigOrDie(rest.AddUserAgent(controllerConfig.KubeConfig, operatorName))
	kubeInformersForNamespaces := v1helpers.NewKubeInformersForNamespaces(kubeClient, defaultNamespace, cloudConfigNamespace, "")
	secretInformer := kubeInformersForNamespaces.InformersFor(defaultNamespace).Core().V1().Secrets()
	nodeInformer := kubeInformersForNamespaces.InformersFor("").Core().V1().Nodes()
	configClient := configclient.NewForConfigOrDie(rest.AddUserAgent(controllerConfig.KubeConfig, operatorName))
	configInformers := configinformers.NewSharedInformerFactory(configClient, 20*time.Minute)
	infraInformer := configInformers.Config().V1().Infrastructures()
	cloudConfigInformer := kubeInformersForNamespaces.InformersFor(cloudConfigNamespace).Core().V1().ConfigMaps()
	gvr := opv1.SchemeGroupVersion.WithResource("clustercsidrivers")
	operatorClient, dynamicInformers, err := goc.NewClusterScopedOperatorClientWithConfigName(controllerConfig.KubeConfig, gvr, instanceName)
	if err != nil {
		return err
	}

	dynamicClient, err := dynamic.NewForConfig(controllerConfig.KubeConfig)
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
		ReadFile,
		[]string{
			"rbac/sa.yaml",
			"rbac/cluster-role.yaml",
			"rbac/cluster-role-binding.yaml",
			"driver/disk-driver.yaml",
			"driver/nas-driver.yaml",
			"driver/oss-driver.yaml",
			"storageclass/disk-avaliable-sc.yaml",
			"storageclass/disk-efficiency-sc.yaml",
			"storageclass/disk-essd-sc.yaml",
			"storageclass/disk-ssd-sc.yaml",
			"storageclass/disk-topology-sc.yaml",
			"plugin/csi-plugin.yaml",
			"plugin/csi-provisioner.yaml",
		},
	).WithCSIConfigObserverController(
		"AlibabaCloudDriverCSIConfigObserverController",
		configInformers,
	).WithCSIDriverControllerService(
		"AlibabaCloudDriverControllerServiceController",
		ReadFile,
		"controller.yaml",
		kubeClient,
		kubeInformersForNamespaces.InformersFor(defaultNamespace),
		configInformers,
		[]factory.Informer{
			secretInformer.Informer(),
			nodeInformer.Informer(),
			cloudConfigInformer.Informer(),
			infraInformer.Informer(),
		},
		csidrivercontrollerservicecontroller.WithSecretHashAnnotationHook(defaultNamespace, secretName, secretInformer),
		csidrivercontrollerservicecontroller.WithObservedProxyDeploymentHook(),
		csidrivercontrollerservicecontroller.WithReplicasHook(nodeInformer.Lister()),
	).WithCSIDriverNodeService(
		"AlibabaCloudDriverNodeServiceController",
		ReadFile,
		"node.yaml",
		kubeClient,
		kubeInformersForNamespaces.InformersFor(defaultNamespace),
		nil, // Node doesn't need to react to any changes
		csidrivernodeservicecontroller.WithObservedProxyDaemonSetHook(),
	).WithServiceMonitorController(
		"AlibabaCloudDriverServiceMonitorController",
		dynamicClient,
		ReadFile,
		"servicemonitor.yaml",
	)
	if err != nil {
		return err
	}

	klog.Info("Starting the informers")
	go kubeInformersForNamespaces.Start(ctx.Done())
	go dynamicInformers.Start(ctx.Done())
	go configInformers.Start(ctx.Done())

	klog.Info("Starting controllerset")
	go csiControllerSet.Run(ctx, 1)

	<-ctx.Done()

	return fmt.Errorf("stopped")
}