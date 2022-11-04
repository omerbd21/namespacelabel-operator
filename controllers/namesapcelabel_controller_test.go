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

package controllers_test

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/omerbd21/namespacelabel-operator/api/v1alpha1"
	"github.com/omerbd21/namespacelabel-operator/controllers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const UpdatedSuffix = "updated"

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
			namespaceLabelReconciler := &controllers.NamespaceLabelReconciler{
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

			By("Reconciling the custom resource")
			namespaceLabelReconciler := &controllers.NamespaceLabelReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
			_, err = namespaceLabelReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})

			By("Creating the custom resource to update the first one")
			namespaceLabelUpdated := &v1alpha1.NamespaceLabel{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name + UpdatedSuffix,
					Namespace: typeNamespaceName.Namespace,
				},
				Spec: v1alpha1.NamespaceLabelSpec{
					Labels: map[string]string{
						"label_1": "2",
					},
				},
			}

			err = k8sClient.Create(ctx, namespaceLabelUpdated)
			Expect(err).To(Not(HaveOccurred()))

			By("Reconciling the custom resource")
			_, err = namespaceLabelReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: namespaceLabelUpdated.Name, Namespace: namespaceLabelUpdated.Namespace},
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
			namespaceLabelReconciler := &controllers.NamespaceLabelReconciler{
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
		It("should successfully use only the last values assgined to a label with 3 NamespaceLabels", func() {
			By("Creating the first custom resource for the Kind NamespaceLabel")
			namespaceLabel := &v1alpha1.NamespaceLabel{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: typeNamespaceName.Namespace,
				},
				Spec: v1alpha1.NamespaceLabelSpec{
					Labels: map[string]string{
						"label_1": "2022",
						"label_2": "2021",
						"label_3": "2011",
					},
				},
			}

			err := k8sClient.Create(ctx, namespaceLabel)
			Expect(err).To(Not(HaveOccurred()))

			By("Reconciling the first custom resource")
			namespaceLabelReconciler := &controllers.NamespaceLabelReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
			_, err = namespaceLabelReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})

			By("Creating the second custom resource for the Kind NamespaceLabel")
			namespaceLabel2 := &v1alpha1.NamespaceLabel{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name + "2",
					Namespace: typeNamespaceName.Namespace,
				},
				Spec: v1alpha1.NamespaceLabelSpec{
					Labels: map[string]string{
						"label_2": "2009",
						"label_3": "2006",
						"label_4": "2005",
					},
				},
			}

			err = k8sClient.Create(ctx, namespaceLabel2)
			Expect(err).To(Not(HaveOccurred()))

			By("Reconciling the second custom resource")
			_, err = namespaceLabelReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: namespaceLabel2.Name, Namespace: namespaceLabel2.Namespace},
			})

			By("Creating the third custom resource for the Kind NamespaceLabel")
			namespaceLabel3 := &v1alpha1.NamespaceLabel{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name + "3",
					Namespace: typeNamespaceName.Namespace,
				},
				Spec: v1alpha1.NamespaceLabelSpec{
					Labels: map[string]string{
						"label_3": "2004",
						"label_4": "2002",
						"label_5": "2001",
					},
				},
			}

			err = k8sClient.Create(ctx, namespaceLabel3)
			Expect(err).To(Not(HaveOccurred()))

			By("Reconciling the third custom resource")
			_, err = namespaceLabelReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: namespaceLabel3.Name, Namespace: namespaceLabel3.Namespace},
			})

			By("Checking if the label was successfully updated in the reconciliation")
			k8sClient.Get(ctx, types.NamespacedName{Name: typeNamespaceName.Namespace}, namespace)
			Expect(
				namespace.GetLabels()["label_1"] == "2022" &&
					namespace.GetLabels()["label_2"] == "2009" &&
					namespace.GetLabels()["label_3"] == "2004" &&
					namespace.GetLabels()["label_4"] == "2002" &&
					namespace.GetLabels()["label_5"] == "2001").Should(BeTrue())

		})
		It("should successfully still use the previous labels after deleting one of 3", func() {
			By("Creating the first custom resource for the Kind NamespaceLabel")
			namespaceLabel := &v1alpha1.NamespaceLabel{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: typeNamespaceName.Namespace,
				},
				Spec: v1alpha1.NamespaceLabelSpec{
					Labels: map[string]string{
						"label_1": "2022",
						"label_2": "2021",
						"label_3": "2011",
					},
				},
			}

			err := k8sClient.Create(ctx, namespaceLabel)
			Expect(err).To(Not(HaveOccurred()))

			By("Reconciling the first custom resource")
			namespaceLabelReconciler := &controllers.NamespaceLabelReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
			_, err = namespaceLabelReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})

			By("Creating the second custom resource for the Kind NamespaceLabel")
			namespaceLabel2 := &v1alpha1.NamespaceLabel{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name + "2",
					Namespace: typeNamespaceName.Namespace,
				},
				Spec: v1alpha1.NamespaceLabelSpec{
					Labels: map[string]string{
						"label_2": "2009",
						"label_3": "2006",
						"label_4": "2005",
					},
				},
			}

			err = k8sClient.Create(ctx, namespaceLabel2)
			Expect(err).To(Not(HaveOccurred()))

			By("Reconciling the second custom resource")
			_, err = namespaceLabelReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: namespaceLabel2.Name, Namespace: namespaceLabel2.Namespace},
			})

			By("Creating the third custom resource for the Kind NamespaceLabel")
			namespaceLabel3 := &v1alpha1.NamespaceLabel{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name + "3",
					Namespace: typeNamespaceName.Namespace,
				},
				Spec: v1alpha1.NamespaceLabelSpec{
					Labels: map[string]string{
						"label_3": "2004",
						"label_4": "2002",
						"label_5": "2001",
					},
				},
			}

			err = k8sClient.Create(ctx, namespaceLabel3)
			Expect(err).To(Not(HaveOccurred()))

			By("Reconciling the third custom resource")
			_, err = namespaceLabelReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: namespaceLabel3.Name, Namespace: namespaceLabel3.Namespace},
			})

			By("Deleting the third custom resource")
			err = k8sClient.Delete(ctx, namespaceLabel3)
			Expect(err).To(Not(HaveOccurred()))

			By("Reconciling deleting the third custom resource")
			_, err = namespaceLabelReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: namespaceLabel3.Name, Namespace: namespaceLabel3.Namespace},
			})

			By("Checking the previous labels exist")
			k8sClient.Get(ctx, types.NamespacedName{Name: typeNamespaceName.Namespace}, namespace)
			labels := namespace.GetLabels()
			Expect(labels["label_4"] == "2005" && labels["label_3"] == "2006").Should(BeTrue())

			By("Deleting the second custom resource")
			err = k8sClient.Delete(ctx, namespaceLabel2)
			Expect(err).To(Not(HaveOccurred()))

			By("Reconciling deleting the third custom resource")
			_, err = namespaceLabelReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: namespaceLabel2.Name, Namespace: namespaceLabel2.Namespace},
			})

			By("Checking the previous labels exist")
			k8sClient.Get(ctx, types.NamespacedName{Name: typeNamespaceName.Namespace}, namespace)
			labels = namespace.GetLabels()
			Expect(labels["label_2"] == "2021" && labels["label_3"] == "2011").Should(BeTrue())

		})
		It("should successfully protect the protected labels from being changed", func() {

			By("Getting the previous labels existing on the namespace")
			k8sClient.Get(ctx, types.NamespacedName{Name: typeNamespaceName.Namespace}, namespace)
			labels := namespace.GetLabels()

			By("Creating the custom resource for the Kind NamespaceLabel")
			namespaceLabel := &v1alpha1.NamespaceLabel{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: typeNamespaceName.Namespace,
				},
				Spec: v1alpha1.NamespaceLabelSpec{
					Labels: map[string]string{
						"kubernetes.io/metadata.name":       "name",
						controllers.AppLabel + "name":       "name",
						controllers.AppLabel + "instance":   "instance",
						controllers.AppLabel + "version":    "version",
						controllers.AppLabel + "component":  "component",
						controllers.AppLabel + "part-of":    "part-of",
						controllers.AppLabel + "managed-by": "managed-by",
					},
				},
			}
			err := k8sClient.Create(ctx, namespaceLabel)
			Expect(err).To(Not(HaveOccurred()))

			By("Reconciling the custom resource")
			namespaceLabelReconciler := &controllers.NamespaceLabelReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
			_, err = namespaceLabelReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})

			By("Checking the previous labels exist")
			k8sClient.Get(ctx, types.NamespacedName{Name: typeNamespaceName.Namespace}, namespace)
			newLabels := namespace.GetLabels()
			Expect(reflect.DeepEqual(labels, newLabels)).Should(BeTrue())

		})

	})
})
