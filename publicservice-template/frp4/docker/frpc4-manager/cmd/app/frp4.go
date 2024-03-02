package app

import (
	"context"
	"flag"

	"github.com/spf13/cobra"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"

	openappclient "github.com/openapp-dev/openapp/pkg/generated/clientset/versioned"
	openappinformer "github.com/openapp-dev/openapp/pkg/generated/informers/externalversions"
	"github.com/openapp-dev/openapp/pkg/utils"
	"github.com/openapp-dev/publicservice/frp4/pkg/controller"
)

func NewFrp4GatewayCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "frp4-gateway",
		Long: `frp4-gateway used to expose an app to public internet`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := run(ctx); err != nil {
				return err
			}
			return nil
		},
	}

	fss := cliflag.NamedFlagSets{}
	logFlagSet := fss.FlagSet("log")
	klog.InitFlags(flag.CommandLine)
	logFlagSet.AddGoFlagSet(flag.CommandLine)
	cmd.Flags().AddFlagSet(logFlagSet)

	return cmd
}

func run(ctx context.Context) error {
	klog.Info("Start frp4-gateway...")
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatalf("Failed to get in-cluster config: %v", err)
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Failed to create k8s client: %v", err)
	}
	openappClient, err := openappclient.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Failed to create client: %v", err)
	}
	k8sFactory := informers.NewSharedInformerFactory(k8sClient, 0)
	openappFactory := openappinformer.NewSharedInformerFactory(openappClient, 0)

	configMapInformer := k8sFactory.Core().V1().ConfigMaps().Informer()
	serviceInformer := k8sFactory.Core().V1().Services().Informer()
	statefulSetInformer := k8sFactory.Apps().V1().StatefulSets().Informer()
	appInstanceInformer := openappFactory.App().V1alpha1().AppInstances().Informer()
	appTemplateInformer := openappFactory.App().V1alpha1().AppTemplates().Informer()
	serviceInstanceInformer := openappFactory.Service().V1alpha1().PublicServiceInstances().Informer()
	serviceTemplateInformer := openappFactory.Service().V1alpha1().PublicServiceTemplates().Informer()

	openappHelper := utils.NewOpenAPPHelper(k8sClient,
		openappClient,
		configMapInformer,
		serviceInformer,
		appInstanceInformer,
		serviceInstanceInformer,
		statefulSetInformer,
		k8sFactory.Core().V1().ConfigMaps().Lister(),
		openappFactory.App().V1alpha1().AppInstances().Lister(),
		openappFactory.App().V1alpha1().AppTemplates().Lister(),
		openappFactory.Service().V1alpha1().PublicServiceInstances().Lister(),
		openappFactory.Service().V1alpha1().PublicServiceTemplates().Lister())

	ctr := controller.NewServiceController(openappHelper)

	k8sFactory.Start(ctx.Done())
	openappFactory.Start(ctx.Done())

	klog.Infof("Wait resource cache sync...")
	if ok := cache.WaitForCacheSync(ctx.Done(),
		appInstanceInformer.HasSynced,
		appTemplateInformer.HasSynced,
		serviceInformer.HasSynced,
		serviceInstanceInformer.HasSynced,
		serviceTemplateInformer.HasSynced); !ok {
		klog.Fatal("Failed to wait for cache sync")
	}

	go ctr.Start()

	<-ctx.Done()
	return nil
}
