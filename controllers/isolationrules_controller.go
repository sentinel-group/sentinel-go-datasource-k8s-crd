/*


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

	"github.com/alibaba/sentinel-golang/core/isolation"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datasourcev1 "github.com/sentinel-group/sentinel-go-datasource-k8s-crd/api/v1"
)

// IsolationRulesReconciler reconciles a IsolationRules object
type IsolationRulesReconciler struct {
	client.Client
	Logger          logr.Logger
	Scheme          *runtime.Scheme
	Namespace       string
	EffectiveCrName string
}

// +kubebuilder:rbac:groups=datasource.sentinel.io,resources=isolationrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=datasource.sentinel.io,resources=isolationrules/status,verbs=get;update;patch

func (r *IsolationRulesReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Logger
	log.Info("receive IsolationRules", "namespace", req.NamespacedName.String())

	if req.Namespace != r.Namespace {
		log.V(int(logging.WarnLevel)).Info("ignore unmatched namespace.", "namespace", r.Namespace, "req", req.String())
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, nil
	}

	if req.Name != r.EffectiveCrName {
		log.V(int(logging.WarnLevel)).Info("ignore unmatched cr.", "cr", r.EffectiveCrName, "req", req.String())
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, nil
	}

	isolationRulesCR := &datasourcev1.IsolationRules{}
	if err := r.Get(ctx, req.NamespacedName, isolationRulesCR); err != nil {
		log.Error(err, "Fail to get datasourcev1.IsolationRules.")
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, err
	}
	log.Info("Get datasourcev1.IsolationRules", "rules:", isolationRulesCR.Spec.Rules)
	// your logic here
	isolationRules := make([]*isolation.Rule, 0, len(isolationRulesCR.Spec.Rules))
	for _, r := range isolationRulesCR.Spec.Rules {
		isolationRules = append(isolationRules, &isolation.Rule{
			ID:         r.ID,
			Resource:   r.Resource,
			MetricType: isolation.Concurrency,
			Threshold:  uint32(r.Threshold),
		})
	}

	_, err := isolation.LoadRules(isolationRules)
	if err != nil {
		log.Error(err, "Fail to Load isolation.Rules")
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, err
	}
	return ctrl.Result{}, nil
}

func (r *IsolationRulesReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datasourcev1.IsolationRules{}).
		Complete(r)
}
