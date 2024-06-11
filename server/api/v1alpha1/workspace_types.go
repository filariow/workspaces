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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WorkspaceVisibility string

const (
	// WorkspaceVisibilityCommunity Community value for Workspaces visibility
	WorkspaceVisibilityCommunity WorkspaceVisibility = "community"
	// WorkspaceVisibilityPrivate Private value for Workspaces visibility
	WorkspaceVisibilityPrivate WorkspaceVisibility = "private"
)

// UserInfo contains information about a user identity
type UserInfo struct {
	//+required
	Email string `json:"email"`
}

// WorkspaceSpec defines the desired state of Workspace
type WorkspaceSpec struct {
	//+required
	Visibility WorkspaceVisibility `json:"visibility"`

	//+required
	Owner UserInfo `json:"owner"`
}

// SpaceInfo Information about a Space
type SpaceInfo struct {
	//+required
	Name string `json:"name"`
}

// UserInfoStatus User info stored in the status
type UserInfoStatus struct {
	//+optional
	Username string `json:"username,omitempty"`
}

// WorkspaceStatus defines the observed state of Workspace
type WorkspaceStatus struct {
	//+optional
	Space *SpaceInfo `json:"space,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Visibility",type="string",JSONPath=`.spec.visibility`

// Workspace is the Schema for the workspaces API
type Workspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceSpec   `json:"spec,omitempty"`
	Status WorkspaceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WorkspaceList contains a list of Workspace
type WorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workspace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workspace{}, &WorkspaceList{})
}
