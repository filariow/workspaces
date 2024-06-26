# Workflows

In this section are detailed the main workflow implemented by this operator.

## Home Workspace

When an [KubeSaw](https://github.com/codeready-toolchain) UserSignup is approved, a Space is created by default.
This workflow monitors ensures an InternalWorkspace exists for the Space.

This workflow is implemented by the [UserSignup Reconciler](https://github.com/konflux-workspaces/workspaces/blob/main/operator/internal/controller/usersignup/usersignup_controller.go).


## Public Viewer

InternalWorkspaces have a property for their visibility.
Visibility can be either `private` or `community`.
A `private` InternalWorkspace is visible only by its owner and the users it's directly shared with.
A `community` InternalWorkspace is visible by every authenticated users.

If an InternalWorkspace visibility is set to `community`, the operator makes sure that a SpaceBinding exists for the special-user `kubesaw-authenticated`, the space related to the InternalWorkspace, and the role `viewer`.
If the visibility is set to `private`, the SpaceBinding is ensured not to exist.

This workflow is mainly implemented in the [InternalWorkspace Reconciler](https://github.com/konflux-workspaces/workspaces/blob/main/operator/internal/controller/internalworkspace/internalworkspace_controller.go).
