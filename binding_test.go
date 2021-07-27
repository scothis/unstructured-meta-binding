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
)

func TestBinding(t *testing.T) {
	tests := []struct {
		name        string
		binding     Binding
		mapping     PodMapping
		seed        runtime.Object
		expected    runtime.Object
		expectedErr bool
	}{
		{
			name:    "podspecable",
			binding: Binding{},
			mapping: PodMapping{},
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
									Env: []corev1.EnvVar{
										{
											Name:  "SERVICE_BINDING_ROOT",
											Value: "/custom/path",
										},
									},
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
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{},
						},
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Name: "init-hello",
									Env: []corev1.EnvVar{
										{
											Name:  "SERVICE_BINDING_ROOT",
											Value: "/bindings",
										},
									},
									VolumeMounts: []corev1.VolumeMount{},
								},
								{
									Name: "init-hello-2",
									Env: []corev1.EnvVar{
										{
											Name:  "SERVICE_BINDING_ROOT",
											Value: "/bindings",
										},
									},
									VolumeMounts: []corev1.VolumeMount{},
								},
							},
							Containers: []corev1.Container{
								{
									Name: "hello",
									Env: []corev1.EnvVar{
										{
											Name:  "SERVICE_BINDING_ROOT",
											Value: "/custom/path",
										},
									},
									VolumeMounts: []corev1.VolumeMount{},
								},
								{
									Name: "hello-2",
									Env: []corev1.EnvVar{
										{
											Name:  "SERVICE_BINDING_ROOT",
											Value: "/bindings",
										},
									},
									VolumeMounts: []corev1.VolumeMount{},
								},
							},
							Volumes: []corev1.Volume{},
						},
					},
				},
			},
		},
		{
			name:    "almost podspecable",
			binding: Binding{},
			mapping: PodMapping{
				Annotations: "/spec/jobTemplate/spec/template/metadata/annotations",
				Containers: []ContainerMapping{
					{
						Path: ".spec.jobTemplate.spec.template.spec.containers[*]",
						Name: "/name",
					},
					{
						Path: ".spec.jobTemplate.spec.template.spec.initContainers[*]",
						Name: "/name",
					},
				},
				Volumes: "/spec/jobTemplate/spec/template/spec/volumes",
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
											Env: []corev1.EnvVar{
												{
													Name:  "SERVICE_BINDING_ROOT",
													Value: "/custom/path",
												},
											},
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
								ObjectMeta: metav1.ObjectMeta{
									Annotations: map[string]string{},
								},
								Spec: corev1.PodSpec{
									InitContainers: []corev1.Container{
										{
											Name: "init-hello",
											Env: []corev1.EnvVar{
												{
													Name:  "SERVICE_BINDING_ROOT",
													Value: "/bindings",
												},
											},
											VolumeMounts: []corev1.VolumeMount{},
										},
										{
											Name: "init-hello-2",
											Env: []corev1.EnvVar{
												{
													Name:  "SERVICE_BINDING_ROOT",
													Value: "/bindings",
												},
											},
											VolumeMounts: []corev1.VolumeMount{},
										},
									},
									Containers: []corev1.Container{
										{
											Name: "hello",
											Env: []corev1.EnvVar{
												{
													Name:  "SERVICE_BINDING_ROOT",
													Value: "/custom/path",
												},
											},
											VolumeMounts: []corev1.VolumeMount{},
										},
										{
											Name: "hello-2",
											Env: []corev1.EnvVar{
												{
													Name:  "SERVICE_BINDING_ROOT",
													Value: "/bindings",
												},
											},
											VolumeMounts: []corev1.VolumeMount{},
										},
									},
									Volumes: []corev1.Volume{},
								},
							},
						},
					},
				},
			},
		},
		{
			name:    "no containers",
			binding: Binding{},
			mapping: PodMapping{},
			seed:    &appsv1.Deployment{},
			expected: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{},
						},
						Spec: corev1.PodSpec{
							Volumes: []corev1.Volume{},
						},
					},
				},
			},
		},
		{
			name:    "invalid container jsonpath",
			binding: Binding{},
			mapping: PodMapping{
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
			mapping:     PodMapping{},
			seed:        &BadMarshalJSON{},
			expectedErr: true,
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			actual := c.seed.DeepCopyObject()
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
	_ runtime.Object = (*BadMarshalJSON)(nil)
)

type BadMarshalJSON struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

func (r *BadMarshalJSON) MarshalJSON() ([]byte, error)   { return nil, fmt.Errorf("bad json marshal") }
func (r *BadMarshalJSON) DeepCopyObject() runtime.Object { return r }
