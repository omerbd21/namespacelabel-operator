package internal

import (
	"context"
	danaiodanaiov1alpha1 "danaiodanaio/omerbd21/namespacelabel-operator/api/v1alpha1"
	"errors"
	"net/http"

	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-v1-namespacelabel,mutating=false,failurePolicy=fail,sideEffects=None,groups=dana.io.dana.io,resources=namespacelabels,verbs=create;update,versions=v1alpha1,name=vnamespacelabel.kb.io,admissionReviewVersions=v1

// NamespaceLabelValidator is  used to handle the webhook in the main file
type NamespaceLabelValidator struct {
	Client  client.Client
	Decoder *admission.Decoder
}

// Handle checks if the namespacelabel name is the same as the namespace
func (hook *NamespaceLabelValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := ctrl.Log.WithName("webhooks").WithName("InitInjector")
	namespaceLabel := &danaiodanaiov1alpha1.NamespaceLabel{}
	err := hook.Decoder.Decode(req, namespaceLabel)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if namespaceLabel.Name != namespaceLabel.Namespace {
		errMsg := "the namespacelabel has to be in the name of the namespace"
		err := errors.New(errMsg)
		log.Error(err, "namespacelabel error")
		return admission.Denied(errMsg)
	}

	return admission.Allowed("Namespacelabel created")
}
