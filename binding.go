package binding

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/jsonpath"
)

type PodTemplateSpecMapping struct {
	// Containers defines mappings for containers.
	Containers []ContainerMapping
	// Volumes is a JSON Pointer to the field holding the container's environment variables. The
	// referenced value must be `[]corev1.Volume` on the discovered container. If the value
	// does not exist it will be created.
	Volumes string
}

func (m *PodTemplateSpecMapping) Default() {
	if m.Volumes == "" {
		m.Volumes = "/spec/template/spec/volumes"
	}
	if len(m.Containers) == 0 {
		m.Containers = []ContainerMapping{
			{
				Path: ".spec.template.spec.containers[*]",
				Name: "/name",
			},
			{
				Path: ".spec.template.spec.initContainers[*]",
				Name: "/name",
			},
		}
	}
	for i := range m.Containers {
		m.Containers[i].Default()
	}
}

type ContainerMapping struct {
	// Path is a JSONPath query for containers on the resource. The query is executed
	// from the root of the object and is not required to return any results.
	Path string
	// Name is a JSON Pointer to the field holding the container's name. If specified, the
	// referenced value must be `string` on the discovered container.
	// +optional
	Name string
	// Env is a JSON Pointer to the field holding the container's environment variables. The
	// referenced value must be `[]corev1.EnvVar` on the discovered container. If the value
	// does not exist it will be created.
	// +optional
	Env string
	// VolumeMounts is a JSON Pointer to the field holding the container's environment variables. The
	// referenced value must be `[]corev1.VolumeMount` on the discovered container. If the value
	// does not exist it will be created.
	VolumeMounts string
}

func (m *ContainerMapping) Default() {
	if m.Env == "" {
		m.Env = "/env"
	}
	if m.VolumeMounts == "" {
		m.Env = "/volumeMounts"
	}
}

type Binding struct {
	// Name of the binding. Shows as the path of the volume mount within the container.
	Name string
	// Secret holding the credentails to bind
	Secret corev1.LocalObjectReference
	// Containers is a set of names of container to bind. If empty, all containers are bound.
	Containers []string
}

func (b *Binding) Bind(obj metav1.Object, m *PodTemplateSpecMapping) error {
	// convert structured type to unstructured
	u, err := runtime.DefaultUnstructuredConverter.
		ToUnstructured(obj)
	if err != nil {
		return err
	}

	for _, containerMapping := range m.Containers {
		cp := jsonpath.New("")
		if err := cp.Parse(fmt.Sprintf("{%s}", containerMapping.Path)); err != nil {
			return err
		}
		cv, err := cp.FindResults(u)
		if err != nil {
			// errors are expected if a path is not found
			continue
		}
		for i := range cv[0] {
			c := cv[0][i].Interface().(map[string]interface{})
			// TODO find/inject SERVICE_BINDING_ROOT
			// for now set the container image to prove we can manipulate the resource
			c["image"] = "world"
		}
	}

	// mutate original object with binding content from unstructured
	return runtime.DefaultUnstructuredConverter.
		FromUnstructured(u, obj)
}
