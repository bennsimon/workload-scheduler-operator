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

package controller

import (
	workloadschedulerv1 "bennsimon.github.io/workload-scheduler-operator/api/v1"
	"bennsimon.github.io/workload-scheduler-operator/handler/workloadScheduleHandler"
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WorkloadScheduleReconciler reconciles a WorkloadSchedule object
type WorkloadScheduleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	workloadScheduleHandler.IWorkloadScheduleHandler
}

//+kubebuilder:rbac:groups=workload-scheduler.bennsimon.github.io,resources=workloadschedules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=workload-scheduler.bennsimon.github.io,resources=workloadschedules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=workload-scheduler.bennsimon.github.io,resources=workloadschedules/finalizers,verbs=update

func (r *WorkloadScheduleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	workloadSchedule := &workloadschedulerv1.WorkloadSchedule{}
	err := r.Get(ctx, client.ObjectKey{Name: req.Name, Namespace: req.Namespace}, workloadSchedule)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = r.IWorkloadScheduleHandler.ValidateWorkloadSchedule(workloadSchedule)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *WorkloadScheduleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&workloadschedulerv1.WorkloadSchedule{}).
		Complete(r)
}
