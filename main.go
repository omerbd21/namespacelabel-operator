/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"os"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/go-logr/zapr"
	ecszap "go.elastic.co/ecszap"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	danaiodanaiov1alpha1 "danaiodanaio/omerbd21/namespacelabel-operator/api/v1alpha1"
	"danaiodanaio/omerbd21/namespacelabel-operator/controllers"
	//+kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
	nllog  = logf.Log.WithName("namespacelabel-resource")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(danaiodanaiov1alpha1.AddToScheme(scheme))
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var protectedPrefixes string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&protectedPrefixes, "protected-Preixes", "", "All the protected label Prefixes you wish to block.")
	flag.Parse()

	encoderConfig := ecszap.NewDefaultEncoderConfig()
	core := ecszap.NewCore(encoderConfig, os.Stdout, zap.DebugLevel)
	log := zap.New(core, zap.AddCaller())
	logf.SetLogger(zapr.NewLogger(log))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "717a25f7.dana.io.dana.io",
	})
	if err != nil {
		log.Error("unable to start manager", zap.Error(err))
		os.Exit(1)
	}

	if err = (&controllers.NamespaceLabelReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		ProtectedPrefixes: protectedPrefixes,
	}).SetupWithManager(mgr); err != nil {
		log.Error("unable to create controller", zap.Error(err))
		os.Exit(1)
	}
	if err = (&danaiodanaiov1alpha1.NamespaceLabel{}).SetupWebhookWithManager(mgr); err != nil {
		nllog.Error(err, "unable to create webhook", "webhook", "NamespaceLabel")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Error("unable to set up health check", zap.Error(err))
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Error("unable to set up ready check", zap.Error(err))
		os.Exit(1)
	}

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error("problem running manager", zap.Error(err))
		os.Exit(1)
	}
}
