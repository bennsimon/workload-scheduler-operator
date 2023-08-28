package workloadScheduleHandler

import (
	workloadschedulerv1 "bennsimon.github.io/workload-scheduler-operator/api/v1"
	"bennsimon.github.io/workload-scheduler-operator/handler/scheduleHandler"
	"bennsimon.github.io/workload-scheduler-operator/util"
	"bennsimon.github.io/workload-scheduler-operator/util/config"
	"context"
	"fmt"
	"github.com/alistanis/cartesian"
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sort"
	"strings"
	"time"
)

type WorkloadScheduleHandler struct {
	ScheduleHandler scheduleHandler.IScheduleHandler
	config.Config
}

type WorkloadHandler interface {
	AdjustReplicas(_workloadSchedule workloadschedulerv1.WorkloadScheduleData, r client.Writer, ctx context.Context, processedWorkloads map[string]string)
}

type IWorkloadScheduleHandler interface {
	EvaluateWorkloadSchedulers(schedulers *workloadschedulerv1.WorkloadScheduleList, r client.Reader, ctx context.Context) (map[string][]workloadschedulerv1.Schedule, map[string]workloadschedulerv1.WorkloadSchedule)
	ProcessWorkloadSchedules(schedules map[string][]workloadschedulerv1.Schedule, schedulerMap map[string]workloadschedulerv1.WorkloadSchedule, c client.Client, ctx context.Context) error
	ValidateWorkloadSchedule(schedule *workloadschedulerv1.WorkloadSchedule, r client.Reader) error
}

type DeploymentHandler struct {
	config.Config
	apps.Deployment
}

func NewDeploymentHandler(deployment *apps.Deployment) *DeploymentHandler {
	return &DeploymentHandler{Config: *config.New(), Deployment: *deployment}
}

type StatefulSetHandler struct {
	config.Config
	apps.StatefulSet
}

func NewStatefulSetHandler(statefulSet *apps.StatefulSet) *StatefulSetHandler {
	return &StatefulSetHandler{Config: *config.New(), StatefulSet: *statefulSet}
}

func New() *WorkloadScheduleHandler {
	return &WorkloadScheduleHandler{ScheduleHandler: scheduleHandler.New(), Config: *config.New()}
}

