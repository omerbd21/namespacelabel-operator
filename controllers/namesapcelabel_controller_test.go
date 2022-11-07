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
	"errors"
	"fmt"
	"reflect"
	"time"

	"danaiodanaio/omerbd21/namespacelabel-operator/api/v1alpha1"
	"danaiodanaio/omerbd21/namespacelabel-operator/controllers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const UpdatedSuffix = "updated"

// time consts for Eventually function
const timeout = time.Second * 10
const interval = time.Millisecond * 250

// Enum-like consts for the changeNamespaceLabelState function
const (
	Create string = "create"
	Delete        = "delete"
)

// Returns the basic namespaceLabelReconciler
func getNamespaceLabelReconciler() *controllers.NamespaceLabelReconciler {
	namespaceLabelReconciler := &controllers.NamespaceLabelReconciler{
		Client: k8sClient,
		Scheme: k8sClient.Scheme(),
	}
	return namespaceLabelReconciler
}

// Returns the labels for each Test Case
func getNamespaceLabelForTest(labelType string) map[string]string {
	namespaceLabels := map[string]map[string]string{
		"basic":        {"label_1": "1"},
		"basicUpdated": {"label_1": "2"},
		"complex1": {"label_1": "2022",
			"label_2": "2021",
			"label_3": "2011",
		},
		"complex2": {"label_2": "2009",
			"label_3": "2006",
			"label_4": "2005",
		},
		"complex3": {"label_3": "2004",
			"label_4": "2002",
			"label_5": "2001",
		},
		"protectedLabels": {"kubernetes.io/metadata.name": "name",
			controllers.AppLabel + "name":       "name",
			controllers.AppLabel + "instance":   "instance",
			controllers.AppLabel + "version":    "version",
			controllers.AppLabel + "component":  "component",
			controllers.AppLabel + "part-of":    "part-of",
			controllers.AppLabel + "managed-by": "managed-by"},
	}
	return namespaceLabels[labelType]
}

