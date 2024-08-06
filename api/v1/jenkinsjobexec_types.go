/*
Copyright 2024.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type ConfigMapReference struct {
	// Name of the ConfigMap
	Name string `json:"name"`
	// Namespace of the ConfigMap (optional, if the ConfigMap is in the same namespace as the CR)
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

type SecretReference struct {
	// Name of the secret
	Name string `json:"name"`
	// Namespace of the secret (optional, if the secret is in the same namespace as the CR)
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// JenkinsJobExecSpec defines the desired state of JenkinsJobExec

type JenkinsJobExecSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// jenkinsURL is an example field of JenkinsJobExec. Edit jenkinsjobexec_types.go to remove/update
	//JenkinsURL string `json:"jenkinsURL,omitempty"`
	JobName string `json:"jobname,omitempty"`
	//Token      string `json:"token,omitempty"`
	//Username   string `json:"username,omitempty"`

	Parameters map[string]string `json:"parameters,omitempty"`
	// SecretRef is a reference to a Kubernetes Secret
	SecretRef SecretReference `json:"secretRef"`
	// ConfigMapRef is a reference to a Kubernetes ConfigMap
	ConfigMapRef ConfigMapReference `json:"configMapRef"`
}

// JenkinsJobExecStatus defines the observed state of JenkinsJobExec
type JenkinsJobExecStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	JobStatus string `json:"jobstatus,omitempty"`
	BuildURL  string `json:"buildurl,omitempty"`
	Processed bool   `json:"processed,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="JobName",type="string",JSONPath=".spec.jobname",description="Job Name"
// +kubebuilder:printcolumn:name="JobStatus",type="string",JSONPath=".status.jobstatus",description="Job Status"
// +kubebuilder:printcolumn:name="BuildURL",type="string",JSONPath=".status.buildurl",description="Job Build URL"

// JenkinsJobExec is the Schema for the jenkinsjobexecs API
type JenkinsJobExec struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JenkinsJobExecSpec   `json:"spec,omitempty"`
	Status JenkinsJobExecStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// JenkinsJobExecList contains a list of JenkinsJobExec
type JenkinsJobExecList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JenkinsJobExec `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JenkinsJobExec{}, &JenkinsJobExecList{})
}
