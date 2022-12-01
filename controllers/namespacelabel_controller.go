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
	"fmt"
	"strings"
	"time"

	danaiodanaiov1alpha1 "danaiodanaio/omerbd21/namespacelabel-operator/api/v1alpha1"
	utils "danaiodanaio/omerbd21/namespacelabel-operator/utils"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// AppLabel is a constant the saves the App Label prefix
const AppLabel = "app.kubernetes.io/"
const FinalizerName = "dana.io.dana.io/finalizer"

// getProtectedPrefixes returns a slice of all "protected" (system-used/application-used) prefixes of labels
func getProtectedPrefixes(Prefixes string) []string {
	split := strings.Split(Prefixes, ",")
	return split
}

// NamespaceLabelReconciler reconciles a NamespaceLabel object
type NamespaceLabelReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ProtectedPrefixes string
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
	log := ctrllog.FromContext(ctx)
	var namespaceLabel danaiodanaiov1alpha1.NamespaceLabel
	if err := r.Get(ctx, req.NamespacedName, &namespaceLabel); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to fetch NamespaceLabel", "namespaceLabel", req.NamespacedName.String())
		return ctrl.Result{Requeue: true}, client.IgnoreNotFound(err)
	}

	var namespace corev1.Namespace
	if err := r.Client.Get(ctx, types.NamespacedName{Name: namespaceLabel.ObjectMeta.Namespace}, &namespace); err != nil {
		log.Error(err, "Unable to fetch namespace", "namespace", namespaceLabel.ObjectMeta.Namespace)
		return ctrl.Result{Requeue: true}, client.IgnoreNotFound(err)
	}
	if namespaceLabel.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&namespaceLabel, FinalizerName) {
			if res, err := r.addFinalizer(ctx, log, &namespaceLabel, &namespace); err != nil {
				return res, err
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(&namespaceLabel, FinalizerName) {
			if res, err := r.deleteLabels(ctx, log, &namespaceLabel, &namespace); err != nil {
				return res, err
			}
		}
	}
	if res, err := r.updateLabels(ctx, log, &namespaceLabel, &namespace); err != nil {
		return res, err
	}
	return ctrl.Result{}, nil
}

func (r *NamespaceLabelReconciler) updateLabels(ctx context.Context, log logr.Logger, namespaceLabel *danaiodanaiov1alpha1.NamespaceLabel, namespace *corev1.Namespace) (ctrl.Result, error) {
	labels := namespace.GetLabels()
	protectedPrefixes := getProtectedPrefixes(r.ProtectedPrefixes)
	for key, val := range namespaceLabel.Spec.Labels {
		_, exists := labels[key]
		if !utils.Contains(protectedPrefixes, strings.Split(key, "/")[0]) && !exists {
			labels[key] = val
			condition := metav1.Condition{Type: "LabelApplied", Status: "True", Reason: "Label Applied", Message: "Label" + key + " = " + val + "was applied", LastTransitionTime: metav1.Time{Time: time.Now()}}
			namespaceLabel.Status.Conditions = append(namespaceLabel.Status.Conditions, condition)
			namespaceLabel.Status.EnforcedLabels = append(namespaceLabel.Status.EnforcedLabels, key)
		} else if exists && !utils.Contains(namespaceLabel.Status.EnforcedLabels, key) {
			condition := metav1.Condition{Type: "LabelApplied", Status: "False", Reason: "LabelNotApplied", Message: "Label was not applied because a label with the same name was applied earlier", LastTransitionTime: metav1.Time{Time: time.Now()}}
			namespaceLabel.Status.Conditions = append(namespaceLabel.Status.Conditions, condition)
			delete(namespaceLabel.Spec.Labels, key)
		} // Checks if the label is protected before adding it to the namespace
	}
	log.Info("labels were put on the namespace", "namespacelabel", fmt.Sprint(namespaceLabel.Name), "labels", fmt.Sprint(labels))
	namespace.SetLabels(labels)
	if err := r.Client.Update(ctx, namespace); err != nil {
		log.Error(err, "unable to fetch namespace while updating new labels", "namespace", namespace.Name, "labels", fmt.Sprint(labels))
		return ctrl.Result{Requeue: true}, client.IgnoreNotFound(err)
	}
	return ctrl.Result{}, nil
}

func (r *NamespaceLabelReconciler) deleteLabels(ctx context.Context, log logr.Logger, namespaceLabel *danaiodanaiov1alpha1.NamespaceLabel, namespace *corev1.Namespace) (ctrl.Result, error) {
	labels := namespace.GetLabels()
	for label, _ := range namespaceLabel.Spec.Labels {
		delete(labels, label)
	}
	namespace.SetLabels(labels)
	if err := r.Client.Update(ctx, namespace); err != nil {
		log.Error(err, "unable to fetch namespace while trying to update its labels post deletion", "namespace", namespace.Name)
		return ctrl.Result{Requeue: true}, client.IgnoreNotFound(err)
	}
	controllerutil.RemoveFinalizer(namespaceLabel, FinalizerName)
	if err := r.Update(ctx, namespaceLabel); err != nil {
		log.Error(err, "unable to update namespacelabel in order to remove finalizer", "namespacelabel", namespaceLabel.Name)
		return ctrl.Result{}, err
	}
	return ctrl.Result{Requeue: true}, nil
}

func (r *NamespaceLabelReconciler) addFinalizer(ctx context.Context, log logr.Logger, namespaceLabel *danaiodanaiov1alpha1.NamespaceLabel, namespace *corev1.Namespace) (ctrl.Result, error) {
	labels := namespace.GetLabels()
	controllerutil.AddFinalizer(namespaceLabel, FinalizerName)
	if err := r.Update(ctx, namespaceLabel); err != nil {
		log.Error(err, "unable to add finalizer to namespacelabel", "namespacelabel", namespaceLabel.Name, "labels", fmt.Sprint(labels))
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceLabelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&danaiodanaiov1alpha1.NamespaceLabel{}).
		Complete(r)
}
