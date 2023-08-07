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
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	"github.com/go-co-op/gocron"
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

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;update;watch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;update;watch

func (r *WorkloadScheduleControllerReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func (r *WorkloadScheduleControllerReconciler) InitiateSchedule() {
	configUtil := config.New()

	s := gocron.NewScheduler(time.Local)
	recon, err := configUtil.LookUpIntEnv(config.ReconciliationDuration)
	if err != nil {
		recon = 60
	}
	_, err = s.Every(recon).Seconds().Do(func() {
		err := r.RunJob(context.Background())
		if err != nil {
			log.Log.Error(err, "error occurred when executing RunJob()")
		}
	})
	if err != nil {
		log.Log.Error(err, "error when scheduling job start not called")
		return
	}

	s.StartAsync()
}

func (r *WorkloadScheduleControllerReconciler) RunJob(ctx context.Context) error {
	configUtil := config.New()
	workloadSchedulers := &workloadschedulerv1.WorkloadScheduleList{}
	err := r.List(ctx, workloadSchedulers)
	t := time.Now().In(time.Local)
	if configUtil.LookUpBooleanEnv(config.Debug) {
		log.Log.Info(fmt.Sprintf("started processing at %s", t))
	}

	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	} else {
		workloadSchedulerAndSchedules, workloadSchedulerMap := r.IWorkloadScheduleHandler.EvaluateWorkloadSchedulers(workloadSchedulers, r, ctx)

		err = r.IWorkloadScheduleHandler.ProcessWorkloadSchedules(workloadSchedulerAndSchedules, workloadSchedulerMap, r.Client, ctx)
		if err != nil {
			log.Log.Error(err, "an error occurred on ProcessWorkloadSchedules()")
			return err
		}
	}

	if configUtil.LookUpBooleanEnv(config.Debug) {
		log.Log.Info(fmt.Sprintf("ended processing at %s, duration: %v", time.Now().In(time.Local), time.Since(t).Seconds()))
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WorkloadScheduleControllerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.TODO(), &apps.Deployment{}, config.IndexedField, func(rawObj client.Object) []string {
		deployment := rawObj.(*apps.Deployment)

		if deployment == nil {
			return nil
		}
		return []string{deployment.ObjectMeta.Name}
	}); err != nil {
		return err
	}
	if err := mgr.GetFieldIndexer().IndexField(context.TODO(), &apps.StatefulSet{}, config.IndexedField, func(rawObj client.Object) []string {
		statefulSet := rawObj.(*apps.StatefulSet)

		if statefulSet == nil {
			return nil
		}
		return []string{statefulSet.ObjectMeta.Name}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&workloadschedulerv1.WorkloadScheduleController{}).
		Complete(r)
}
