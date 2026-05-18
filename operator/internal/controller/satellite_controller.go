/*
Copyright 2026.

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
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	fleetv1alpha1 "github.com/ashimrai123/harbor-satellite-fleet/operator/api/v1alpha1"
	gcclient "github.com/ashimrai123/harbor-satellite-fleet/operator/internal/gcclient"
)

const (
	satelliteFinalizer  = "fleet.harbor.io/satellite-finalizer"
	requeueInterval     = 60 * time.Second
)

// SatelliteReconciler reconciles a Satellite object.
type SatelliteReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=fleet.harbor.io,resources=satellites,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fleet.harbor.io,resources=satellites/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=fleet.harbor.io,resources=satellites/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch

// Reconcile reads the state of the Satellite CR and reconciles it with Ground Control.
func (r *SatelliteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the Satellite CR
	satellite := &fleetv1alpha1.Satellite{}
	if err := r.Get(ctx, req.NamespacedName, satellite); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Build the Ground Control client from the referenced Secret
	gcClient, err := r.buildGCClient(ctx, satellite)
	if err != nil {
		logger.Error(err, "Failed to build Ground Control client")
		r.setCondition(satellite, fleetv1alpha1.ConditionReady, metav1.ConditionFalse,
			"CredentialsError", fmt.Sprintf("Failed to read credentials: %v", err))
		_ = r.Status().Update(ctx, satellite)
		return ctrl.Result{RequeueAfter: requeueInterval}, nil
	}

	// Handle deletion via finalizer
	if !satellite.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, satellite, gcClient)
	}

	// Add finalizer if missing
	if !controllerutil.ContainsFinalizer(satellite, satelliteFinalizer) {
		controllerutil.AddFinalizer(satellite, satelliteFinalizer)
		if err := r.Update(ctx, satellite); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Set initial phase
	if satellite.Status.Phase == "" {
		satellite.Status.Phase = fleetv1alpha1.SatellitePhasePending
		_ = r.Status().Update(ctx, satellite)
	}

	// Step 1: Ensure satellite is registered in Ground Control
	if err := r.ensureRegistered(ctx, satellite, gcClient); err != nil {
		logger.Error(err, "Failed to register satellite")
		return ctrl.Result{RequeueAfter: requeueInterval}, nil
	}

	// Step 2: Ensure groups are synced
	if err := r.ensureGroupsSynced(ctx, satellite, gcClient); err != nil {
		logger.Error(err, "Failed to sync groups")
		return ctrl.Result{RequeueAfter: requeueInterval}, nil
	}

	// Step 3: Ensure config is assigned
	if err := r.ensureConfigAssigned(ctx, satellite, gcClient); err != nil {
		logger.Error(err, "Failed to assign config")
		return ctrl.Result{RequeueAfter: requeueInterval}, nil
	}

	// All done — set Ready
	satellite.Status.Phase = fleetv1alpha1.SatellitePhaseReady
	now := metav1.Now()
	satellite.Status.LastSyncTime = &now
	r.setCondition(satellite, fleetv1alpha1.ConditionReady, metav1.ConditionTrue,
		"ReconcileSuccess", "Satellite is fully reconciled")

	if err := r.Status().Update(ctx, satellite); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Satellite reconciled successfully", "name", satellite.Name)
	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

// ensureRegistered checks if the satellite exists in GC and registers it if not.
func (r *SatelliteReconciler) ensureRegistered(ctx context.Context, sat *fleetv1alpha1.Satellite, gc *gcclient.Client) error {
	logger := log.FromContext(ctx)

	// Check if already registered
	existing, err := gc.GetSatellite(sat.Name)
	if err == nil {
		// Already registered — update the ID in status
		sat.Status.GroundControlID = existing.ID
		r.setCondition(sat, fleetv1alpha1.ConditionRegistered, metav1.ConditionTrue,
			"AlreadyRegistered", "Satellite is registered in Ground Control")
		return r.Status().Update(ctx, sat)
	}

	// Register the satellite
	logger.Info("Registering satellite in Ground Control", "name", sat.Name)
	sat.Status.Phase = fleetv1alpha1.SatellitePhaseRegistering
	_ = r.Status().Update(ctx, sat)

	resp, err := gc.RegisterSatellite(gcclient.RegisterSatelliteRequest{
		Name:       sat.Name,
		ConfigName: sat.Spec.ConfigName,
		Groups:     sat.Spec.Groups,
	})
	if err != nil {
		r.setCondition(sat, fleetv1alpha1.ConditionRegistered, metav1.ConditionFalse,
			"RegistrationFailed", err.Error())
		_ = r.Status().Update(ctx, sat)
		return fmt.Errorf("registration failed: %w", err)
	}

	// Store the ZTR token in a Secret
	tokenSecretName := sat.Spec.TokenSecretName
	if tokenSecretName == "" {
		tokenSecretName = sat.Name + "-token"
	}
	if err := r.storeToken(ctx, sat, tokenSecretName, resp.Token); err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	sat.Status.TokenSecret = tokenSecretName
	r.setCondition(sat, fleetv1alpha1.ConditionRegistered, metav1.ConditionTrue,
		"Registered", "Satellite registered successfully")
	return r.Status().Update(ctx, sat)
}

// ensureGroupsSynced makes sure the satellite is in the correct groups.
func (r *SatelliteReconciler) ensureGroupsSynced(ctx context.Context, sat *fleetv1alpha1.Satellite, gc *gcclient.Client) error {
	if len(sat.Spec.Groups) == 0 {
		r.setCondition(sat, fleetv1alpha1.ConditionGroupsSynced, metav1.ConditionTrue,
			"NoGroups", "No groups specified")
		return r.Status().Update(ctx, sat)
	}

	for _, group := range sat.Spec.Groups {
		if err := gc.AddSatelliteToGroup(sat.Name, group); err != nil {
			r.setCondition(sat, fleetv1alpha1.ConditionGroupsSynced, metav1.ConditionFalse,
				"GroupSyncFailed", fmt.Sprintf("Failed to add to group %s: %v", group, err))
			_ = r.Status().Update(ctx, sat)
			return err
		}
	}

	r.setCondition(sat, fleetv1alpha1.ConditionGroupsSynced, metav1.ConditionTrue,
		"GroupsSynced", "Satellite is in all specified groups")
	return r.Status().Update(ctx, sat)
}

// ensureConfigAssigned assigns the config to the satellite.
func (r *SatelliteReconciler) ensureConfigAssigned(ctx context.Context, sat *fleetv1alpha1.Satellite, gc *gcclient.Client) error {
	if err := gc.SetSatelliteConfig(sat.Name, sat.Spec.ConfigName); err != nil {
		r.setCondition(sat, fleetv1alpha1.ConditionConfigAssigned, metav1.ConditionFalse,
			"ConfigAssignFailed", err.Error())
		_ = r.Status().Update(ctx, sat)
		return err
	}

	r.setCondition(sat, fleetv1alpha1.ConditionConfigAssigned, metav1.ConditionTrue,
		"ConfigAssigned", "Config assigned to satellite")
	return r.Status().Update(ctx, sat)
}

// handleDeletion removes the satellite from Ground Control and removes the finalizer.
func (r *SatelliteReconciler) handleDeletion(ctx context.Context, sat *fleetv1alpha1.Satellite, gc *gcclient.Client) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if controllerutil.ContainsFinalizer(sat, satelliteFinalizer) {
		logger.Info("Deleting satellite from Ground Control", "name", sat.Name)

		if err := gc.DeleteSatellite(sat.Name); err != nil {
			logger.Error(err, "Failed to delete satellite from Ground Control")
			// Don't block deletion if GC is unreachable
		}

		controllerutil.RemoveFinalizer(sat, satelliteFinalizer)
		if err := r.Update(ctx, sat); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// buildGCClient reads credentials from the referenced Secret and builds a GC client.
func (r *SatelliteReconciler) buildGCClient(ctx context.Context, sat *fleetv1alpha1.Satellite) (*gcclient.Client, error) {
	secret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      sat.Spec.SecretRef,
		Namespace: sat.Namespace,
	}, secret); err != nil {
		return nil, fmt.Errorf("secret %q not found: %w", sat.Spec.SecretRef, err)
	}

	username := string(secret.Data["username"])
	password := string(secret.Data["password"])
	if username == "" || password == "" {
		return nil, fmt.Errorf("secret %q must have 'username' and 'password' keys", sat.Spec.SecretRef)
	}

	c := gcclient.NewClient(sat.Spec.GroundControlURL, "")
	if err := c.Login(username, password); err != nil {
		return nil, fmt.Errorf("Ground Control login failed: %w", err)
	}

	return c, nil
}

// storeToken creates or updates a Secret to hold the ZTR bootstrap token.
func (r *SatelliteReconciler) storeToken(ctx context.Context, sat *fleetv1alpha1.Satellite, secretName, token string) error {
	tokenSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: sat.Namespace,
		},
		StringData: map[string]string{
			"token": token,
		},
	}
	// Set owner reference so the secret is garbage collected with the CR
	if err := controllerutil.SetControllerReference(sat, tokenSecret, r.Scheme); err != nil {
		return err
	}

	existing := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: sat.Namespace}, existing)
	if errors.IsNotFound(err) {
		return r.Create(ctx, tokenSecret)
	}
	if err != nil {
		return err
	}

	existing.StringData = tokenSecret.StringData
	return r.Update(ctx, existing)
}

// setCondition is a helper to update a specific condition on the status.
func (r *SatelliteReconciler) setCondition(sat *fleetv1alpha1.Satellite, condType string, status metav1.ConditionStatus, reason, message string) {
	meta.SetStatusCondition(&sat.Status.Conditions, metav1.Condition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})
}

// SetupWithManager sets up the controller with the Manager.
func (r *SatelliteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fleetv1alpha1.Satellite{}).
		Named("satellite").
		Complete(r)
}