func (w *WorkloadScheduleHandler) ValidateWorkloadSchedule(workloadSchedule *workloadschedulerv1.WorkloadSchedule, r client.Reader) error {
	if workloadSchedule.Spec.Schedules == nil || len(workloadSchedule.Spec.Schedules) == 0 {
		return fmt.Errorf("schedules need to be defined")

	} else {
		for _, schedule := range workloadSchedule.Spec.Schedules {
			if errs := validation.IsDNS1123Label(schedule.Schedule); errs != nil {
				return fmt.Errorf("schedule: %s is not valid. %v", schedule.Schedule, errs)
			}

			_, err := w.ScheduleHandler.GetScheduleByName(schedule.Schedule, r, context.Background())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *WorkloadScheduleHandler) EvaluateWorkloadSchedulers(workloadSchedulers *workloadschedulerv1.WorkloadScheduleList, r client.Reader, ctx context.Context) (map[string][]workloadschedulerv1.Schedule, map[string]workloadschedulerv1.WorkloadSchedule) {
	var workloadSchedulerAndSchedules = make(map[string][]workloadschedulerv1.Schedule)
	var workloadSchedulerMap = make(map[string]workloadschedulerv1.WorkloadSchedule)
	for _, workloadScheduleItem := range workloadSchedulers.Items {
		workloadSchedule := workloadScheduleItem.Spec
		schedules, err := w.ScheduleHandler.FetchWorkloadSchedules(workloadSchedule.Schedules, r, ctx)
		if err != nil {
			log.Log.Error(err, fmt.Sprintf("skipped, error occurred when validating schedules for %s.", workloadScheduleItem.Name))
			continue
		}
		workloadSchedulerMap[workloadScheduleItem.Name] = workloadScheduleItem
		workloadSchedulerAndSchedules[workloadScheduleItem.Name] = schedules
	}
	return workloadSchedulerAndSchedules, workloadSchedulerMap
}

func (w *WorkloadScheduleHandler) ProcessWorkloadSchedules(_workloadScheduleAndSchedules map[string][]workloadschedulerv1.Schedule, workloadSchedulerMap map[string]workloadschedulerv1.WorkloadSchedule, r client.Client, ctx context.Context) error {
	var workloadScheduleAndSchedules = w.extractSchedulesOfInstant(_workloadScheduleAndSchedules, workloadSchedulerMap)
	var workloadSchedules = w.RankWorkloadScheduleBySelectors(workloadScheduleAndSchedules)

	return w.executeAction(workloadSchedules, r, ctx)
}

func (w *WorkloadScheduleHandler) BuildSpecMap(_workloadSchedule workloadschedulerv1.WorkloadSchedule, specMap map[string]map[string][]workloadschedulerv1.WorkloadScheduleData, schedule workloadschedulerv1.Schedule) {
	_workloadScheduleSelector := _workloadSchedule.Spec.Selector
	namespaces := _workloadScheduleSelector.Namespaces
	names := _workloadScheduleSelector.Names
	kinds := _workloadScheduleSelector.Kinds
	labels := _workloadScheduleSelector.Labels

	//used for ranking no of selectors, any other better way?
	keysArr := []rune("0000")
	if len(namespaces) != 0 {
		keysArr[0] = '1'
	} else {
		namespaces = []string{util.ALL}
	}

	if len(kinds) != 0 {
		keysArr[1] = '1'
	} else {
		kinds = []string{util.DEPLOYMENT, util.STATEFULSET}
	}

	if len(names) != 0 {
		keysArr[2] = '1'
	} else {
		names = []string{util.ALL}
	}

	if len(labels) != 0 {
		keysArr[3] = '1'
	}

	keyStr := string(keysArr)
	//
	var extractedKeys = cartesian.Product(namespaces, kinds, names)

	for _, _keyComb := range extractedKeys {
		keyComb := strings.Join(_keyComb, config.MapKeySeparator)

		if specMap[keyStr] == nil {
			specMap[keyStr] = make(map[string][]workloadschedulerv1.WorkloadScheduleData)
		}
		if desired, err := w.getDesired(schedule, _workloadSchedule.Spec.Schedules); err == nil {
			workloadScheduleData := workloadschedulerv1.WorkloadScheduleData{Labels: _workloadSchedule.Spec.Selector.Labels, WorkloadScheduler: _workloadSchedule.Name, Namespace: _keyComb[0], Kind: _keyComb[1], Name: _keyComb[2], Desired: desired}
			specMap[keyStr][keyComb] = append(specMap[keyStr][keyComb], workloadScheduleData)
		} else {
			log.Log.Error(err, "error occurred when matching schedules")
			break
		}
	}
}

func (w *WorkloadScheduleHandler) RankWorkloadScheduleBySelectors(_specMap map[string]map[string][]workloadschedulerv1.WorkloadScheduleData) []workloadschedulerv1.WorkloadScheduleData {
	var _workloadSchedules []workloadschedulerv1.WorkloadScheduleData

	if len(_specMap) == 0 {
		if w.Config.LookUpBooleanEnv(config.Debug) {
			log.Log.Info("no workload schedule found to match this instant")
		}
		return _workloadSchedules
	}

	var workloadSchedulers = make([]string, len(_specMap))
	for workloadScheduler := range _specMap {
		workloadSchedulers = append(workloadSchedulers, workloadScheduler)
	}

	sort.Strings(workloadSchedulers)

	for idx := len(workloadSchedulers) - 1; idx >= 0; idx-- {
		v := _specMap[workloadSchedulers[idx]]
		for _, workloadSchedule := range v {
			_workloadSchedules = append(_workloadSchedules, workloadSchedule...)
		}
	}

	return _workloadSchedules
}

func (d *DeploymentHandler) AdjustReplicas(_workloadSchedule workloadschedulerv1.WorkloadScheduleData, r client.Writer, ctx context.Context, processedWorkloads map[string]string) {
	deployment := d.Deployment
	processedWorkloadKey := fmt.Sprintf("%s/%s/%s", deployment.Namespace, util.DEPLOYMENT, deployment.Name)
	if _, ok := d.Config.GetIgnoredNamespacesMap()[deployment.Namespace]; !ok {
		if d.Config.LookUpBooleanEnv(config.Debug) {
			log.Log.Info(fmt.Sprintf("fetched: %s .... %v", processedWorkloadKey, processedWorkloads))
		}
		if _, ok := processedWorkloads[processedWorkloadKey]; !ok {
			deploymentSpec := deployment.Spec
			currentReplicaCount := *deploymentSpec.Replicas

			if currentReplicaCount != _workloadSchedule.Desired {
				//if d.Config.LookUpBooleanEnv(config.Debug) {
				log.Log.Info(fmt.Sprintf("%v updating NS: %v, Name: %v, from %v to %v", _workloadSchedule.WorkloadScheduler, deployment.Namespace, deployment.Name, currentReplicaCount, _workloadSchedule.Desired))
				//}
				*deploymentSpec.Replicas = _workloadSchedule.Desired
				err := r.Update(ctx, &deployment)
				if err != nil {
					log.Log.Error(err, fmt.Sprintf("failed to update %s from %d to %d for workloadschedule %s.", util.DEPLOYMENT, currentReplicaCount, _workloadSchedule.Desired, _workloadSchedule.WorkloadScheduler))
				} else {
					//if d.Config.LookUpBooleanEnv(config.Debug) {
					log.Log.Info(fmt.Sprintf("%v updated NS: %v, Name: %v, from %v to %v", _workloadSchedule.WorkloadScheduler, deployment.Namespace, deployment.Name, currentReplicaCount, _workloadSchedule.Desired))

					//}
				}
			} else {
				//if d.Config.LookUpBooleanEnv(config.Debug) {
				log.Log.Info(fmt.Sprintf("got %s %s in order with %s. Namespace: %s, Name: %s, Desired: %d", deployment.Name, util.DEPLOYMENT, _workloadSchedule.WorkloadScheduler, deployment.Namespace, deployment.Name, _workloadSchedule.Desired))
				//}
			}
			processedWorkloads[processedWorkloadKey] = processedWorkloadKey

		} else {
			if d.Config.LookUpBooleanEnv(config.Debug) {
				log.Log.Info(fmt.Sprintf("skipped in loop: NS: %s, Kind: %s, Name: %s.", _workloadSchedule.Namespace, _workloadSchedule.Kind, _workloadSchedule.Name))
			}
		}
	} else {
		if d.Config.LookUpBooleanEnv(config.Debug) {
			log.Log.Info(fmt.Sprintf("ignored workload in %s namespace.", deployment.Namespace))
		}
	}
}

func (w *StatefulSetHandler) AdjustReplicas(_workloadSchedule workloadschedulerv1.WorkloadScheduleData, r client.Writer, ctx context.Context, processedWorkloads map[string]string) {
	statefulSet := w.StatefulSet
	processedWorkloadKey := fmt.Sprintf("%s/%s/%s", statefulSet.Namespace, util.STATEFULSET, statefulSet.Name)
	if _, ok := w.Config.GetIgnoredNamespacesMap()[statefulSet.Namespace]; !ok {
		if w.Config.LookUpBooleanEnv(config.Debug) {
			log.Log.Info(fmt.Sprintf("fetched: %s .... %v", processedWorkloadKey, processedWorkloads))
		}
		if _, ok := processedWorkloads[processedWorkloadKey]; !ok {
			statefulSetSpec := statefulSet.Spec
			currentReplicaCount := *statefulSet.Spec.Replicas

			if *statefulSet.Spec.Replicas != _workloadSchedule.Desired {
				//if w.Config.LookUpBooleanEnv(config.Debug) {
				log.Log.Info(fmt.Sprintf("%v updating NS: %v, Name: %v, from %v to %v", _workloadSchedule.WorkloadScheduler, statefulSet.Namespace, statefulSet.Name, currentReplicaCount, _workloadSchedule.Desired))
				//}
				*statefulSetSpec.Replicas = _workloadSchedule.Desired
				err := r.Update(ctx, &statefulSet)
				if err != nil {
					log.Log.Error(err, fmt.Sprintf("failed to update %s from %d to %d for workloadschedule %s.", util.STATEFULSET, currentReplicaCount, _workloadSchedule.Desired, _workloadSchedule.WorkloadScheduler))
				} else {
					//if w.Config.LookUpBooleanEnv(config.Debug) {
					log.Log.Info(fmt.Sprintf("%v updated NS: %v, Name: %v, from %v to %v", _workloadSchedule.WorkloadScheduler, statefulSet.Namespace, statefulSet.Name, currentReplicaCount, _workloadSchedule.Desired))
					//}
				}
			} else {
				//if w.Config.LookUpBooleanEnv(config.Debug) {
				log.Log.Info(fmt.Sprintf("got %s %s in order with %s. Namespace: %s, Name: %s, Desired: %d", statefulSet.Name, util.STATEFULSET, _workloadSchedule.WorkloadScheduler, statefulSet.Namespace, statefulSet.Name, _workloadSchedule.Desired))
				//}
			}
			processedWorkloads[processedWorkloadKey] = processedWorkloadKey

		} else {
			if w.Config.LookUpBooleanEnv(config.Debug) {
				log.Log.Info(fmt.Sprintf("skipped in loop: NS: %s, Kind: %s, Name: %s", _workloadSchedule.Namespace, _workloadSchedule.Kind, _workloadSchedule.Name))
			}
		}
	} else {
		if w.Config.LookUpBooleanEnv(config.Debug) {
			log.Log.Info(fmt.Sprintf("Ignored workload in %s namespace", statefulSet.Namespace))
		}
	}
}

func (w *WorkloadScheduleHandler) AdjustReplicas(_workloadSchedule workloadschedulerv1.WorkloadScheduleData, r client.Writer, ctx context.Context, processedWorkloads map[string]string, handler WorkloadHandler) {
	handler.AdjustReplicas(_workloadSchedule, r, ctx, processedWorkloads)
}

func (w *WorkloadScheduleHandler) extractSchedulesOfInstant(_workloadScheduleAndSchedules map[string][]workloadschedulerv1.Schedule, workloadSchedulerMap map[string]workloadschedulerv1.WorkloadSchedule) map[string]map[string][]workloadschedulerv1.WorkloadScheduleData {
	var specMap = make(map[string]map[string][]workloadschedulerv1.WorkloadScheduleData)
	now := time.Now().In(time.Local)

	for workloadScheduleName, _schedules := range _workloadScheduleAndSchedules {
		if _workloadSchedule, ok := workloadSchedulerMap[workloadScheduleName]; ok {
			for _, schedule := range _schedules {
				scheduleSpec := schedule.Spec
				scheduleUnits := scheduleSpec.ScheduleUnits
				if scheduleUnits != nil {
					for idx, scheduleUnit := range scheduleUnits {
						if scheduleUnit.Days != nil && len(scheduleUnit.Days) > 0 {
							isThisDayIncluded := w.ScheduleHandler.IsThisDayIncluded(scheduleUnit.Days, now)
							if !isThisDayIncluded {
								delete(_workloadScheduleAndSchedules, workloadScheduleName)
								continue
							}
						}

						startTime, processTimeErr := util.ProcessScheduleTimeUnit(scheduleUnit.Start, now)
						if processTimeErr != nil || startTime.IsZero() {
							if processTimeErr != nil {
								log.Log.Error(processTimeErr, "startTime process time error.")
							}
							continue
						}

						endTime, processTimeErr := util.ProcessScheduleTimeUnit(scheduleUnit.End, now)
						if processTimeErr != nil || endTime.IsZero() {
							if processTimeErr != nil {
								log.Log.Error(processTimeErr, "endTime process time error.")
							}
							continue
						}

						if now.After(startTime) && now.Before(endTime) {
							w.BuildSpecMap(_workloadSchedule, specMap, schedule)
							break
						} else {
							if w.Config.LookUpBooleanEnv(config.Debug) {
								log.Log.Info(fmt.Sprintf("schedule %s for %s ws with scheduleUnit %d not valid for now: %s, start %s, end %s", schedule.Name, _workloadSchedule.Name, idx, now, startTime, endTime))
							}
						}
					}
				}
			}
		}
	}

	return specMap
}

func (w *WorkloadScheduleHandler) executeAction(_workloadSchedules []workloadschedulerv1.WorkloadScheduleData, r client.Client, ctx context.Context) error {
	processedWorkloads := make(map[string]string)

	for _, _workloadSchedule := range _workloadSchedules {
		namespace := _workloadSchedule.Namespace
		name := _workloadSchedule.Name
		kind := _workloadSchedule.Kind
		labels := _workloadSchedule.Labels

		keyComb := fmt.Sprintf("%s/%s/%s", namespace, kind, name)
		_, exists := processedWorkloads[keyComb]
		if !exists {
			if w.Config.LookUpBooleanEnv(config.Debug) {
				log.Log.Info(fmt.Sprintf("evaluating: %s,  NS: %s, Kind: %s, Name: %s", _workloadSchedule.WorkloadScheduler, namespace, kind, name))
			}
			opts := []client.ListOption{}
			if name != util.ALL {
				opts = append(opts, client.MatchingFields{config.IndexedField: name})
			}

			if namespace != util.ALL {
				opts = append(opts, client.InNamespace(namespace))
			}

			if labels != nil {
				opts = append(opts, client.MatchingLabels(_workloadSchedule.Labels))
			}

			if kind == util.DEPLOYMENT {
				w.executeActionOnDeployment(r, ctx, opts, _workloadSchedule, processedWorkloads)
			} else if kind == util.STATEFULSET {
				w.executeActionOnStatefulSet(r, ctx, opts, _workloadSchedule, processedWorkloads)
			}
		} else {
			if w.Config.LookUpBooleanEnv(config.Debug) {
				log.Log.Info(fmt.Sprintf("skipped: NS: %s, Kind: %s, Name: %s", namespace, kind, name))
			}
		}
	}

	return nil
}

func (w *WorkloadScheduleHandler) executeActionOnStatefulSet(r client.Client, ctx context.Context, opts []client.ListOption, _workloadSchedule workloadschedulerv1.WorkloadScheduleData, processedWorkloads map[string]string) {
	var statefulSetWorkloads apps.StatefulSetList
	err := r.List(ctx, &statefulSetWorkloads, opts...)
	if err != nil {
		log.Log.Error(err, fmt.Sprintf("error occurred when fetching %s", util.STATEFULSET))
	} else {
		for _, statefulSet := range statefulSetWorkloads.Items {
			w.AdjustReplicas(_workloadSchedule, r, ctx, processedWorkloads, NewStatefulSetHandler(&statefulSet))
		}
	}
}

func (w *WorkloadScheduleHandler) executeActionOnDeployment(r client.Client, ctx context.Context, opts []client.ListOption, _workloadSchedule workloadschedulerv1.WorkloadScheduleData, processedWorkloads map[string]string) {
	var deploymentWorkloads apps.DeploymentList
	err := r.List(ctx, &deploymentWorkloads, opts...)
	if err != nil {
		log.Log.Error(err, fmt.Sprintf("error occurred when fetching %s", util.DEPLOYMENT))
	} else {
		for _, deployment := range deploymentWorkloads.Items {
			w.AdjustReplicas(_workloadSchedule, r, ctx, processedWorkloads, NewDeploymentHandler(&deployment))
		}
	}
}

func (w *WorkloadScheduleHandler) getDesired(schedule workloadschedulerv1.Schedule, schedules []workloadschedulerv1.WorkloadScheduleUnit) (int32, error) {
	for _, workloadScheduleUnit := range schedules {
		if schedule.Name == workloadScheduleUnit.Schedule {
			return workloadScheduleUnit.Desired, nil
		}
	}
	return 0, fmt.Errorf("no matching schedule found for %s", schedule.Name)
}