var _ = Describe("NamespaceLabel controller", func() {
	Context("NamespaceLabel controller test", func() {

		const NamespaceLabelName = "test-namespacelabel"

		ctx := context.Background()
		namespace := &corev1.Namespace{}
		testCounter := 0
		typeNamespaceName := types.NamespacedName{Name: NamespaceLabelName, Namespace: NamespaceLabelName}

		BeforeEach(func() {
			testCounter++
			// Each test case gets its own namespace
			namespace = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: NamespaceLabelName + fmt.Sprint(testCounter),
				},
			}
			err := k8sClient.Create(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))
			typeNamespaceName.Namespace = NamespaceLabelName + fmt.Sprint(testCounter)
		})

		AfterEach(func() {
			err := k8sClient.Delete(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should successfully reconcile a custom resource for NamespaceLabel", func() {
			By("Creating the custom resource for the Kind NamespaceLabel")
			err := changeNamespaceLabelState(getNamespaceLabelForTest("basic"), ctx, typeNamespaceName.Name, typeNamespaceName.Namespace, Create)
			Expect(err).Should(BeNil())

			By("Reconciling the custom resource created")
			_, err = reconcileResource(*getNamespaceLabelReconciler(), ctx, typeNamespaceName.Name, typeNamespaceName.Namespace, true)
			Expect(err).Should(BeNil())

			By("Checking if label was successfully assigned to the namespace in the reconciliation")
			k8sClient.Get(ctx, types.NamespacedName{Name: typeNamespaceName.Namespace}, namespace)
			Expect(namespace.GetLabels()["label_1"] == "1").Should(BeTrue())

		})
		It("should successfully update a custom resource for NamespaceLabel", func() {
			By("Creating the custom resource for the Kind NamespaceLabel")
			err := changeNamespaceLabelState(getNamespaceLabelForTest("basic"), ctx, typeNamespaceName.Name, typeNamespaceName.Namespace, Create)
			Expect(err).Should(BeNil())

			By("Reconciling the custom resource created")
			namespaceLabelReconciler := *getNamespaceLabelReconciler()
			_, err = reconcileResource(namespaceLabelReconciler, ctx, typeNamespaceName.Name, typeNamespaceName.Namespace, true)
			Expect(err).Should(BeNil())

			By("Creating the custom resource to update the first one")
			err = changeNamespaceLabelState(getNamespaceLabelForTest("basicUpdated"), ctx, typeNamespaceName.Name+"2", typeNamespaceName.Namespace, Create)
			Expect(err).Should(BeNil())

			By("Reconciling the custom resource updated")
			_, err = reconcileResource(namespaceLabelReconciler, ctx, typeNamespaceName.Name+"2", typeNamespaceName.Namespace, true)
			Expect(err).Should(BeNil())

			By("Checking if the label was successfully updated in the reconciliation")
			k8sClient.Get(ctx, types.NamespacedName{Name: typeNamespaceName.Namespace}, namespace)
			Expect(namespace.GetLabels()["label_1"] == "2").Should(BeTrue())

		})
		It("should successfully delete a custom resource for NamespaceLabel", func() {
			By("Creating the custom resource for the Kind NamespaceLabel")
			err := changeNamespaceLabelState(getNamespaceLabelForTest("basicUpdated"), ctx, typeNamespaceName.Name, typeNamespaceName.Namespace, Create)
			Expect(err).Should(BeNil())

			By("Deleting the custom resource for the Kind NamespaceLabel")
			err = changeNamespaceLabelState(getNamespaceLabelForTest("basicUpdated"), ctx, typeNamespaceName.Name, typeNamespaceName.Namespace, Delete)
			Expect(err).Should(BeNil())

			By("Reconciling the custom resource deleted")
			_, err = reconcileResource(*getNamespaceLabelReconciler(), ctx, typeNamespaceName.Name, typeNamespaceName.Namespace, false)
			Expect(err).Should(BeNil())

			By("Checking if the label was successfully deleted in the reconciliation")
			k8sClient.Get(ctx, types.NamespacedName{Name: typeNamespaceName.Namespace}, namespace)
			Expect(namespace.GetLabels()["label_1"] == "2").Should(BeFalse())

		})
		It("should successfully use only the last values assgined to a label with 3 NamespaceLabels", func() {
			labels1 := getNamespaceLabelForTest("complex1")
			labels2 := getNamespaceLabelForTest("complex2")
			labels3 := getNamespaceLabelForTest("complex3")

			By("Creating the first custom resource for the Kind NamespaceLabel")
			err := changeNamespaceLabelState(labels1, ctx, typeNamespaceName.Name, typeNamespaceName.Namespace, Create)
			Expect(err).Should(BeNil())

			By("Reconciling the first custom resource created")
			namespaceLabelReconciler := *getNamespaceLabelReconciler()
			_, err = reconcileResource(namespaceLabelReconciler, ctx, typeNamespaceName.Name, typeNamespaceName.Namespace, true)
			Expect(err).Should(BeNil())

			By("Creating the second custom resource for the Kind NamespaceLabel")
			err = changeNamespaceLabelState(labels2, ctx, typeNamespaceName.Name+"2", typeNamespaceName.Namespace, Create)
			Expect(err).Should(BeNil())

			By("Reconciling the second custom resource created")
			_, err = reconcileResource(namespaceLabelReconciler, ctx, typeNamespaceName.Name+"2", typeNamespaceName.Namespace, true)
			Expect(err).Should(BeNil())

			By("Creating the third custom resource for the Kind NamespaceLabel")
			err = changeNamespaceLabelState(labels3, ctx, typeNamespaceName.Name+"3", typeNamespaceName.Namespace, Create)
			Expect(err).Should(BeNil())

			By("Reconciling the third custom resource created")
			_, err = reconcileResource(namespaceLabelReconciler, ctx, typeNamespaceName.Name+"3", typeNamespaceName.Namespace, true)
			Expect(err).Should(BeNil())

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
			labels1 := getNamespaceLabelForTest("complex1")
			labels2 := getNamespaceLabelForTest("complex2")
			labels3 := getNamespaceLabelForTest("complex3")
			err := changeNamespaceLabelState(labels1, ctx, typeNamespaceName.Name, typeNamespaceName.Namespace, Create)
			Expect(err).Should(BeNil())

			By("Reconciling the first custom resource created")
			namespaceLabelReconciler := *getNamespaceLabelReconciler()
			_, err = reconcileResource(namespaceLabelReconciler, ctx, typeNamespaceName.Name, typeNamespaceName.Namespace, true)
			Expect(err).Should(BeNil())

			By("Creating the second custom resource for the Kind NamespaceLabel")
			err = changeNamespaceLabelState(labels2, ctx, typeNamespaceName.Name+"2", typeNamespaceName.Namespace, Create)
			Expect(err).Should(BeNil())

			By("Reconciling the second custom resource created")
			_, err = reconcileResource(namespaceLabelReconciler, ctx, typeNamespaceName.Name+"2", typeNamespaceName.Namespace, true)
			Expect(err).Should(BeNil())

			By("Creating the third custom resource for the Kind NamespaceLabel")
			err = changeNamespaceLabelState(labels3, ctx, typeNamespaceName.Name+"3", typeNamespaceName.Namespace, Create)
			Expect(err).Should(BeNil())

			By("Reconciling the third custom resource created")
			_, err = reconcileResource(namespaceLabelReconciler, ctx, typeNamespaceName.Name+"3", typeNamespaceName.Namespace, true)
			Expect(err).Should(BeNil())

			By("Deleting the third custom resource")
			err = changeNamespaceLabelState(labels3, ctx, typeNamespaceName.Name+"3", typeNamespaceName.Namespace, Delete)
			Expect(err).Should(BeNil())

			By("Reconciling deleting the third custom resource deleted")
			_, err = reconcileResource(namespaceLabelReconciler, ctx, typeNamespaceName.Name+"3", typeNamespaceName.Namespace, false)
			Expect(err).Should(BeNil())

			By("Checking the previous labels exist")
			k8sClient.Get(ctx, types.NamespacedName{Name: typeNamespaceName.Namespace}, namespace)
			labels := namespace.GetLabels()
			Expect(labels["label_4"] == "2005" && labels["label_3"] == "2006").Should(BeTrue())

			By("Deleting the second custom resource")
			err = changeNamespaceLabelState(labels2, ctx, typeNamespaceName.Name+"2", typeNamespaceName.Namespace, Delete)
			Expect(err).Should(BeNil())

			By("Reconciling the third custom resource deleted")
			_, err = reconcileResource(namespaceLabelReconciler, ctx, typeNamespaceName.Name+"2", typeNamespaceName.Namespace, false)
			Expect(err).Should(BeNil())

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
			err := changeNamespaceLabelState(getNamespaceLabelForTest("protectedLabels"), ctx, typeNamespaceName.Name, typeNamespaceName.Namespace, Create)
			Expect(err).Should(BeNil())

			By("Reconciling the custom resource created")
			_, err = reconcileResource(*getNamespaceLabelReconciler(), ctx, typeNamespaceName.Name, typeNamespaceName.Namespace, true)
			Expect(err).Should(BeNil())

			By("Checking the previous labels exist")
			k8sClient.Get(ctx, types.NamespacedName{Name: typeNamespaceName.Namespace}, namespace)
			newLabels := namespace.GetLabels()
			Expect(reflect.DeepEqual(labels, newLabels)).Should(BeTrue())

		})
		It("should return an error if there is no such namespace", func() {

			By("Creating the custom resource for the Kind NamespaceLabel")
			err := changeNamespaceLabelState(getNamespaceLabelForTest("basic"), ctx, typeNamespaceName.Name, typeNamespaceName.Namespace, Create)
			Expect(err).Should(BeNil())

			By("Reconciling the custom resource created")
			_, err = reconcileResource(*getNamespaceLabelReconciler(), ctx, fmt.Sprint(typeNamespaceName)+"a", typeNamespaceName.Namespace, false)
			Expect(err).ShouldNot(BeNil())
		})

		It("should return an error if there is no such namespacelabel", func() {

			By("Creating the custom resource for the Kind NamespaceLabel")
			err := changeNamespaceLabelState(getNamespaceLabelForTest("basic"), ctx, typeNamespaceName.Name, typeNamespaceName.Namespace, Create)
			Expect(err).Should(BeNil())

			By("Reconciling the custom resource created")
			_, err = reconcileResource(*getNamespaceLabelReconciler(), ctx, fmt.Sprint(typeNamespaceName)+"a", typeNamespaceName.Namespace, false)
			Expect(err).ShouldNot(BeNil())
		})

	})
})

