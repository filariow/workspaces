package mapper

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

func (m *Mapper) InternalWorkspaceToWorkspace(workspace *workspacesv1alpha1.InternalWorkspace) (*restworkspacesv1alpha1.Workspace, error) {
	// retrieve external labels
	wll := map[string]string{}
	for k, v := range workspace.GetLabels() {
		if !strings.HasPrefix(k, workspacesv1alpha1.LabelInternalDomain) {
			wll[k] = v
		}
	}

	return &restworkspacesv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:       workspace.Spec.DisplayName,
			Namespace:  workspace.Status.Owner.Username,
			Labels:     wll,
			Generation: workspace.Generation,
		},
		Spec: restworkspacesv1alpha1.WorkspaceSpec{
			Visibility: restworkspacesv1alpha1.WorkspaceVisibility(workspace.Spec.Visibility),
			Owner: restworkspacesv1alpha1.UserInfo{
				JwtInfo: restworkspacesv1alpha1.JwtInfo{
					Email:  workspace.Spec.Owner.JwtInfo.Email,
					UserId: workspace.Spec.Owner.JwtInfo.UserId,
					Sub:    workspace.Spec.Owner.JwtInfo.Sub,

					AccountId:         workspace.Spec.Owner.JwtInfo.AccountId,
					PreferredUsername: workspace.Spec.Owner.JwtInfo.PreferredUsername,
					Company:           workspace.Spec.Owner.JwtInfo.Company,
					GivenName:         workspace.Spec.Owner.JwtInfo.GivenName,
					FamilyName:        workspace.Spec.Owner.JwtInfo.FamilyName,
				},
			},
		},
		Status: restworkspacesv1alpha1.WorkspaceStatus{
			Space: &restworkspacesv1alpha1.SpaceInfo{
				Name:   workspace.Spec.Space,
				IsHome: workspace.Status.Space.IsHome,
			},
		},
	}, nil
}
