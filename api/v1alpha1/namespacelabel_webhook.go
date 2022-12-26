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

package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var namespacelabellog = logf.Log.WithName("namespacelabel-resource")

func (r *NamespaceLabel) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

func (r *NamespaceLabel) validateNamespaceLabelName() *field.Error {
	// The field helpers from the kubernetes API machinery help us return nicely
	// structured validation errors.
	if r.Name != r.Namespace {
		return field.Invalid(field.NewPath("name"), r.Name, "The name you put is invalid; you have to put the name of the namespace")
	}
	return nil
}

func (r *NamespaceLabel) validateNamespaceLabel() error {
	if err := r.validateNamespaceLabelName(); err != nil {
		return apierrors.NewInvalid(
			schema.GroupKind{Group: "dana.io.dana.io", Kind: "NamespaceLabel"},
			r.Name, field.ErrorList{err})
	}
	return nil

}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-dana-io-dana-io-v1alpha1-namespacelabel,mutating=false,failurePolicy=fail,sideEffects=None,groups=dana.io.dana.io,resources=namespacelabels,verbs=create;update,versions=v1alpha1,name=vnamespacelabel.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &NamespaceLabel{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *NamespaceLabel) ValidateCreate() error {
	namespacelabellog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return r.validateNamespaceLabel()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *NamespaceLabel) ValidateUpdate(old runtime.Object) error {
	namespacelabellog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return r.validateNamespaceLabel()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *NamespaceLabel) ValidateDelete() error {
	namespacelabellog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
