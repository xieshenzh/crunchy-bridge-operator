//go:build dbaas
// +build dbaas

package main

import (
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	dbaasredhatcomv1alpha1 "github.com/CrunchyData/crunchy-bridge-operator/apis/dbaas.redhat.com/v1alpha1"
	dbaasredhatcomv1alpha2 "github.com/CrunchyData/crunchy-bridge-operator/apis/dbaas.redhat.com/v1alpha2"
	dbaasredhatcomcontrollers "github.com/CrunchyData/crunchy-bridge-operator/controllers/dbaas.redhat.com"
	dbaasv1alpha1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	dbaasv1alpha2 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha2"
)

func init() {
	utilruntime.Must(dbaasredhatcomv1alpha1.AddToScheme(scheme))
	utilruntime.Must(dbaasredhatcomv1alpha2.AddToScheme(scheme))
	utilruntime.Must(dbaasv1alpha1.AddToScheme(scheme))
	utilruntime.Must(dbaasv1alpha2.AddToScheme(scheme))

	dbaasInit = enableDBaaSExtension
}

func enableDBaaSExtension(mgrOpts manager.Options, crunchybridgeAPIURL string) manager.Manager {
	mgrOpts.NewCache = cache.BuilderWithOptions(cache.Options{
		SelectorsByObject: cache.SelectorsByObject{
			&corev1.Secret{}: {
				Label: labels.SelectorFromSet(labels.Set{
					dbaasv1alpha1.TypeLabelKey: dbaasv1alpha1.TypeLabelValue,
				}),
			},
		},
	})
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), mgrOpts)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	clientset, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create clientset")
		os.Exit(1)
	}

	inventoryReconciler := &dbaasredhatcomcontrollers.CrunchyBridgeInventoryReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		APIBaseURL: crunchybridgeAPIURL,
		Log:        setupLog,
	}

	if err := inventoryReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CrunchyBridgeInventory")
		os.Exit(1)
	}

	dbaaSProviderReconciler := &dbaasredhatcomcontrollers.DBaaSProviderReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		Log:       setupLog,
		Clientset: clientset,
	}

	if err := dbaaSProviderReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DBaaSProvider")
		os.Exit(1)
	}

	if err := (&dbaasredhatcomcontrollers.CrunchyBridgeConnectionReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		Clientset:  clientset,
		APIBaseURL: crunchybridgeAPIURL,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CrunchyBridgeConnection")
		os.Exit(1)
	}

	if err := (&dbaasredhatcomcontrollers.CrunchyBridgeInstanceReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		APIBaseURL: crunchybridgeAPIURL,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CrunchyBridgeInstance")
		os.Exit(1)
	}

	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		if err = (&dbaasredhatcomv1alpha1.CrunchyBridgeInventory{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "CrunchyBridgeInventory")
			os.Exit(1)
		}
		if err = (&dbaasredhatcomv1alpha1.CrunchyBridgeConnection{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "CrunchyBridgeConnection")
			os.Exit(1)
		}
		if err = (&dbaasredhatcomv1alpha2.CrunchyBridgeInventory{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "CrunchyBridgeInventory")
			os.Exit(1)
		}
		if err = (&dbaasredhatcomv1alpha2.CrunchyBridgeConnection{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "CrunchyBridgeConnection")
			os.Exit(1)
		}
	}

	return mgr
}
