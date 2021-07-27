package binding

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/jsonpath"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PodMapping struct {
	// Annotations is a JSON Pointer to the field holding the container's environment variables. The
	// referenced value must be `map[string]string` on the discovered container. If the value
	// does not exist it will be created.
	Annotations string
	// Containers defines mappings for containers.
	Containers []ContainerMapping
	// Volumes is a JSON Pointer to the field holding the container's environment variables. The
	// referenced value must be `[]corev1.Volume` on the discovered container. If the value
	// does not exist it will be created.
	Volumes string
}

func (m *PodMapping) Default() {
	if m.Annotations == "" {
		m.Annotations = "/spec/template/metadata/annotations"
	}
	if len(m.Containers) == 0 {
		m.Containers = []ContainerMapping{
			{
				Path: ".spec.template.spec.initContainers[*]",
				Name: "/name",
			},
			{
				Path: ".spec.template.spec.containers[*]",
				Name: "/name",
			},
		}
	}
	for i := range m.Containers {
		m.Containers[i].Default()
	}
	if m.Volumes == "" {
		m.Volumes = "/spec/template/spec/volumes"
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
	// +optional
	VolumeMounts string
}

func (m *ContainerMapping) Default() {
	if m.Env == "" {
		m.Env = "/env"
	}
	if m.VolumeMounts == "" {
		m.VolumeMounts = "/volumeMounts"
	}
}

func (m *PodMapping) ToMeta(obj client.Object) (MetaPodTemplate, error) {
	mpt := MetaPodTemplate{
		Annotations: map[string]string{},
		Containers:  []MetaContainer{},
		Volumes:     []corev1.Volume{},
	}

	u, err := runtime.DefaultUnstructuredConverter.
		ToUnstructured(obj)
	if err != nil {
		return mpt, err
	}
	uv := reflect.ValueOf(u)

	if err := m.getAt(m.Annotations, uv, &mpt.Annotations); err != nil {
		return mpt, err
	}
	for i := range m.Containers {
		cp := jsonpath.New("")
		if err := cp.Parse(fmt.Sprintf("{%s}", m.Containers[i].Path)); err != nil {
			return mpt, err
		}
		cr, err := cp.FindResults(u)
		if err != nil {
			// errors are expected if a path is not found
			continue
		}
		for _, cv := range cr[0] {
			mc := MetaContainer{
				Name:         "",
				Env:          []corev1.EnvVar{},
				VolumeMounts: []corev1.VolumeMount{},
			}

			if m.Containers[i].Name != "" {
				// name is optional
				if err := m.getAt(m.Containers[i].Name, cv, &mc.Name); err != nil {
					return mpt, err
				}
			}
			if err := m.getAt(m.Containers[i].Env, cv, &mc.Env); err != nil {
				return mpt, err
			}
			if err := m.getAt(m.Containers[i].VolumeMounts, cv, &mc.VolumeMounts); err != nil {
				return mpt, err
			}

			mpt.Containers = append(mpt.Containers, mc)
		}
	}
	if err := m.getAt(m.Volumes, uv, &mpt.Volumes); err != nil {
		return mpt, err
	}

	return mpt, nil
}

func (m *PodMapping) FromMeta(obj client.Object, mpt MetaPodTemplate) error {
	// convert structured type to unstructured
	u, err := runtime.DefaultUnstructuredConverter.
		ToUnstructured(obj)
	if err != nil {
		return err
	}
	uv := reflect.ValueOf(u)

	if err := m.setAt(m.Annotations, &mpt.Annotations, uv); err != nil {
		return err
	}
	ci := 0
	for i := range m.Containers {
		cp := jsonpath.New("")
		if err := cp.Parse(fmt.Sprintf("{%s}", m.Containers[i].Path)); err != nil {
			return err
		}
		cr, err := cp.FindResults(u)
		if err != nil {
			// errors are expected if a path is not found
			continue
		}
		for _, cv := range cr[0] {
			if m.Containers[i].Name != "" {
				if err := m.setAt(m.Containers[i].Name, &mpt.Containers[ci].Name, cv); err != nil {
					return err
				}
			}
			if err := m.setAt(m.Containers[i].Env, &mpt.Containers[ci].Env, cv); err != nil {
				return err
			}
			if err := m.setAt(m.Containers[i].VolumeMounts, &mpt.Containers[ci].VolumeMounts, cv); err != nil {
				return err
			}

			ci++
		}
	}
	if err := m.setAt(m.Volumes, &mpt.Volumes, uv); err != nil {
		return err
	}

	// mutate original object with binding content from unstructured
	return runtime.DefaultUnstructuredConverter.
		FromUnstructured(u, obj)
}

func (m *PodMapping) getAt(ptr string, source reflect.Value, target interface{}) error {
	parent := reflect.ValueOf(nil)
	createIfNil := false
	v, _, _, err := m.find(source, parent, m.keys(ptr), "", createIfNil)
	if err != nil {
		return err
	}
	if !v.IsValid() || v.IsNil() {
		return nil
	}
	b, err := json.Marshal(v.Interface())
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}

func (m *PodMapping) setAt(ptr string, value interface{}, target reflect.Value) error {
	keys := m.keys(ptr)
	parent := reflect.ValueOf(nil)
	createIfNil := true
	_, vp, lk, err := m.find(target, parent, keys, "", createIfNil)
	if err != nil {
		return err
	}
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var out interface{}
	switch reflect.ValueOf(value).Elem().Kind() {
	case reflect.Map:
		out = map[string]interface{}{}
	case reflect.Slice:
		out = []interface{}{}
	case reflect.String:
		out = ""
	default:
		return fmt.Errorf("unsupported kind %s", reflect.ValueOf(value).Kind())
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	vp.SetMapIndex(reflect.ValueOf(lk), reflect.ValueOf(out))
	return nil
}

func (m *PodMapping) keys(ptr string) []string {
	// TODO use a real json pointer parser, this does not support escaped sequences
	ptr = strings.TrimPrefix(ptr, "/")
	return strings.Split(ptr, "/")
}

func (m *PodMapping) find(value, parent reflect.Value, keys []string, lastKey string, createIfNil bool) (reflect.Value, reflect.Value, string, error) {
	if !value.IsValid() || value.IsNil() {
		if !createIfNil {
			return reflect.ValueOf(nil), reflect.ValueOf(nil), "", nil
		}
		value = reflect.ValueOf(make(map[string]interface{}))
		parent.SetMapIndex(reflect.ValueOf(lastKey), value)
	}
	if len(keys) == 0 {
		return value, parent, lastKey, nil
	}
	switch value.Kind() {
	case reflect.Map:
		lastKey = keys[0]
		keys = keys[1:]
		parent = value
		value = value.MapIndex(reflect.ValueOf(lastKey))
		return m.find(value, parent, keys, lastKey, createIfNil)
	case reflect.Interface:
		parent = value
		value = value.Elem()
		return m.find(value, parent, keys, lastKey, createIfNil)
	default:
		return reflect.ValueOf(nil), parent, lastKey, fmt.Errorf("unhandled kind %q", value.Kind())
	}
}
