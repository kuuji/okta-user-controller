/*
Copyright 2022 kuuji.

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
	"bytes"
	"html/template"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GroupSyncerSpec defines the desired state of GroupSyncer
type GroupSyncerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The name of the group to sync
	Group    string `json:"group,omitempty"`
	Template string `json:"template,omitempty"`
}

// GroupSyncerStatus defines the observed state of GroupSyncer
type GroupSyncerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GroupSyncer is the Schema for the groupsyncers API
type GroupSyncer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupSyncerSpec   `json:"spec,omitempty"`
	Status GroupSyncerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GroupSyncerList contains a list of GroupSyncer
type GroupSyncerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GroupSyncer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GroupSyncer{}, &GroupSyncerList{})
}

func (in GroupSyncer) ProcessTemplate(o interface{}) (string, error) {
	funcMap := template.FuncMap{
		"replace": replace,
		"lower":   lower,
	}
	tpl, _ := template.New("").Funcs(funcMap).Parse(string(in.Spec.Template))
	var out bytes.Buffer
	err := tpl.Execute(&out, o)
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func replace(input, from, to string) string {
	return strings.Replace(input, from, to, -1)
}

func lower(input string) string {
	return strings.ToLower(input)
}
