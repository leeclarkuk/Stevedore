// Package render turns the Stevedore model into Kubernetes manifests laid out
// as a Kustomize base.
package render

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/leeclarkuk/stevedore/internal/model"
)

// File is one rendered manifest. Render returns files in a fixed order so
// callers can write or compare them without sorting.
type File struct {
	Path    string
	Content []byte
}

// kustomization is written by hand rather than importing the kustomize API
// module, which would pull in a large dependency tree for four fields.
type kustomization struct {
	APIVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	Resources  []string `json:"resources"`
}

// Render produces a Kustomize base for app: one Deployment per service and a
// kustomization.yaml listing them.
func Render(app model.App) ([]File, error) {
	var files []File
	var resources []string

	for _, svc := range app.Services {
		path := svc.Name + "-deployment.yaml"
		content, err := marshal(deployment(app, svc))
		if err != nil {
			return nil, fmt.Errorf("render service %q: %w", svc.Name, err)
		}
		files = append(files, File{Path: path, Content: content})
		resources = append(resources, path)
	}

	content, err := marshal(kustomization{
		APIVersion: "kustomize.config.k8s.io/v1beta1",
		Kind:       "Kustomization",
		Resources:  resources,
	})
	if err != nil {
		return nil, fmt.Errorf("render kustomization: %w", err)
	}
	files = append(files, File{Path: "kustomization.yaml", Content: content})
	return files, nil
}

func deployment(app model.App, svc model.Service) *appsv1.Deployment {
	selector := map[string]string{
		"app.kubernetes.io/name":     svc.Name,
		"app.kubernetes.io/instance": app.Name,
	}
	labels := map[string]string{
		"app.kubernetes.io/name":       svc.Name,
		"app.kubernetes.io/instance":   app.Name,
		"app.kubernetes.io/managed-by": "stevedore",
	}
	replicas := int32(1)

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:   svc.Name,
			Labels: labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: selector},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  svc.Name,
						Image: svc.Image,
					}},
				},
			},
		},
	}
}

// marshal serialises obj as YAML. sigs.k8s.io/yaml goes through JSON, which
// sorts map keys, so label order is stable.
func marshal(obj any) ([]byte, error) {
	return yaml.Marshal(obj)
}
