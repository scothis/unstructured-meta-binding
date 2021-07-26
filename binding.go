package binding

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/jsonpath"
)

type Binding struct {
	// Name of the binding. Shows as the path of the volume mount within the container.
	Name string
	// Secret holding the credentails to bind
	Secret corev1.LocalObjectReference
	// Containers is a set of names of container to bind. If empty, all containers are bound.
	Containers []string
}

func (b *Binding) Bind(obj metav1.Object, m *PodMapping) error {
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
