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
	"time"

	"github.com/omerbd21/namespacelabel-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("NamespaceLabel controller", func() {
	Context("NamespaceLabel controller test", func() {

		const NamespaceLabelName = "test-namespacelabel"

		ctx := context.Background()
		clientSet := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())
		/*namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: NamespaceLabelName,
			},
		}*/

		typeNamespaceName := types.NamespacedName{Name: NamespaceLabelName, Namespace: "default"}

		BeforeEach(func() {
			By("Creating the Namespace to perform the tests")
			//err := k8sClient.Create(ctx, namespace)
			//Expect(err).To(Not(HaveOccurred()))

		})

		AfterEach(func() {
			// TODO(user): Attention if you improve this code by adding other context test you MUST
			// be aware of the current delete namespace limitations. More info: https://book.kubebuilder.io/reference/envtest.html#testing-considerations
			//By("Deleting the Namespace to perform the tests")
			//_ = k8sClient.Delete(ctx, namespace)
		})

		It("should successfully reconcile a custom resource for NamespaceLabel", func() {
			By("Creating the custom resource for the Kind NamespaceLabel")
			namespaceLabel := &v1alpha1.NamespaceLabel{
				ObjectMeta: metav1.ObjectMeta{
					Name:      NamespaceLabelName,
					Namespace: "default",
				},
				Spec: v1alpha1.NamespaceLabelSpec{
					Labels: map[string]string{
						"label_1": "1",
					},
				},
			}

			err := k8sClient.Create(ctx, namespaceLabel)
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &v1alpha1.NamespaceLabel{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Reconciling the custom resource created")
			namespaceLabelReconciler := &NamespaceLabelReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = namespaceLabelReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})
			Expect(err).To(Not(HaveOccurred()))
			By("Checking if label was successfully assigned to the namespace in the reconciliation")
			namespace, err := clientSet.CoreV1().Namespaces().Get(ctx, namespaceLabel.Namespace, v1.GetOptions{})
			Expect(namespace.GetLabels()["label_1"] == "1").Should(BeTrue())

		})
	})
})
