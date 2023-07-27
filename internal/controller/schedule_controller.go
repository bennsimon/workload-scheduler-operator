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
	"bennsimon.github.io/workload-scheduler-operator/handler/scheduleHandler"
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	workloadschedulerv1 "bennsimon.github.io/workload-scheduler-operator/api/v1"
)

// ScheduleReconciler reconciles a Schedule object
type ScheduleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	scheduleHandler.IScheduleHandler
}

//+kubebuilder:rbac:groups=workload-scheduler.bennsimon.github.io,resources=schedules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=workload-scheduler.bennsimon.github.io,resources=schedules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=workload-scheduler.bennsimon.github.io,resources=schedules/finalizers,verbs=update
//+kubebuilder:resource:scope=cluster

func (r *ScheduleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	schedule := &workloadschedulerv1.Schedule{}
	err := r.Get(ctx, client.ObjectKey{Name: req.Name, Namespace: req.Namespace}, schedule)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = r.IScheduleHandler.ValidateSchedule(schedule)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScheduleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&workloadschedulerv1.Schedule{}, builder.WithPredicates(r.FilterEvents())).
		Complete(r)
}

func (r *ScheduleReconciler) FilterEvents() predicate.Predicate {

	return predicate.Funcs{CreateFunc: func(createEvent event.CreateEvent) bool {
		return true
	}, UpdateFunc: func(updateEvent event.UpdateEvent) bool {
		return true
	}, DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
		return false
	}, GenericFunc: func(genericEvent event.GenericEvent) bool {
		return false
	},
	}
}
