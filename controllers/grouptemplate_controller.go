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

package controllers

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	oktav1 "github.com/kuuji/okta-user-controller/api/v1"
	"github.com/kuuji/okta-user-controller/internal/okta"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// GroupTemplateReconciler reconciles a GroupTemplate object
type GroupTemplateReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	OktaConfig *okta.OktaConfig
}

//+kubebuilder:rbac:groups=okta.github.com,resources=grouptemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=okta.github.com,resources=grouptemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=okta.github.com,resources=grouptemplates/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GroupTemplate object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *GroupTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctrlLog := log.FromContext(ctx)

	var gs oktav1.GroupTemplate
	if err := r.Get(ctx, req.NamespacedName, &gs); err != nil {
		ctrlLog.Info("unable to fetch GroupTemplate, likely deleted")
		ctrlLog.Info("removing associated configmap", "configmap", req.NamespacedName)
		err := r.Client.Delete(ctx, &corev1.ConfigMap{ObjectMeta: v1.ObjectMeta{Name: req.Name, Namespace: req.Namespace}})
		if err != nil {
			ctrlLog.Info("failed to delete configmap, likely already gone", "cm", req.NamespacedName)
			ctrlLog.Info("stopping from requeing")
			return ctrl.Result{Requeue: false}, nil
		}
		ctrlLog.Info("configmap removed", "configmap", req.NamespacedName)
		return ctrl.Result{Requeue: false}, nil
	}

	o := okta.NewOktaService(*r.OktaConfig)
	users, err := o.GetUsersFromGroup(gs.Spec.Group)
	if err != nil {
		return ctrl.Result{}, err
	}

	var cmData string
	tplErrs := make(map[string]string)
	for _, v := range users {
		out, err := gs.ProcessTemplate((*v))
		if err != nil {
			profile := *v.Profile
			if login, ok := profile["login"]; ok {
				tplErrs[login.(string)] = err.Error()
			}

		}
		obj := make(map[string]interface{})
		err = yaml.Unmarshal([]byte(out), &obj)
		if err != nil {
			return ctrl.Result{}, err
		}
		if kind, ok := obj["kind"]; ok {
			switch kind {
			case "Secret":
				secret := corev1.Secret{}
				err := yaml.Unmarshal([]byte(out), &secret)
				if err != nil {
					return ctrl.Result{}, err
				}
				secret.Namespace = gs.Namespace
				err = r.Client.Create(ctx, &secret)
				if errors.IsAlreadyExists(err) {
					err = r.Client.Update(ctx, &secret)
					if err != nil {
						return ctrl.Result{}, err
					}
					ctrlLog.Info("secret updated", "secret", secret)
				} else if err != nil {
					return ctrl.Result{}, err
				} else {
					ctrlLog.Info("secret created", "secret", secret)
				}
			}
		}
		cmData = fmt.Sprintf("%s%s", cmData, out)
	}
	gs.Status.TemplateErrors = tplErrs
	if err := r.Status().Update(ctx, &gs); err != nil {
		ctrlLog.Error(err, "unable to update GroupTemplate status")
		return ctrl.Result{}, err
	}
	if len(tplErrs) > 0 {
		return ctrl.Result{}, fmt.Errorf("error in the template, aborting early")
	}
	// cm := &corev1.ConfigMap{
	// 	ObjectMeta: v1.ObjectMeta{Name: gs.Name, Namespace: gs.Namespace},
	// 	Data: map[string]string{
	// 		"values": cmData,
	// 	},
	// }

	// err = r.Client.Create(ctx, cm)
	// if errors.IsAlreadyExists(err) {
	// 	err = r.Client.Update(ctx, cm)
	// 	if err != nil {
	// 		return ctrl.Result{}, err
	// 	}
	// 	ctrlLog.Info("configMap updated", "cm", cm)
	// } else if err != nil {
	// 	return ctrl.Result{}, err
	// } else {
	// 	ctrlLog.Info("configMap created", "cm", cm)
	// }

	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GroupTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&oktav1.GroupTemplate{}).
		Complete(r)
}
