package main

import (
	"fmt"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Kube mg.Namespace

// Deploy deploys the application to Kubernetes using Helm
func (Kube) Deploy() error {
	fmt.Println("Deploying application to Kubernetes using Helm...")

	// Define Helm release name and chart path
	releaseName := "golang-todo-app"
	chartPath := "../helm-chart"

	// Run Helm upgrade/install command
	return sh.RunV("helm", "upgrade", "--install", releaseName, chartPath, "--namespace", "default", "--create-namespace")
}
