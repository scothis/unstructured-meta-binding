package binding

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestBinding(t *testing.T) {
	tests := []struct {
		name        string
		binding     Binding
		mapping     PodTemplateSpecMapping
		seed        client.Object
		expected    client.Object
		expectedErr bool
	}{
		{
			name:    "podspecable",
			binding: Binding{},
			mapping: PodTemplateSpecMapping{},
			seed: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Name: "init-hello",
								},
								{
									Name: "init-hello-2",
								},
							},
							Containers: []corev1.Container{
								{
									Name: "hello",
								},
								{
									Name: "hello-2",
								},
							},
						},
					},
				},
			},
			expected: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Name:  "init-hello",
									Image: "world",
								},
								{
									Name:  "init-hello-2",
									Image: "world",
								},
							},
							Containers: []corev1.Container{
								{
									Name:  "hello",
									Image: "world",
								},
								{
									Name:  "hello-2",
									Image: "world",
								},
							},
						},
					},
				},
			},
		},
		{
			name:    "almost podspecable",
			binding: Binding{},
			mapping: PodTemplateSpecMapping{
				Containers: []ContainerMapping{
					{
						Path: ".spec.jobTemplate.spec.template.spec.containers[*]",
						// Name: "/name",
					},
					{
						Path: ".spec.jobTemplate.spec.template.spec.initContainers[*]",
						// Name: "/name",
					},
				},
			},
			seed: &batchv1.CronJob{
				Spec: batchv1.CronJobSpec{
					JobTemplate: batchv1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									InitContainers: []corev1.Container{
										{
											Name: "init-hello",
										},
										{
											Name: "init-hello-2",
										},
									},
									Containers: []corev1.Container{
										{
											Name: "hello",
										},
										{
											Name: "hello-2",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &batchv1.CronJob{
				Spec: batchv1.CronJobSpec{
					JobTemplate: batchv1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									InitContainers: []corev1.Container{
										{
											Name:  "init-hello",
											Image: "world",
										},
										{
											Name:  "init-hello-2",
											Image: "world",
										},
									},
									Containers: []corev1.Container{
										{
											Name:  "hello",
											Image: "world",
										},
										{
											Name:  "hello-2",
											Image: "world",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:     "no containers",
			binding:  Binding{},
			mapping:  PodTemplateSpecMapping{},
			seed:     &appsv1.Deployment{},
			expected: &appsv1.Deployment{},
		},
		{
			name:    "invalid container jsonpath",
			binding: Binding{},
			mapping: PodTemplateSpecMapping{
				Containers: []ContainerMapping{
					{
						Path: "[",
					},
				},
			},
			seed:        &appsv1.Deployment{},
			expectedErr: true,
		},
		{
			name:        "conversion error",
			binding:     Binding{},
			mapping:     PodTemplateSpecMapping{},
			seed:        &BadMarshalJSON{},
			expectedErr: true,
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			actual := c.seed.DeepCopyObject().(client.Object)
			m := &c.mapping
			m.Default()
			err := c.binding.Bind(actual, m)

			if (err != nil) != c.expectedErr {
				t.Errorf("Bind() expected err: %v", err)
			}
			if c.expectedErr {
				return
			}
			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("Bind() (-expected, +actual): %s", diff)
			}
		})
	}
}

var (
	_ client.Object = (*BadMarshalJSON)(nil)
)

type BadMarshalJSON struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

func (r *BadMarshalJSON) MarshalJSON() ([]byte, error)   { return nil, fmt.Errorf("bad json marshal") }
func (r *BadMarshalJSON) DeepCopyObject() runtime.Object { return r }
