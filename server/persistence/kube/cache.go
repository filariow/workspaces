package kube

import (
	"context"
	"fmt"
	"log"
	"slices"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	workspacesapiv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

const (
	LabelWorkspaceVisibility string = "workspaces.io/visibility"
)

var errWorkspaceMissingLabel = fmt.Errorf("error invalid workspace as it's missing label")

type Cache struct {
	mgr ctrl.Manager

	mux        sync.RWMutex
	workspaces []workspacesapiv1alpha1.Workspace
}

func buildScheme() (*runtime.Scheme, error) {
	s := runtime.NewScheme()
	addToSchemes := []func(*runtime.Scheme) error{
		corev1.AddToScheme,
		metav1.AddMetaToScheme,
		workspacesv1alpha1.AddToScheme,
		workspacesapiv1alpha1.AddToScheme,
		toolchainv1alpha1.AddToScheme,
	}
	for _, addToScheme := range addToSchemes {
		if err := addToScheme(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func NewCache(ctx context.Context, cfg *rest.Config, workspacesNamespaces, kubesawNamespaces string) (*Cache, error) {
	s, err := buildScheme()
	if err != nil {
		return nil, err
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{Scheme: s, Cache: cache.Options{
		DefaultNamespaces: map[string]cache.Config{
			workspacesNamespaces: {},
			kubesawNamespaces:    {},
		},
	}})
	if err != nil {
		return nil, err
	}

	c := &Cache{mgr: mgr, workspaces: []workspacesapiv1alpha1.Workspace{}}
	if err := ctrl.NewControllerManagedBy(mgr).
		For(&workspacesv1alpha1.Workspace{}).
		Watches(&toolchainv1alpha1.SpaceBinding{}, handler.EnqueueRequestsFromMapFunc(
			func(ctx context.Context, o client.Object) []reconcile.Request {
				sb, ok := o.(*toolchainv1alpha1.SpaceBinding)
				if !ok {
					return []reconcile.Request{}
				}

				space, ok := sb.GetLabels()[toolchainv1alpha1.SpaceBindingSpaceLabelKey]
				if !ok {
					return []reconcile.Request{}
				}

				mur, ok := sb.GetLabels()[toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey]
				if !ok {
					return []reconcile.Request{}
				}

				return []reconcile.Request{
					{NamespacedName: types.NamespacedName{Namespace: mur, Name: space}},
				}
			})).
		Complete(reconcile.Func(c.reconcile)); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Cache) Start(ctx context.Context) error {
	return c.mgr.Start(ctx)
}

func (c *Cache) WaitForCacheSync(ctx context.Context) bool {
	return c.mgr.GetCache().WaitForCacheSync(ctx)
}

func (c *Cache) reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log.Printf("reconciling request: %v", request)
	w := workspacesv1alpha1.Workspace{}
	if err := c.mgr.GetClient().Get(ctx, request.NamespacedName, &w); err != nil {
		log.Printf("error reconciling request: %v: %v", request, err)
		if errors.IsNotFound(err) {
			c.ensureWorkspaceNotExists(ctx, request.NamespacedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if err := c.ensureWorkspaceExists(ctx, w); err != nil {
		log.Printf("error reconciling request: %v: %v", request, err)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (c *Cache) ensureWorkspaceNotExists(ctx context.Context, obj types.NamespacedName) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.workspaces = slices.DeleteFunc(c.workspaces, func(workspace workspacesapiv1alpha1.Workspace) bool {
		return workspace.Name == obj.Name && workspace.Namespace == obj.Namespace
	})
}

func (c *Cache) ensureWorkspaceExists(ctx context.Context, workspace workspacesv1alpha1.Workspace) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	// build workspace
	w, err := c.buildWorkspaceApiFromWorkspace(ctx, workspace)
	if err != nil {
		return err
	}

	ww := slices.DeleteFunc(c.workspaces, func(w workspacesapiv1alpha1.Workspace) bool {
		return w.Name == workspace.Name && w.Namespace == workspace.Namespace
	})

	c.workspaces = append(ww, *w)
	return nil
}

func (c *Cache) buildWorkspaceApiFromWorkspace(ctx context.Context, workspace workspacesv1alpha1.Workspace) (*workspacesapiv1alpha1.Workspace, error) {
	mur, ok := workspace.GetLabels()[workspacesv1alpha1.LabelWorkspaceOwner]
	if !ok {
		return nil, fmt.Errorf("%w '%s'", errWorkspaceMissingLabel, workspacesv1alpha1.LabelWorkspaceOwner)
	}

	name, ok := workspace.GetLabels()[workspacesv1alpha1.LabelWorkspaceName]
	if !ok {
		return nil, fmt.Errorf("%w '%s'", errWorkspaceMissingLabel, workspacesv1alpha1.LabelWorkspaceName)
	}

	w := workspacesapiv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: mur,
			Name:      name,
			Labels:    workspace.Labels,
		},
		Spec: workspacesv1alpha1.WorkspaceSpec{
			Visibility: workspace.Spec.Visibility,
		},
		Status: workspacesapiv1alpha1.WorkspaceStatus{
			Space: workspace.Name,
		},
	}

	return &w, nil
}

func (c *Cache) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	out, ok := obj.(*workspacesapiv1alpha1.Workspace)
	if !ok {
		return fmt.Errorf("resource type not managed")
	}

	idx := slices.IndexFunc(c.workspaces, func(workspace workspacesapiv1alpha1.Workspace) bool {
		return workspace.Name == key.Name && workspace.Namespace == key.Namespace
	})

	if idx == -1 {
		return errors.NewNotFound(workspacesapiv1alpha1.GroupVersion.WithResource("workspaces").GroupResource(), key.Name)
	}

	c.workspaces[idx].DeepCopyInto(out)
	return nil
}

func (c *Cache) List(ctx context.Context, obj client.ObjectList, opts ...client.ListOption) error {
	_, ok := obj.(*workspacesapiv1alpha1.WorkspaceList)
	if !ok {
		return fmt.Errorf("resource type not managed")
	}

	panic("not implemented")
}
