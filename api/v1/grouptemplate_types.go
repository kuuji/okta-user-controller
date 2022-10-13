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

	"github.com/Masterminds/sprig/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GroupTemplateSpec defines the desired state of GroupTemplate
type GroupTemplateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The name of the group to sync
	Group    string `json:"group,omitempty"`
	Template string `json:"template,omitempty"`
}

// GroupTemplateStatus defines the observed state of GroupTemplate
type GroupTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	TemplateErrors map[string]string `json:"templateErrors,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GroupTemplate is the Schema for the grouptemplates API
type GroupTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupTemplateSpec   `json:"spec,omitempty"`
	Status GroupTemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GroupTemplateList contains a list of GroupTemplate
type GroupTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GroupTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GroupTemplate{}, &GroupTemplateList{})
}

func (in GroupTemplate) ProcessTemplate(o interface{}) (string, error) {
	// funcMap := template.FuncMap{
	// 	"replace": replace,
	// 	"lower":   lower,
	// }

	tpl, err := template.New("").Funcs(sprig.FuncMap()).Parse(string(in.Spec.Template))
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	err = tpl.Execute(&out, o)
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
