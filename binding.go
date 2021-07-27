package binding

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Binding struct {
	// Name of the binding. Shows as the path of the volume mount within the container.
	Name string
	// Secret holding the credentails to bind
	Secret corev1.LocalObjectReference
	// Containers is a set of names of container to bind. If empty, all containers are bound.
	Containers []string
}

func (b *Binding) Bind(obj runtime.Object, m *PodMapping) error {
	mpt, err := m.ToMeta(obj)
	if err != nil {
		return err
	}
	// TODO inject volume
	for i := range mpt.Containers {
		c := &mpt.Containers[i]
		// TODO skip container if not allowed
		serviceBindingRoot := ""
		for _, e := range c.Env {
			if e.Name == "SERVICE_BINDING_ROOT" {
				serviceBindingRoot = e.Value
				break
			}
		}
		if serviceBindingRoot == "" {
			serviceBindingRoot = "/bindings"
			c.Env = append(c.Env, corev1.EnvVar{
				Name:  "SERVICE_BINDING_ROOT",
				Value: serviceBindingRoot,
			})
		}
		// TODO do other stuff with the container
	}
	return m.FromMeta(obj, mpt)
}
