package controller

import (
	"context"
	"net"
	"os"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"

	"github.com/openapp-dev/openapp/pkg/generated/clientset/versioned"
	openapputils "github.com/openapp-dev/openapp/pkg/utils"
	"github.com/openapp-dev/publicservice/frp4/pkg/utils"
)

type ServiceController struct {
	serviceClass  string
	k8sClient     kubernetes.Interface
	openappClient versioned.Interface
	workqueue     *openapputils.WorkQueue
}

func NewServiceController(openappHelper *openapputils.OpenAPPHelper) *ServiceController {
	sc := &ServiceController{
		serviceClass: os.Getenv("SERVICE_CLASS"),
	}
	sc.workqueue = openapputils.NewWorkQueue(sc.Reconcile)
	sc.openappClient = openappHelper.OpenAPPClient
	sc.k8sClient = openappHelper.K8sClient

	openappHelper.ServiceInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			svc := obj.(*corev1.Service)
			if svc.Labels == nil {
				return false
			}
			return svc.Namespace == openapputils.InstanceNamespace &&
				sc.serviceClass == svc.Labels[openapputils.ServiceExposeClassLabelKey]
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				sc.workqueue.Add(obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				oldAppIns := oldObj.(*corev1.Service)
				newAppIns := newObj.(*corev1.Service)
				if reflect.DeepEqual(oldAppIns.Spec, newAppIns.Spec) {
					return
				}
				sc.workqueue.Add(newObj)
			},
			DeleteFunc: func(obj interface{}) {
				sc.workqueue.Add(obj)
			},
		},
	})

	return sc
}

func (sc *ServiceController) Start() {
	go sc.workqueue.Run()
}

func (sc *ServiceController) Reconcile(obj interface{}) error {
	klog.Infof("Reconciling app service...")

	svc, ok := obj.(*corev1.Service)
	if !ok {
		klog.Errorf("Failed to convert object to service")
		return nil
	}

	svcExist, err := sc.k8sClient.CoreV1().Services(svc.Namespace).
		Get(context.Background(), svc.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return utils.DeleteProxy(svc)
		}
		klog.Errorf("Failed to list services: %v", err)
		return err
	}

	url, port, updated, err := utils.AddOrUpdateProxy(svc)
	if err != nil {
		return err
	}
	if !updated {
		return nil
	}

	svcCopy := svcExist.DeepCopy()
	svcCopy.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{
		{
			Ports: []corev1.PortStatus{
				{Port: int32(port), Protocol: svc.Spec.Ports[0].Protocol},
			},
		},
	}

	if net.ParseIP(url) == nil {
		svcCopy.Status.LoadBalancer.Ingress[0].Hostname = url
	} else {
		svcCopy.Status.LoadBalancer.Ingress[0].IP = url
	}

	_, err = sc.k8sClient.CoreV1().Services(svc.Namespace).UpdateStatus(context.Background(), svcCopy, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("Failed to update service status: %v", err)
		return err
	}
	return nil
}
