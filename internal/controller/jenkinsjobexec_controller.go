/*
Copyright 2024.

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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	jenkinsv1 "github.com/rajiv2205/sample-jenkins-jobexec-operator/api/v1"
	jenutil "github.com/rajiv2205/sample-jenkins-jobexec-operator/pkg/jenkinsutil"
)

// JenkinsJobExecReconciler reconciles a JenkinsJobExec object
type JenkinsJobExecReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=jenkins.operatortest.io,resources=jenkinsjobexecs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=jenkins.operatortest.io,resources=jenkinsjobexecs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=jenkins.operatortest.io,resources=jenkinsjobexecs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the JenkinsJobExec object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *JenkinsJobExecReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	jenkinsJob := &jenkinsv1.JenkinsJobExec{}

	// Fetch the Jenkins instance
	if err := r.Get(ctx, req.NamespacedName, jenkinsJob); err != nil {
		// Jenkins Job object not found, it might have been deleted
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if jenkinsJob.Status.Processed {
		log.Log.Info("Skipping already processed resource", "name", jenkinsJob.Name)
		return ctrl.Result{}, nil
	}

	// Fetch the ConfigMap referenced in the CR
	configMap := &corev1.ConfigMap{}
	configMapNamespace := jenkinsJob.Namespace
	if jenkinsJob.Spec.ConfigMapRef.Namespace != "" {
		configMapNamespace = jenkinsJob.Spec.ConfigMapRef.Namespace
	}
	configMapName := types.NamespacedName{Name: jenkinsJob.Spec.ConfigMapRef.Name, Namespace: configMapNamespace}

	if err := r.Get(ctx, configMapName, configMap); err != nil {
		log.Log.Error(err, "unable to fetch ConfigMap", "configMap", configMapName)
		return ctrl.Result{}, err
	}

	// Use the ConfigMap data
	jenkinsURL := configMap.Data["jenkinsURL"]

	// Fetch the Secret referenced in the CR
	secret := &corev1.Secret{}
	secretNamespace := jenkinsJob.Namespace
	if jenkinsJob.Spec.SecretRef.Namespace != "" {
		secretNamespace = jenkinsJob.Spec.SecretRef.Namespace
	}
	secretName := types.NamespacedName{Name: jenkinsJob.Spec.SecretRef.Name, Namespace: secretNamespace}

	if err := r.Get(ctx, secretName, secret); err != nil {
		log.Log.Error(err, "unable to fetch Secret", "secret", secretName)
		return ctrl.Result{}, err
	}

	// Use the secret data
	username := string(secret.Data["username"])
	token := string(secret.Data["token"])

	log.Log.Info(fmt.Sprintf("Jenkins Job to be executed: %s\n", jenkinsJob.Spec.JobName))

	queueBuildID, err := jenutil.TriggerJobWithAndWithoutParams(jenkinsURL, jenkinsJob.Spec.JobName, jenkinsJob.Spec.Parameters, username, token)
	if err != nil {
		log.Log.Info("Error triggering job: reqeueuing with a delay of 10 sec: ", err)
		return ctrl.Result{Requeue: true,
			RequeueAfter: time.Second * 10}, nil
	}

	log.Log.Info(fmt.Sprintf("queue build Id is: %d", queueBuildID))

	jobURL, err := jenutil.PollQueueBuild(jenkinsURL, jenkinsJob.Spec.JobName, queueBuildID, username, token)
	if err != nil {
		log.Log.Info(fmt.Sprintf("Looks like user cancelled the job from Jenkins, received status: %s", jobURL))
		jenkinsJob.Status.JobStatus = jobURL
		jenkinsJob.Status.Processed = true
		if err := r.Status().Update(ctx, jenkinsJob); err != nil {
			log.Log.Error(err, "unable to update Cancelled JenkinsJob status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	strippedJobURL := jobURL[1 : len(jobURL)-1]

	log.Log.Info(fmt.Sprintf("Build Job URL: %s", strippedJobURL))
	//status, jobbuildID, _, err := jenutil.PollBuildStatus(strippedJobURL, username, token)
	status, _, _, err := jenutil.PollBuildStatus(strippedJobURL, username, token)
	log.Log.Info(fmt.Sprintf("Build %s status: %s\n", strippedJobURL, status))
	jenkinsJob.Status.BuildURL = strippedJobURL
	jenkinsJob.Status.JobStatus = status
	jenkinsJob.Status.Processed = true
	if err := r.Status().Update(ctx, jenkinsJob); err != nil {
		log.Log.Error(err, "unable to update JenkinsJob status")
		return ctrl.Result{}, err
	}
	log.Log.Info(fmt.Sprintf("Updated the build status url: %s status: %s\n", strippedJobURL, status))
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JenkinsJobExecReconciler) SetupWithManager(mgr ctrl.Manager) error {
	createPredicate := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&jenkinsv1.JenkinsJobExec{}).
		WithEventFilter(createPredicate).
		Complete(r)
}
