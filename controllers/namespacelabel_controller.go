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

package controllers

import (
	"context"
	"reflect"

	danaiodanaiov1alpha1 "github.com/omerbd21/namespacelabel-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// AppLabel is a constant the saves the App Label prefix
const AppLabel = "app.kubernetes.io/"

// getProtectedLabels returns a slice of all "protected" (system-used/application-used) labels
func getProtectedLabels() []string {
	return []string{"kubernetes.io/metadata.name",
		AppLabel + "name",
		AppLabel + "instance",
		AppLabel + "version",
		AppLabel + "component",
		AppLabel + "part-of",
		AppLabel + "managed-by",
	}
}

// contains gets a slice of strings and a string and returns whether the string is in the slice
func contains(strings []string, element string) bool {
	for _, str := range strings {
		if str == element {
			return true
		}
	}
	return false
}

// NamespaceLabelReconciler reconciles a NamespaceLabel object
type NamespaceLabelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dana.io.dana.io,resources=namespacelabels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dana.io.dana.io,resources=namespacelabels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dana.io.dana.io,resources=namespacelabels/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// This reconcile function adds the labels from the NamespaceLabel to the namespace it runs against,
// and deletes the labels when the resource is deleted.
func (r *NamespaceLabelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	clientSet := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())

	var namespaceLabel danaiodanaiov1alpha1.NamespaceLabel
	if err := r.Get(ctx, req.NamespacedName, &namespaceLabel); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch NamespaceLabel")
		return ctrl.Result{Requeue: true}, client.IgnoreNotFound(err)
	}
	namespace, err := clientSet.CoreV1().Namespaces().Get(ctx, namespaceLabel.Namespace, v1.GetOptions{})
	if err != nil {
		log.Error(err, "unable to fetch namespace")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	labels := namespace.GetLabels()
	protectedLabels := getProtectedLabels()
	for key, val := range namespaceLabel.Spec.Labels {
		if !contains(protectedLabels, key) {
			labels[key] = val
		}
	}

	namespace.SetLabels(labels)
	clientSet.CoreV1().Namespaces().Update(ctx, namespace, v1.UpdateOptions{})

	finalizerName := "dana.io.dana.io/finalizer"
	if namespaceLabel.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&namespaceLabel, finalizerName) {
			controllerutil.AddFinalizer(&namespaceLabel, finalizerName)
			if err := r.Update(ctx, &namespaceLabel); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(&namespaceLabel, finalizerName) {
			var namespaceLabels danaiodanaiov1alpha1.NamespaceLabelList
			if err := r.List(ctx, &namespaceLabels); err != nil {
				log.Error(err, "unable to fetch NamespaceLabels")
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
			for key := range namespaceLabel.Spec.Labels {
				delete(labels, key)
				for _, nlabel := range namespaceLabels.Items {
					if reflect.DeepEqual(nlabel.Spec.Labels, namespaceLabel.Spec.Labels) {
						continue
					} else if _, ok := nlabel.Spec.Labels[key]; ok {
						labels[key] = nlabel.Spec.Labels[key]
					}
				}
			}
			namespace.SetLabels(labels)
			clientSet.CoreV1().Namespaces().Update(ctx, namespace, v1.UpdateOptions{})

			controllerutil.RemoveFinalizer(&namespaceLabel, finalizerName)
			if err := r.Update(ctx, &namespaceLabel); err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceLabelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&danaiodanaiov1alpha1.NamespaceLabel{}).
		Complete(r)
}
