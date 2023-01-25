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

package controllers

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// SecretReconciler reconciles a Secret object
type SecretReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.padok.fr,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.padok.fr,resources=secrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.padok.fr,resources=secrets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Secret object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
var c controller.Controller
var mgr manager.Manager
var globalLog = logf.Log.WithName("global")
var secretnames []string

func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	secret := &corev1.Secret{}
	err := r.Get(ctx, req.NamespacedName, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			// Pod not found
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	var pods corev1.PodList
	if err := r.Client.List(ctx, &pods, client.InNamespace(req.Namespace)); err != nil {
		log.Println("Failed to fetch deployment resources")
		return ctrl.Result{}, err
	}
	isInList := false
	i := secret.Name
	for _, v := range secretnames {
		if i == v {
			isInList = true
			break
		}
	}
	if isInList {
		fmt.Println(secret.Name)
		re1 := regexp.MustCompile(secret.Name)
		for _, pod := range pods.Items {
			if len(pod.Spec.Containers[0].EnvFrom) > 0 {
				envfrom1 := len(pod.Spec.Containers[0].EnvFrom)
				for i := 0; i < envfrom1; i++ {
					envfrom := pod.Spec.Containers[0].EnvFrom[i].String()
					if len(re1.FindString(envfrom)) > 0 {
						if secret.Name == pod.Spec.Containers[0].EnvFrom[i].SecretRef.Name {
							log.Println("pod name: ", pod.Name, pod.Spec.Containers[0].EnvFrom[i].SecretRef.Name)
							r.Delete(ctx, &pod)
						}
					}
				}
			}
			if len(pod.Spec.Containers[0].Env) > 0 {
				envl := (len(pod.Spec.Containers[0].Env))
				for i := 0; i < envl; i++ {
					env2 := pod.Spec.Containers[0].Env[i].String()
					if (len(re1.FindString(env2))) > 0 {
						if secret.Name == pod.Spec.Containers[0].Env[i].ValueFrom.SecretKeyRef.Name {
							log.Println("pod name: ", pod.Name, pod.Spec.Containers[0].Env[i].ValueFrom.SecretKeyRef.Name)
							r.Delete(ctx, &pod)
						}
					}
				}
			}

			if len(pod.Spec.Volumes) > 0 {
				vol := len(pod.Spec.Volumes)
				for i := 0; i < vol; i++ {
					vol1 := pod.Spec.Volumes[i].String()
					if (len(re1.FindString(vol1))) > 0 {
						if secret.Name == pod.Spec.Volumes[i].Secret.SecretName {
							log.Println("pod name: ", pod.Name, pod.Spec.Volumes[i].Secret.SecretName)
							r.Delete(ctx, &pod)
						}

					}
				}
			}
		}
	} else {
		secretnames = append(secretnames, secret.Name)
	}
	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		Complete(r)
}
