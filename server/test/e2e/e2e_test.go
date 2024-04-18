/*
Copyright 2024 The Workspaces Authors.

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

package e2e

import (
	"context"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	workspacesapiv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/persistence/kube"
	"github.com/konflux-workspaces/workspaces/server/test/utils"
)

var cfg *rest.Config

var _ = BeforeSuite(func() {
	var err error
	var cmd *exec.Cmd

	// read kubeconfig
	apiConfig, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	Expect(err).To(BeNil())

	clientConfig := clientcmd.NewDefaultClientConfig(*apiConfig, nil)
	cfg, err = clientConfig.ClientConfig()
	Expect(err).To(BeNil())

	By("installing CRDs")
	cmd = exec.Command("make", "install")
	_, err = utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
})

var _ = Describe("controller", Ordered, func() {
	var cache *kube.Cache
	var ctx context.Context
	var workspacesNamespace = corev1.Namespace{ObjectMeta: metav1.ObjectMeta{GenerateName: "workspaces-system-"}}

	BeforeAll(func() {
		// create context
		ctx = context.TODO()

		cli, err := buildClient(ctx, cfg, corev1.AddToScheme)
		Expect(err).To(BeNil())

		err = cli.Create(ctx, &workspacesNamespace)
		Expect(err).To(BeNil())

		// create cache
		c, err := kube.NewWorkspacesCache(ctx, cfg, workspacesNamespace.Name, "kubesaw-system")
		Expect(err).To(BeNil())
		cache = c
	})

	AfterAll(func() {
		cli, err := buildClient(ctx, cfg, corev1.AddToScheme)
		Expect(err).To(BeNil())

		err = cli.Delete(ctx, &workspacesNamespace)
		Expect(err).To(BeNil())
	})

	Context("Cache", func() {
		It("should start successfully", func() {
			go func() {
				err := cache.Start(ctx)
				Expect(err).To(BeNil())
			}()

			synced := cache.WaitForCacheSync(ctx)
			Expect(synced).To(BeTrue())
		})

		When("an internal workspace is added", func() {
			It("should exists an api workspace", func() {
				c, err := buildClient(ctx, cfg, workspacesv1alpha1.AddToScheme, workspacesapiv1alpha1.AddToScheme)
				Expect(err).To(BeNil())

				w := workspacesv1alpha1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unique-id",
						Namespace: workspacesNamespace.Name,
						Labels: map[string]string{
							workspacesv1alpha1.LabelWorkspaceName:  "my-workspaces-name",
							workspacesv1alpha1.LabelWorkspaceOwner: "my-workspaces-namespace",
						},
					},
					Spec: workspacesv1alpha1.WorkspaceSpec{
						Visibility: workspacesv1alpha1.WorkspaceVisibilityPrivate,
					},
				}
				err = c.Create(ctx, &w)
				Expect(err).To(BeNil())

				aw := workspacesapiv1alpha1.Workspace{}
				err = wait.PollUntilContextTimeout(ctx, time.Second, 10*time.Second, true, func(ctx context.Context) (done bool, err error) {
					if err := cache.Get(ctx, types.NamespacedName{Namespace: "my-workspaces-namespace", Name: "my-workspaces-name"}, &aw); err != nil {
						if errors.IsNotFound(err) {
							return false, nil
						}
						return false, err
					}
					return true, nil
				})
				Expect(err).To(BeNil())
			})
		})
	})
})

func buildClient(ctx context.Context, cfg *rest.Config, addsToScheme ...func(*runtime.Scheme) error) (client.Client, error) {
	s := runtime.NewScheme()
	for _, addToScheme := range addsToScheme {
		if err := addToScheme(s); err != nil {
			return nil, err
		}
	}

	hc, err := rest.HTTPClientFor(cfg)
	if err != nil {
		return nil, err
	}

	m, err := apiutil.NewDynamicRESTMapper(cfg, hc)
	if err != nil {
		return nil, err
	}

	return client.New(cfg, client.Options{Scheme: s, Mapper: m})
}
