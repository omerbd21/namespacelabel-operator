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

const AppLabel = "app.kubernetes.io/"

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
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NamespaceLabel object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
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

	myFinalizerName := "dana.io.dana.io/finalizer"
	if namespaceLabel.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(&namespaceLabel, myFinalizerName) {
			controllerutil.AddFinalizer(&namespaceLabel, myFinalizerName)
			if err := r.Update(ctx, &namespaceLabel); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(&namespaceLabel, myFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if _, err := r.deleteExternalResources(ctx, &namespaceLabel); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(&namespaceLabel, myFinalizerName)
			if err := r.Update(ctx, &namespaceLabel); err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

func (r *NamespaceLabelReconciler) deleteExternalResources(ctx context.Context, namespaceLabel *danaiodanaiov1alpha1.NamespaceLabel) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	clientSet := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())
	namespace, err := clientSet.CoreV1().Namespaces().Get(ctx, namespaceLabel.Namespace, v1.GetOptions{})
	if err != nil {
		log.Error(err, "unable to fetch namespace")
		return ctrl.Result{}, client.IgnoreNotFound(err)

	}
	var namespaceLabels danaiodanaiov1alpha1.NamespaceLabelList
	if err := r.List(ctx, &namespaceLabels); err != nil {
		log.Error(err, "unable to fetch NamespaceLabels")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	labels := namespace.GetLabels()
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
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceLabelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&danaiodanaiov1alpha1.NamespaceLabel{}).
		Complete(r)
}
