package main

import (
	"context"
	"fmt"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	notificationv1 "github.com/fluxcd/notification-controller/api/v1"
	notificationv1b2 "github.com/fluxcd/notification-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/gofiber/fiber/v2"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"log"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const (
	targetNamespace    = "notifications-test"
	providerName       = "webhook-test"
	alertName          = "alert-test"
	helmRepositoryName = "bitnami"
	helmReleaseName    = "nginx"
)

func main() {
	app := fiber.New()

	app.Get("/test", testHandler)
	app.Get("/healthy", healthyHandler)
	app.Get("/ready", readyHandler)

	app.Post("/install", handleInstall)
	app.Delete("/uninstall", handleUninstall)

	app.Post("/installNotifications", handleInstallNotifications)
	app.Delete("/uninstallNotifications", handleUninstallNotifications)

	log.Fatal(app.Listen(":8888"))
}

func handleInstall(fiberCtx *fiber.Ctx) error {
	// register the GitOps Toolkit schema definitions
	scheme := runtime.NewScheme()
	_ = sourcev1.AddToScheme(scheme)
	_ = helmv2.AddToScheme(scheme)

	// init Kubernetes client
	kubeClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		panic(err)
	}

	// set a deadline for the Kubernetes API operations
	ctx, cancel := context.WithTimeout(fiberCtx.Context(), 60*time.Second)
	defer cancel()

	// create a Helm repository pointing to Bitnami
	helmRepository := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      helmReleaseName,
			Namespace: targetNamespace,
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL: "https://charts.bitnami.com/bitnami",
			Interval: metav1.Duration{
				Duration: 30 * time.Minute,
			},
		},
	}
	if err := kubeClient.Create(ctx, helmRepository); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("HelmRepository bitnami created")
	}

	// create a Helm release for nginx
	helmRelease := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      helmReleaseName,
			Namespace: targetNamespace,
		},
		Spec: helmv2.HelmReleaseSpec{
			ReleaseName: "nginx",
			Interval: metav1.Duration{
				Duration: 5 * time.Minute,
			},
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:   "nginx",
					Version: "8.x",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind: sourcev1.HelmRepositoryKind,
						Name: helmRepositoryName,
					},
				},
			},
			Values: &apiextensionsv1.JSON{Raw: []byte(`{"service": {"type": "ClusterIP"}}`)},
		},
	}

	if err := kubeClient.Create(ctx, helmRelease); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("HelmRelease nginx created")
	}
	return fiberCtx.SendStatus(200)
}

func handleUninstall(fiberCtx *fiber.Ctx) error {
	// register the GitOps Toolkit schema definitions
	scheme := runtime.NewScheme()
	_ = sourcev1.AddToScheme(scheme)
	_ = helmv2.AddToScheme(scheme)

	// init Kubernetes client
	kubeClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		panic(err)
	}

	// set a deadline for the Kubernetes API operations
	ctx, cancel := context.WithTimeout(fiberCtx.Context(), 60*time.Second)
	defer cancel()

	helmRepository := &unstructured.Unstructured{}
	_ = kubeClient.Get(context.Background(), client.ObjectKey{
		Namespace: targetNamespace,
		Name:      helmRepositoryName,
	}, helmRepository)

	if err := kubeClient.Delete(ctx, helmRepository); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("HelmRepository deleted")
	}

	helmRelease := &unstructured.Unstructured{}
	_ = kubeClient.Get(context.Background(), client.ObjectKey{
		Namespace: targetNamespace,
		Name:      helmReleaseName,
	}, helmRepository)
	if err := kubeClient.Create(ctx, helmRelease); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("HelmRelease deleted")
	}

	return fiberCtx.SendStatus(200)
}

func handleInstallNotifications(fiberCtx *fiber.Ctx) error {
	// register the GitOps Toolkit schema definitions
	scheme := runtime.NewScheme()
	_ = notificationv1.AddToScheme(scheme)

	// init Kubernetes client
	kubeClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		panic(err)
	}

	// set a deadline for the Kubernetes API operations
	ctx, cancel := context.WithTimeout(fiberCtx.Context(), 60*time.Second)
	defer cancel()

	provider := &notificationv1b2.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      providerName,
			Namespace: targetNamespace,
		},
		Spec: notificationv1b2.ProviderSpec{
			Type:    notificationv1b2.GenericProvider,
			Address: "http://192.168.105.7:31759/notifications",
		},
	}

	if err := kubeClient.Create(ctx, provider); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Provider created")
	}

	eventSources := []notificationv1.CrossNamespaceObjectReference{
		{
			Kind: kustomizev1.KustomizationKind,
			Name: "*",
		},
	}

	alert := &notificationv1b2.Alert{
		ObjectMeta: metav1.ObjectMeta{
			Name:      alertName,
			Namespace: targetNamespace,
		},
		Spec: notificationv1b2.AlertSpec{
			Summary: "teeeest",
			ProviderRef: meta.LocalObjectReference{
				Name: providerName,
			},
			EventSeverity: "info",
			EventSources:  eventSources,
		},
	}
	if err := kubeClient.Create(ctx, alert); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Alert created")
	}

	return fiberCtx.SendStatus(200)
}

func handleUninstallNotifications(fiberCtx *fiber.Ctx) error {
	// register the GitOps Toolkit schema definitions
	scheme := runtime.NewScheme()
	_ = notificationv1.AddToScheme(scheme)

	// init Kubernetes client
	kubeClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		panic(err)
	}

	// set a deadline for the Kubernetes API operations
	ctx, cancel := context.WithTimeout(fiberCtx.Context(), 60*time.Second)
	defer cancel()

	provider := &unstructured.Unstructured{}
	_ = kubeClient.Get(context.Background(), client.ObjectKey{
		Namespace: targetNamespace,
		Name:      providerName,
	}, provider)

	if err := kubeClient.Delete(ctx, provider); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Provider deleted")
	}

	alert := &unstructured.Unstructured{}
	_ = kubeClient.Get(context.Background(), client.ObjectKey{
		Namespace: targetNamespace,
		Name:      alertName,
	}, provider)

	if err := kubeClient.Delete(ctx, alert); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Alert deleted")
	}

	return fiberCtx.SendStatus(200)
}

func healthyHandler(ctx *fiber.Ctx) error {
	return ctx.SendStatus(200)
}

func readyHandler(ctx *fiber.Ctx) error {
	return ctx.SendStatus(200)
}

func testHandler(ctx *fiber.Ctx) error {
	agent := fiber.Get("http://notification-receiver.notifications-test/ready")
	s, r, e := agent.Bytes()
	fmt.Printf("---%v, ----%v, ----%e\n", s, r, e)
	return ctx.SendStatus(200)
}
