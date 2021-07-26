package binding

import (
	corev1 "k8s.io/api/core/v1"
)

type MetaPodTemplate struct {
	Annotations map[string]string
	Containers  []MetaContainer
	Volumes     []corev1.Volume
}

type MetaContainer struct {
	Name         string
	Env          []corev1.EnvVar
	VolumeMounts []corev1.VolumeMount
}
