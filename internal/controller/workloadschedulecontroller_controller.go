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
	"bennsimon.github.io/workload-scheduler-operator/handler/workloadScheduleHandler"
	"bennsimon.github.io/workload-scheduler-operator/util/config"
	"context"
	"fmt"
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	workloadschedulerv1 "bennsimon.github.io/workload-scheduler-operator/api/v1"
)

// WorkloadScheduleControllerReconciler reconciles a WorkloadScheduleController object
type WorkloadScheduleControllerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	workloadScheduleHandler.IWorkloadScheduleHandler
}

//+kubebuilder:rbac:groups=workload-scheduler.bennsimon.github.io,resources=workloadschedulecontrollers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=workload-scheduler.bennsimon.github.io,resources=workloadschedulecontrollers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=workload-scheduler.bennsimon.github.io,resources=workloadschedulecontrollers/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;update;watch
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;update;watch

func (r *WorkloadScheduleControllerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	configUtil := config.New()
	workloadSchedulers := &workloadschedulerv1.WorkloadScheduleList{}
	err := r.List(ctx, workloadSchedulers)
	t := time.Now()

	if configUtil.LookUpBooleanEnv(config.Debug) {
		log.Log.Info(fmt.Sprintf("started processing at %s", t))
	}

	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	} else {
		workloadSchedulerAndSchedules, workloadSchedulerMap := r.IWorkloadScheduleHandler.EvaluateWorkloadSchedulers(workloadSchedulers, r, ctx)

		err = r.IWorkloadScheduleHandler.ProcessWorkloadSchedules(workloadSchedulerAndSchedules, workloadSchedulerMap, r.Client, ctx)
		if err != nil {
			log.Log.Error(err, "an error occurred on ProcessWorkloadSchedules()")
			return ctrl.Result{}, err
		}
	}

	if configUtil.LookUpBooleanEnv(config.Debug) {
		log.Log.Info(fmt.Sprintf("ended processing at %s, duration: %v", time.Now(), time.Since(t).Seconds()))
	}

	recon, err := configUtil.LookUpIntEnv(config.ReconciliationDuration)
	if err != nil {
		recon = 60
	}

	return ctrl.Result{RequeueAfter: time.Duration(recon) * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WorkloadScheduleControllerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &apps.Deployment{}, config.IndexedField, func(rawObj client.Object) []string {
		deployment := rawObj.(*apps.Deployment)

		if deployment == nil {
			return nil
		}
		return []string{deployment.Name}
	}); err != nil {
		return err
	}
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &workloadschedulerv1.Schedule{}, config.IndexedField, func(rawObj client.Object) []string {
		schedule := rawObj.(*workloadschedulerv1.Schedule)

		if schedule == nil {
			return nil
		}

		if schedule.Kind != "Schedule" {
			return nil
		}

		return []string{schedule.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&workloadschedulerv1.WorkloadScheduleController{}, builder.WithPredicates(r.FilterEvents())).
		Complete(r)
}

func (r *WorkloadScheduleControllerReconciler) FilterEvents() predicate.Predicate {
	return predicate.Funcs{CreateFunc: func(createEvent event.CreateEvent) bool {
		return r.reconcileIfInitialController()
	}, UpdateFunc: func(updateEvent event.UpdateEvent) bool {
		return false
	}, DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
		return false
	}, GenericFunc: func(genericEvent event.GenericEvent) bool {
		return false
	},
	}
}

func (r *WorkloadScheduleControllerReconciler) reconcileIfInitialController() bool {
	workloadScheduleControllerList := workloadschedulerv1.WorkloadScheduleControllerList{}
	err := r.List(context.Background(), &workloadScheduleControllerList)
	if err != nil {
		return false
	}

	if len(workloadScheduleControllerList.Items) > 1 {
		err := r.Delete(context.Background(), &workloadScheduleControllerList.Items[len(workloadScheduleControllerList.Items)-1])
		if err != nil {
			return false
		}
		log.Log.Error(fmt.Errorf("a single WorkloadScheduleController CR is required. Found %d. Additional CRs deleted", len(workloadScheduleControllerList.Items)), "")
	}

	return true
}
