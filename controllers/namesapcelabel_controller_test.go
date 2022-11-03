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
	"time"

	"github.com/omerbd21/namespacelabel-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("NamespaceLabel controller", func() {
	Context("NamespaceLabel controller test", func() {

		const NamespaceLabelName = "test-namespacelabel"

		ctx := context.Background()
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: NamespaceLabelName,
			},
		}
		testCounter := 0
		typeNamespaceName := types.NamespacedName{Name: NamespaceLabelName, Namespace: NamespaceLabelName}

		BeforeEach(func() {
			testCounter++
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: NamespaceLabelName + fmt.Sprint(testCounter),
				},
			}
			err := k8sClient.Create(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))
			typeNamespaceName.Namespace = NamespaceLabelName + fmt.Sprint(testCounter)
		})

		AfterEach(func() {
			//err := k8sClient.Delete(ctx, namespace)
			//Expect(err).To(Not(HaveOccurred()))
		})

		It("should successfully reconcile a custom resource for NamespaceLabel", func() {
			By("Creating the custom resource for the Kind NamespaceLabel")
			namespaceLabel := &v1alpha1.NamespaceLabel{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: typeNamespaceName.Namespace,
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
			k8sClient.Get(ctx, types.NamespacedName{Name: typeNamespaceName.Namespace}, namespace)
			Expect(namespace.GetLabels()["label_1"] == "1").Should(BeTrue())

		})
		It("should successfully update a custom resource for NamespaceLabel", func() {
			By("Setting namespace labels beforehand")
			namespace.SetLabels(map[string]string{
				"label_1": "1",
			})
			By("Creating the custom resource for the Kind NamespaceLabel")
			namespaceLabel := &v1alpha1.NamespaceLabel{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: typeNamespaceName.Namespace,
				},
				Spec: v1alpha1.NamespaceLabelSpec{
					Labels: map[string]string{
						"label_1": "2",
					},
				},
			}

			err := k8sClient.Create(ctx, namespaceLabel)
			Expect(err).To(Not(HaveOccurred()))

			By("Reconciling the custom resource")
			namespaceLabelReconciler := &NamespaceLabelReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
			_, err = namespaceLabelReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})

			By("Checking if the labek was successfully updated in the reconciliation")
			k8sClient.Get(ctx, types.NamespacedName{Name: typeNamespaceName.Namespace}, namespace)
			Expect(namespace.GetLabels()["label_1"] == "2").Should(BeTrue())

		})
		It("should successfully delete a custom resource for NamespaceLabel", func() {
			By("Creating the custom resource for the Kind NamespaceLabel")
			namespaceLabel := &v1alpha1.NamespaceLabel{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: typeNamespaceName.Namespace,
				},
				Spec: v1alpha1.NamespaceLabelSpec{
					Labels: map[string]string{
						"label_1": "2",
					},
				},
			}

			err := k8sClient.Create(ctx, namespaceLabel)
			Expect(err).To(Not(HaveOccurred()))

			By("Deleting the custom resource for the Kind NamespaceLabel")
			err = k8sClient.Delete(ctx, namespaceLabel)
			Expect(err).To(Not(HaveOccurred()))

			By("Reconciling the custom resource")
			namespaceLabelReconciler := &NamespaceLabelReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
			_, err = namespaceLabelReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})

			By("Checking if the labek was successfully updated in the reconciliation")
			k8sClient.Get(ctx, types.NamespacedName{Name: typeNamespaceName.Namespace}, namespace)
			_, ok := namespace.GetLabels()["label_1"]
			Expect(ok).Should(BeFalse())

		})
	})
})
