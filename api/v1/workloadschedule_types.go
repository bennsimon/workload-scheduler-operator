/*
Copyright 2023.

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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WorkloadScheduleSpec defines the desired state of WorkloadSchedule
type WorkloadScheduleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Selector  WorkloadSelector       `json:"selector,omitempty"`
	Schedules []WorkloadScheduleUnit `json:"schedules,omitempty"`
}

type WorkloadScheduleUnit struct {
	Schedule string `json:"schedule,omitempty"`
	Desired  int32  `json:"desired,omitempty"`
}

// WorkloadScheduleStatus defines the observed state of WorkloadSchedule
type WorkloadScheduleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

type WorkloadSelector struct {
	Namespaces []string          `json:"namespaces,omitempty"`
	Kinds      []string          `json:"kinds,omitempty"` //TODO: make it enum
	Names      []string          `json:"names,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}

type WorkloadScheduleData struct {
	WorkloadScheduler string            `json:"workloadScheduler,omitempty"`
	Namespace         string            `json:"namespace,omitempty"`
	Kind              string            `json:"kind,omitempty"` //TODO: make it enum
	Name              string            `json:"name,omitempty"`
	Desired           int32             `json:"desired,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// WorkloadSchedule is the Schema for the workloadschedules API
type WorkloadSchedule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkloadScheduleSpec   `json:"spec,omitempty"`
	Status WorkloadScheduleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WorkloadScheduleList contains a list of WorkloadSchedule
type WorkloadScheduleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkloadSchedule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WorkloadSchedule{}, &WorkloadScheduleList{})
}
