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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type InternalWorkspaceVisibility string

const (
	// DisplayNameDefaultWorkspace display name for the default Workspace
	DisplayNameDefaultWorkspace string = "default"

	// InternalWorkspaceVisibilityCommunity Community value for InternalWorkspaces visibility
	InternalWorkspaceVisibilityCommunity InternalWorkspaceVisibility = "community"
	// InternalWorkspaceVisibilityPrivate Private value for InternalWorkspaces visibility
	InternalWorkspaceVisibilityPrivate InternalWorkspaceVisibility = "private"

	// LabelInternalDomain domain for internal labels
	LabelInternalDomain string = "internal.workspaces.konflux.io/"
	// LabelDisplayName label for storing the user chosen display-name for the workspace
	LabelDisplayName string = LabelInternalDomain + "display-name"
	// LabelWorkspaceOwner owner label
	// Deprecated: use field
	LabelWorkspaceOwner string = LabelInternalDomain + "owner"

	// PublicViewerName the name of the KubeSaw's PublicViewer user
	PublicViewerName string = "kubesaw-authenticated"
)

// UserInfo contains information about a user identity
type UserInfo struct {
	//+required
	JwtInfo JwtInfo `json:"jwtInfo"`

	//+required
	Identity IdentityInfo `json:"identity"`
}

// +kubebuilder:validation:MaxProperties:=1
// IdentityInfo decouples integration with an identity management system
type IdentityInfo struct {
	//+optional
	UserSignupRef *UserSignupRef `json:"userSignupRef,omitempty"`
}

// UserSignupRef reference to an UserSignup
type UserSignupRef struct {
	corev1.ObjectReference `json:",inline"`
}

// JwtInfo contains information extracted from the user JWT Token
type JwtInfo struct {
	//+required
	Email string `json:"email"`
	//+required
	UserId string `json:"userId"`
	//+required
	Sub string `json:"sub"`

	//+optional
	PreferredUsername string `json:"preferredUsername,omitempty"`
	//+optional
	AccountId string `json:"accountId,omitempty"`
	//+optional
	Company string `json:"company,omitempty"`
	//+optional
	GivenName string `json:"giveName,omitempty"`
	//+optional
	FamilyName string `json:"familyName,omitempty"`
}

// InternalWorkspaceSpec defines the desired state of Workspace
type InternalWorkspaceSpec struct {
	//+DisplayName
	DisplayName string `json:"displayName"`
	//+required
	Visibility InternalWorkspaceVisibility `json:"visibility"`
	//+required
	Owner UserInfo `json:"owner"`
}

// SpaceInfo Information about a Space
type SpaceInfo struct {
	//+required
	Name string `json:"name"`
	//+required
	IsHome bool `json:"isHome"`
}

// InternalWorkspaceStatus defines the observed state of Workspace
type InternalWorkspaceStatus struct {
	// Space contains information about the underlying Space
	//+optional
	Space *SpaceInfo `json:"space,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Visibility",type="string",JSONPath=`.spec.visibility`

// InternalWorkspace is the Schema for the workspaces API
type InternalWorkspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InternalWorkspaceSpec   `json:"spec,omitempty"`
	Status InternalWorkspaceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InternalWorkspaceList contains a list of Workspace
type InternalWorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InternalWorkspace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InternalWorkspace{}, &InternalWorkspaceList{})
}
