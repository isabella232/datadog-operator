// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package datadogmonitor

import (
	"context"
	"fmt"

	datadoghqv1alpha1 "github.com/DataDog/datadog-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/DataDog/datadog-operator/pkg/controller/utils"
)

const (
	datadogMonitorFinalizer = "finalizer.monitor.datadoghq.com"
)

func (r *Reconciler) handleFinalizer(logger logr.Logger, dm *datadoghqv1alpha1.DatadogMonitor) (ctrl.Result, error) {
	// Check if the DatadogMonitor instance is marked to be deleted, which is indicated by the deletion timestamp being set
	if dm.GetDeletionTimestamp() != nil {
		if utils.ContainsString(dm.GetFinalizers(), datadogMonitorFinalizer) {
			r.finalizeDatadogMonitor(logger, dm)

			dm.SetFinalizers(utils.RemoveString(dm.GetFinalizers(), datadogMonitorFinalizer))
			err := r.Client.Update(context.TODO(), dm)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Add finalizer for this resource if it doesn't already exist
	if !utils.ContainsString(dm.GetFinalizers(), datadogMonitorFinalizer) {
		if err := r.addFinalizer(logger, dm); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true, RequeueAfter: defaultRequeuePeriod}, nil
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) finalizeDatadogMonitor(logger logr.Logger, dm *datadoghqv1alpha1.DatadogMonitor) {
	if dm.Status.Primary {
		err := r.deleteMonitor(dm.Status.ID)
		if err != nil {
			logger.Error(err, "failed to finalize monitor")
			return
		}
	}
	logger.Info("Successfully finalized DatadogMonitor", "Monitor ID", fmt.Sprint(dm.Status.ID))
}

func (r *Reconciler) addFinalizer(logger logr.Logger, dm *datadoghqv1alpha1.DatadogMonitor) error {
	logger.Info("Adding Finalizer for the DatadogMonitor")

	dm.SetFinalizers(append(dm.GetFinalizers(), datadogMonitorFinalizer))

	err := r.Client.Update(context.TODO(), dm)
	if err != nil {
		logger.Error(err, "Failed to update DatadogMonitor with finalizer")
		return err
	}
	return nil
}
