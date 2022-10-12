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

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	oktav1 "github.com/kuuji/okta-user-controller/api/v1"
	"github.com/kuuji/okta-user-controller/internal/okta"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

// GroupSyncerReconciler reconciles a GroupSyncer object
type GroupSyncerReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	OktaConfig *okta.OktaConfig
}

//+kubebuilder:rbac:groups=okta.github.com,resources=groupsyncers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=okta.github.com,resources=groupsyncers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=okta.github.com,resources=groupsyncers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GroupSyncer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *GroupSyncerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctrlLog := log.FromContext(ctx)

	var gs oktav1.GroupSyncer
	if err := r.Get(ctx, req.NamespacedName, &gs); err != nil {
		ctrlLog.Info("unable to fetch GroupSyncer, likely deleted")
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
	for _, v := range users {
		out, err := gs.ProcessTemplate((*v))
		if err != nil {
			return ctrl.Result{}, err
		}
		cmData = fmt.Sprintf("%s%s", cmData, out)
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{Name: gs.Name, Namespace: gs.Namespace},
		Data: map[string]string{
			"values": cmData,
		},
	}

	err = r.Client.Create(ctx, cm)
	if errors.IsAlreadyExists(err) {
		err = r.Client.Update(ctx, cm)
		if err != nil {
			return ctrl.Result{}, err
		}
		ctrlLog.Info("configMap updated", "cm", cm)
	} else if err != nil {
		return ctrl.Result{}, err
	} else {
		ctrlLog.Info("configMap created", "cm", cm)
	}

	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GroupSyncerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&oktav1.GroupSyncer{}).
		Complete(r)
}
