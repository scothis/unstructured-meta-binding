package binding

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestMapping_ToMeta(t *testing.T) {
	testAnnotations := map[string]string{
		"key": "value",
	}
	testEnv := corev1.EnvVar{
		Name:  "NAME",
		Value: "value",
	}
	testVolume := corev1.Volume{
		Name: "name",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: "my-secret",
			},
		},
	}
	testVolumeMount := corev1.VolumeMount{
		Name:      "name",
		MountPath: "/mount/path",
	}

	tests := []struct {
		name        string
		mapping     PodMapping
		seed        client.Object
		expected    MetaPodTemplate
		expectedErr bool
	}{
		{
			name:    "podspecable",
			mapping: PodMapping{},
			seed: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: testAnnotations,
						},
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
									Name:         "hello",
									Env:          []corev1.EnvVar{testEnv},
									VolumeMounts: []corev1.VolumeMount{testVolumeMount},
								},
								{
									Name: "hello-2",
								},
							},
							Volumes: []corev1.Volume{testVolume},
						},
					},
				},
			},
			expected: MetaPodTemplate{
				Annotations: testAnnotations,
				Containers: []MetaContainer{
					{
						Name:         "init-hello",
						Env:          []corev1.EnvVar{},
						VolumeMounts: []corev1.VolumeMount{},
					},
					{
						Name:         "init-hello-2",
						Env:          []corev1.EnvVar{},
						VolumeMounts: []corev1.VolumeMount{},
					},
					{
						Name:         "hello",
						Env:          []corev1.EnvVar{testEnv},
						VolumeMounts: []corev1.VolumeMount{testVolumeMount},
					},
					{
						Name:         "hello-2",
						Env:          []corev1.EnvVar{},
						VolumeMounts: []corev1.VolumeMount{},
					},
				},
				Volumes: []corev1.Volume{testVolume},
			},
		},
		{
			name: "almost podspecable",
			mapping: PodMapping{
				Annotations: "/spec/jobTemplate/spec/template/metadata/annotations",
				Containers: []ContainerMapping{
					{
						Path: ".spec.jobTemplate.spec.template.spec.initContainers[*]",
						Name: "/name",
					},
					{
						Path: ".spec.jobTemplate.spec.template.spec.containers[*]",
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
								ObjectMeta: metav1.ObjectMeta{
									Annotations: testAnnotations,
								},
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
											Name:         "hello",
											Env:          []corev1.EnvVar{testEnv},
											VolumeMounts: []corev1.VolumeMount{testVolumeMount},
										},
										{
											Name: "hello-2",
										},
									},
									Volumes: []corev1.Volume{testVolume},
								},
							},
						},
					},
				},
			},
			expected: MetaPodTemplate{
				Annotations: testAnnotations,
				Containers: []MetaContainer{
					{
						Name:         "init-hello",
						Env:          []corev1.EnvVar{},
						VolumeMounts: []corev1.VolumeMount{},
					},
					{
						Name:         "init-hello-2",
						Env:          []corev1.EnvVar{},
						VolumeMounts: []corev1.VolumeMount{},
					},
					{
						Name:         "hello",
						Env:          []corev1.EnvVar{testEnv},
						VolumeMounts: []corev1.VolumeMount{testVolumeMount},
					},
					{
						Name:         "hello-2",
						Env:          []corev1.EnvVar{},
						VolumeMounts: []corev1.VolumeMount{},
					},
				},
				Volumes: []corev1.Volume{testVolume},
			},
		},
		{
			name:    "no containers",
			mapping: PodMapping{},
			seed:    &appsv1.Deployment{},
			expected: MetaPodTemplate{
				Annotations: map[string]string{},
				Containers:  []MetaContainer{},
				Volumes:     []corev1.Volume{},
			},
		},
		{
			name:    "empty container",
			mapping: PodMapping{},
			seed: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{},
							},
						},
					},
				},
			},
			expected: MetaPodTemplate{
				Annotations: map[string]string{},
				Containers: []MetaContainer{
					{
						Name:         "",
						Env:          []corev1.EnvVar{},
						VolumeMounts: []corev1.VolumeMount{},
					},
				},
				Volumes: []corev1.Volume{},
			},
		},
		{
			name: "misaligned path",
			mapping: PodMapping{
				Containers: []ContainerMapping{
					{
						Path: ".foo.bar",
					},
				},
			},
			seed: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: testAnnotations,
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "hello",
									Env: []corev1.EnvVar{
										testEnv,
									},
									VolumeMounts: []corev1.VolumeMount{
										testVolumeMount,
									},
								},
							},
							Volumes: []corev1.Volume{
								testVolume,
							},
						},
					},
				},
			},
			expected: MetaPodTemplate{
				Annotations: testAnnotations,
				Containers:  []MetaContainer{},
				Volumes: []corev1.Volume{
					testVolume,
				},
			},
		},
		{
			name: "misaligned pointers",
			mapping: PodMapping{
				Annotations: "/foo/nar",
				Containers: []ContainerMapping{
					{
						Path:         ".spec.template.spec.containers[*]",
						Name:         "/foo/bar",
						Env:          "/foo/bar",
						VolumeMounts: "/foo/bar",
					},
				},
				Volumes: "/foo/bar",
			},
			seed: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: testAnnotations,
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:         "hello",
									Env:          []corev1.EnvVar{testEnv},
									VolumeMounts: []corev1.VolumeMount{testVolumeMount},
								},
							},
							Volumes: []corev1.Volume{
								testVolume,
							},
						},
					},
				},
			},
			expected: MetaPodTemplate{
				Annotations: map[string]string{},
				Containers: []MetaContainer{
					{
						Name:         "",
						Env:          []corev1.EnvVar{},
						VolumeMounts: []corev1.VolumeMount{},
					},
				},
				Volumes: []corev1.Volume{},
			},
		},
		{
			name: "invalid container jsonpath",
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
			mapping:     PodMapping{},
			seed:        &BadMarshalJSON{},
			expectedErr: true,
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			m := &c.mapping
			m.Default()
			actual, err := m.ToMeta(c.seed)

			if c.expectedErr && err == nil {
				t.Errorf("ToMeta() expected to err")
			} else if !c.expectedErr && err != nil {
				t.Errorf("ToMeta() unexpected err: %v", err)
			}
			if c.expectedErr {
				return
			}
			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("ToMeta() (-expected, +actual): %s", diff)
			}
		})
	}
}