// reconcileResource gets the reconciler object, the context, the name and namespace of the object to be reconciled, and should the namespaceLabel
// exist in the evantually function. it returns whether the resource exists and an error.
func reconcileResource(r controllers.NamespaceLabelReconciler, ctx context.Context, name string, namespace string, shouldBe bool) (bool, error) {
	namespacedName := types.NamespacedName{Name: name, Namespace: namespace}
	_, err := r.Reconcile(ctx, reconcile.Request{
		NamespacedName: namespacedName,
	})
	be := BeNil()
	if shouldBe {
		be = BeTrue()
	} else {
		be = BeFalse()
	}
	exists := Eventually(func() bool {
		var namespaceLabel v1alpha1.NamespaceLabel
		err := k8sClient.Get(ctx, namespacedName, &namespaceLabel)
		if err != nil {
			return false
		}
		return true
	}, timeout, interval).Should(be)
	return exists, err
}

//changeNamespaceLabelState gets a map of labels, the context, the name and namespace of the resource to be created/deleted and the action
// (whether to delete or create the resource). It returns an error.
func changeNamespaceLabelState(labels map[string]string, ctx context.Context, name string, namespace string, action string) error {
	namespaceLabel := &v1alpha1.NamespaceLabel{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.NamespaceLabelSpec{
			Labels: labels,
		},
	}
	err := errors.New("")
	switch action {
	case "create":
		err = k8sClient.Create(ctx, namespaceLabel)
	case "delete":
		err = k8sClient.Delete(ctx, namespaceLabel)
	default:
		return errors.New("Invalid action")
	}
	return err
}
