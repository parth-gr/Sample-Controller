package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	//Go Lang library to send messages to Slack via Incoming Webhooks.

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var resource map[string]string

// type SlackRequestBody struct {
// 	Text string `json: "text"`
// }

func slackNotification(podName *v1.Pod, container string) {
	// webhookURL := "https://hooks.slack.com/services/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"

	// attachment1 := slack.Attachment{}
	// attachment1.AddField(slack.Field{Title: "Pod Name", Value: podName.Name}).AddField(slack.Field{Title: "Container Name", Value: container})
	// payload := slack.Payload{
	// 	Text:        "Pod Crash Notification",
	// 	Username:    "Kube Bot",
	// 	Channel:     "#kubernetes-demo",
	// 	IconEmoji:   ":monkey_face",
	// 	Attachments: []slack.Attachment{attachment1},
	// }
	// err := slack.Send(webhookURL, "", payload)
	// if len(err) > 0 {
	// 	fmt.Printf("error: %s\n", err)
	// }
	slack, err := exec.Command("/bin/sh", "slack.sh").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(slack)
	log.Printf("completed")
}

//this is controller struct like in Kubebuilder
type ReconciledPod struct {
	Client client.Client
	Logger logr.Logger
}

func (r *ReconciledPod) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log := r.Logger
	pod := &corev1.Pod{}
	//getting the object to be reconciled
	err := r.Client.Get(context.Background(), request.NamespacedName, pod)
	if errors.IsNotFound(err) {
		log.Error(err, "Pod Not Found. could have been deleted")
		return reconcile.Result{}, nil
	}
	if err != nil {
		log.Error(err, "Error fetching pod. Going to requeue")
		return reconcile.Result{Requeue: true}, err
	}
	if _, ok := resource[pod.Name]; !ok {
		resource[pod.Name] = pod.ResourceVersion
	} else if pod.ResourceVersion != resource[pod.Name] {
		log.Info("Reconciling Container: " + pod.Name + "Check Slack Notification")
		slackNotification(pod, pod.Name)
	}

	return reconcile.Result{}, nil
}

func main() {
	log := zapr.NewLogger(zap.NewExample()).WithName("pod-crash-controller")
	resource = make(map[string]string)

	log.Info("Setting up manager")
	//here we are not setting manager options like passing it a scheme
	//manager is the one who triggers controller and provides it resources
	//it also is responsible for graceful shutdowns using SIGKILL, SIGINT etc
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Error(err, "Unable to setup manager. Please check if KUBECONFIG is available")
		os.Exit(1)
	}

	log.Info("Setting up controller")
	ctrl, err := controller.New("pod-crash-controller", mgr, controller.Options{
		Reconciler: &ReconciledPod{Client: mgr.GetClient(), Logger: log},
	})
	if err != nil {
		log.Error(err, "Failed to setup controller")
		os.Exit(1)
	}

	log.Info("Watching Pods")
	if err := ctrl.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForObject{}); err != nil {
		log.Error(err, "Failed to watch pods")
		os.Exit(1)
	}

	log.Info("Starting the manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Failed to start manager")
		os.Exit(1)
	}
}

// kill port sudo kill -9 `sudo lsof -t -i:8080`
