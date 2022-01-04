package k8s

import (
	"context"
	"strings"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/pkg/errors"
	crdv1alpha1 "github.com/sentinel-group/sentinel-go-datasource-k8s-crd/api/v1alpha1"
	"github.com/sentinel-group/sentinel-go-datasource-k8s-crd/controllers"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = crdv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

type CRDType int32

const (
	FlowRulesCRD CRDType = iota
	IsolationRulesCRD
	CircuitBreakerRulesCRD
	HotspotRulesCRD
	SystemRulesCRD
)

func (c CRDType) String() string {
	switch c {
	case FlowRulesCRD:
		return "FlowRulesCRD"
	case CircuitBreakerRulesCRD:
		return "CircuitBreakerRulesCRD"
	case HotspotRulesCRD:
		return "HotspotRulesCRD"
	case SystemRulesCRD:
		return "SystemRulesCRD"
	default:
		return "Undefined"
	}
}

type DataSource struct {
	crdManager  ctrl.Manager
	controllers map[CRDType]reconcile.Reconciler
	namespace   string

	ctx       context.Context
	ctxCancel context.CancelFunc
}

// NewDataSource creates a K8S DataSource with given namespace
// All Controllers take effective only when match namespace.
func NewDataSource(namespace string) (*DataSource, error) {
	ctrl.SetLogger(&k8SLogger{
		l:             logging.GetGlobalLogger(),
		level:         logging.GetGlobalLoggerLevel(),
		names:         make([]string, 0),
		keysAndValues: make([]interface{}, 0),
	})
	k8sConfig, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}
	mgr, err := ctrl.NewManager(k8sConfig, ctrl.Options{
		Scheme: scheme,
		// disable metric server
		MetricsBindAddress:     "0",
		HealthProbeBindAddress: "0",
		LeaderElection:         false,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	k := &DataSource{
		crdManager:  mgr,
		controllers: make(map[CRDType]reconcile.Reconciler, 4),
		namespace:   namespace,
		ctx:         ctx,
		ctxCancel:   cancel,
	}
	return k, nil
}

// RegisterController register given type crd and crd name
// For each type CRD can only register once.
func (k *DataSource) RegisterController(crd CRDType, crName string) error {
	if len(strings.TrimSpace(crName)) == 0 {
		return errors.New("empty crd name")
	}

	_, exist := k.controllers[crd]
	if exist {
		return errors.Errorf("duplicated register crd for %s", crd.String())
	}

	switch crd {
	case FlowRulesCRD:
		controller := &controllers.FlowRulesReconciler{
			Client:         k.crdManager.GetClient(),
			Logger:         ctrl.Log.WithName("controllers").WithName("FlowRules"),
			Scheme:         k.crdManager.GetScheme(),
			Namespace:      k.namespace,
			ExpectedCrName: crName,
		}
		err := controller.SetupWithManager(k.crdManager)
		if err != nil {
			return err
		}
		k.controllers[FlowRulesCRD] = controller
		setupLog.Info("Sentinel FlowRules CRD controller has been registered successfully")
		return nil
	case IsolationRulesCRD:
		controller := &controllers.IsolationRulesReconciler{
			Client:         k.crdManager.GetClient(),
			Logger:         ctrl.Log.WithName("controllers").WithName("IsolationRules"),
			Scheme:         k.crdManager.GetScheme(),
			Namespace:      k.namespace,
			ExpectedCrName: crName,
		}
		err := controller.SetupWithManager(k.crdManager)
		if err != nil {
			return err
		}
		k.controllers[IsolationRulesCRD] = controller
		setupLog.Info("Sentinel IsolationRules CRD controller has been registered successfully")
		return nil
	case CircuitBreakerRulesCRD:
		controller := &controllers.CircuitBreakerRulesReconciler{
			Client:         k.crdManager.GetClient(),
			Logger:         ctrl.Log.WithName("controllers").WithName("CircuitBreakerRules"),
			Scheme:         k.crdManager.GetScheme(),
			Namespace:      k.namespace,
			ExpectedCrName: crName,
		}
		err := controller.SetupWithManager(k.crdManager)
		if err != nil {
			return err
		}
		k.controllers[CircuitBreakerRulesCRD] = controller
		setupLog.Info("Sentinel CircuitBreakerRules CRD controller has been registered successfully")
		return nil
	case HotspotRulesCRD:
		controller := &controllers.HotspotRulesReconciler{
			Client:         k.crdManager.GetClient(),
			Logger:         ctrl.Log.WithName("controllers").WithName("HotspotRules"),
			Scheme:         k.crdManager.GetScheme(),
			Namespace:      k.namespace,
			ExpectedCrName: crName,
		}
		err := controller.SetupWithManager(k.crdManager)
		if err != nil {
			return err
		}
		k.controllers[HotspotRulesCRD] = controller
		setupLog.Info("Sentinel HotspotRules CRD controller has been registered successfully")
		return nil
	case SystemRulesCRD:
		controller := &controllers.SystemRulesReconciler{
			Client:         k.crdManager.GetClient(),
			Logger:         ctrl.Log.WithName("controllers").WithName("SystemRules"),
			Scheme:         k.crdManager.GetScheme(),
			Namespace:      k.namespace,
			ExpectedCrName: crName,
		}
		err := controller.SetupWithManager(k.crdManager)
		if err != nil {
			return err
		}
		k.controllers[SystemRulesCRD] = controller
		setupLog.Info("Sentinel SystemRules CRD controller has been registered successfully")
		return nil
	default:
		return errors.Errorf("unsupported CRDType: %d", int(crd))
	}
}

// Close exit the K8S DataSource
func (k *DataSource) Close() error {
	k.ctxCancel()
	return nil
}

// Run runs the k8s DataSource
func (k *DataSource) Run() error {
	// +kubebuilder:scaffold:builder
	go util.RunWithRecover(func() {
		setupLog.Info("Starting Sentinel CRD manager")
		if err := k.crdManager.Start(k.ctx); err != nil {
			setupLog.Error(err, "problem running Sentinel CRD manager")
		}
		setupLog.Info("Sentinel CRD data-source closed")
	})
	return nil
}
