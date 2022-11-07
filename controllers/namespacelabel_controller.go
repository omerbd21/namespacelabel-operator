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
	"errors"
	"reflect"

	danaiodanaiov1alpha1 "danaiodanaio/omerbd21/namespacelabel-operator/api/v1alpha1"
	utils "danaiodanaio/omerbd21/namespacelabel-operator/utils"
	"github.com/sirupsen/logrus"
	"go.elastic.co/ecslogrus"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

// NamespaceLabelReconciler reconciles a NamespaceLabel object
type NamespaceLabelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dana.io.dana.io,resources=namespacelabels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dana.io.dana.io,resources=namespacelabels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dana.io.dana.io,resources=namespacelabels/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// This reconcile function adds the labels from the NamespaceLabel to the namespace it runs against,
// and deletes the labels when the resource is deleted.
func (r *NamespaceLabelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logrus.New()
	log.SetFormatter(&ecslogrus.Formatter{})

	var namespaceLabel danaiodanaiov1alpha1.NamespaceLabel
	if err := r.Get(ctx, req.NamespacedName, &namespaceLabel); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.WithError(errors.New(err.Error())).WithFields(logrus.Fields{"NamespaceLabel": req.NamespacedName}).Error("unable to fetch NamespaceLabel")
		return ctrl.Result{Requeue: true}, client.IgnoreNotFound(err)
	}

	var namespace corev1.Namespace
	if err := r.Client.Get(ctx, types.NamespacedName{Name: namespaceLabel.ObjectMeta.Namespace}, &namespace); err != nil {
		log.WithError(errors.New(err.Error())).WithFields(logrus.Fields{"namespace": namespaceLabel.ObjectMeta.Namespace}).Error("unable to fetch namespace while getting previous labels")
		return ctrl.Result{Requeue: true}, client.IgnoreNotFound(err)
	}

	labels := namespace.GetLabels()
	protectedLabels := getProtectedLabels()
	for key, val := range namespaceLabel.Spec.Labels {
		if !utils.Contains(protectedLabels, key) {
			labels[key] = val
		} // Checks if the label is protected before adding it to the namespace
	}
	log.WithFields(logrus.Fields{"labels": labels, "namespace": namespace.Name}).Info("labels were put on the namesapce")
	namespace.SetLabels(labels)

	if err := r.Client.Update(ctx, &namespace); err != nil {
		log.WithError(errors.New(err.Error())).WithFields(logrus.Fields{"namespace": namespace.Name, "labels": labels}).Error("unable to fetch namespace while updating new labels")
		return ctrl.Result{Requeue: true}, client.IgnoreNotFound(err)
	}

	finalizerName := "dana.io.dana.io/finalizer"
	if namespaceLabel.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&namespaceLabel, finalizerName) {
			controllerutil.AddFinalizer(&namespaceLabel, finalizerName)
			if err := r.Update(ctx, &namespaceLabel); err != nil {
				log.WithError(errors.New(err.Error())).WithFields(logrus.Fields{"namespacelabel": namespaceLabel.Name}).Error("unable to add finalizer to namespacelabel")
				return ctrl.Result{}, err
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(&namespaceLabel, finalizerName) {
			var namespaceLabels danaiodanaiov1alpha1.NamespaceLabelList
			if err := r.List(ctx, &namespaceLabels); err != nil {
				log.WithError(errors.New(err.Error())).WithFields(logrus.Fields{"deleted_namespacelabel": namespaceLabel.Name}).Error("unable to fetch NamespaceLabels while trying to delete one of them")
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}

			labels = map[string]string{}
			for _, nlabel := range namespaceLabels.Items {
				if reflect.DeepEqual(nlabel.Spec.Labels, namespaceLabel.Spec.Labels) {
					continue
				} // If the NamespaceLabel is the current one, skip it and don't add it to the labels map
				for key, val := range nlabel.Spec.Labels {
					labels[key] = val
				}
			}
			namespace.SetLabels(labels)
			if err := r.Client.Update(ctx, &namespace); err != nil {
				log.WithError(errors.New(err.Error())).WithFields(logrus.Fields{"namespace": namespace.Name}).Error("unable to fetch namespace while trying to update its labels post deletion")
				return ctrl.Result{Requeue: true}, client.IgnoreNotFound(err)
			}

			controllerutil.RemoveFinalizer(&namespaceLabel, finalizerName)
			if err := r.Update(ctx, &namespaceLabel); err != nil {
				log.WithError(errors.New(err.Error())).WithFields(logrus.Fields{"namespacelabel": namespaceLabel.Name}).Error("unable to update namespacelabel in order to remove finalizer")
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
