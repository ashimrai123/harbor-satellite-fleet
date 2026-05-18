/*
Copyright 2026.

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

// Condition types for Satellite
const (
	// ConditionRegistered indicates the satellite is registered in Ground Control.
	ConditionRegistered = "Registered"
	// ConditionGroupsSynced indicates all specified groups are synced in Ground Control.
	ConditionGroupsSynced = "GroupsSynced"
	// ConditionConfigAssigned indicates the config is assigned to the satellite.
	ConditionConfigAssigned = "ConfigAssigned"
	// ConditionReady indicates the satellite is fully reconciled.
	ConditionReady = "Ready"
)

// SatellitePhase represents the lifecycle phase of a Satellite.
type SatellitePhase string

const (
	SatellitePhasePending     SatellitePhase = "Pending"
	SatellitePhaseRegistering SatellitePhase = "Registering"
	SatellitePhaseReady       SatellitePhase = "Ready"
	SatellitePhaseError       SatellitePhase = "Error"
)

// SatelliteSpec defines the desired state of a Satellite.
type SatelliteSpec struct {
	// GroundControlURL is the Ground Control API endpoint.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^https?://.*`
	GroundControlURL string `json:"groundControlURL"`

	// ConfigName is the name of the Ground Control config to assign to this satellite.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	ConfigName string `json:"configName"`

	// Groups is the list of Ground Control groups this satellite should belong to.
	// +optional
	Groups []string `json:"groups,omitempty"`

	// SecretRef is the name of the Secret in the same namespace containing
	// Ground Control credentials (keys: username, password).
	// +kubebuilder:validation:Required
	SecretRef string `json:"secretRef"`

	// TokenSecretName is the name of the Secret where the ZTR bootstrap token
	// will be stored after registration. Defaults to "<satellite-name>-token".
	// +optional
	TokenSecretName string `json:"tokenSecretName,omitempty"`
}

// SatelliteStatus defines the observed state of Satellite.
type SatelliteStatus struct {
	// Phase is the current lifecycle phase of the Satellite.
	// +optional
	Phase SatellitePhase `json:"phase,omitempty"`

	// TokenSecret is the name of the Secret storing the ZTR bootstrap token.
	// +optional
	TokenSecret string `json:"tokenSecret,omitempty"`

	// GroundControlID is the ID assigned to this satellite in Ground Control.
	// +optional
	GroundControlID int32 `json:"groundControlID,omitempty"`

	// LastSyncTime is when the operator last reconciled this satellite.
	// +optional
	LastSyncTime *metav1.Time `json:"lastSyncTime,omitempty"`

	// Conditions represent the current state of the Satellite resource.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="GC-ID",type=integer,JSONPath=`.status.groundControlID`
// +kubebuilder:printcolumn:name="Config",type=string,JSONPath=`.spec.configName`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Satellite is the Schema for the satellites API.
// It represents a single Harbor Satellite instance managed by Ground Control.
type Satellite struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of Satellite.
	// +required
	Spec SatelliteSpec `json:"spec"`

	// status defines the observed state of Satellite.
	// +optional
	Status SatelliteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SatelliteList contains a list of Satellite.
type SatelliteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Satellite `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Satellite{}, &SatelliteList{})
}
