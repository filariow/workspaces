package dump

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	yaml "sigs.k8s.io/yaml/goyaml.v3"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

var resourcesToDump = []client.ObjectList{
	&workspacesv1alpha1.InternalWorkspaceList{},
	&toolchainv1alpha1.UserSignupList{},
	&toolchainv1alpha1.MasterUserRecordList{},
	&toolchainv1alpha1.SpaceList{},
	&toolchainv1alpha1.SpaceBindingList{},
}

func DumpAll(ctx context.Context) error {
	rr := slices.Clone(resourcesToDump)

	errs := []error{}
	for _, r := range rr {
		err := dumpResourceInAllNamespaces(ctx, r)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func dumpResourceInAllNamespaces(ctx context.Context, resource client.ObjectList) error {
	// retrieve host client
	cli := tcontext.RetrieveHostClient(ctx)

	// list resources
	if err := cli.Client.List(ctx, resource, client.InNamespace(metav1.NamespaceAll)); err != nil {
		return err
	}

	return dumpCleanUnstructuredOrResource(resource)
}

func dumpCleanUnstructuredOrResource(resource client.ObjectList) error {
	if err := dumpCleanUnstructured(resource); err == nil {
		return nil
	}
	return dumpResource(resource)
}

func dumpCleanUnstructured(resource client.ObjectList) error {
	ur, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
	if err != nil {
		return err
	}

	ii := ur["items"].([]interface{})
	for _, i := range ii {
		ie, ok := i.(map[string]interface{})
		if !ok {
			continue
		}

		om, ok := ie["objectmeta"].(map[string]interface{})
		if !ok {
			continue
		}

		om["managedfields"] = nil
	}

	return dumpResource(ur)
}

func dumpResource(resource any) error {
	// marshal to yaml
	o, err := yaml.Marshal(resource)
	if err != nil {
		return err
	}

	// print on stderr
	if _, err := fmt.Fprintln(os.Stderr, string(o)); err != nil {
		return err
	}

	return nil

}
