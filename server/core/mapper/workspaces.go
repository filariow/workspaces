package mapper

import (
	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

func ToInternal(w toolchainv1alpha1.Workspace) (*workspacesv1alpha1.InternalWorkspace, error) {
  return &workspacesv1alpha1.InternalWorkspace{}, nil
}

func FromInternal(w workspacesv1alpha1.InternalWorkspace) (*toolchainv1alpha1.Workspace, error) {
	return nil, nil
}
