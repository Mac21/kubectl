/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package transformers

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubectl/pkg/kinflate/gvkn"
)

func makeConfigMap(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": name,
			},
		},
	}
}

func makeConfigMaps(name1InGVKN, name2InGVKN, name1InObj, name2InObj string) map[gvkn.GroupVersionKindName]*unstructured.Unstructured {
	cm1 := makeConfigMap(name1InObj)
	cm2 := makeConfigMap(name2InObj)
	return map[gvkn.GroupVersionKindName]*unstructured.Unstructured{
		{
			GVK:  schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"},
			Name: name1InGVKN,
		}: cm1,
		{
			GVK:  schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"},
			Name: name2InGVKN,
		}: cm2,
	}
}

func compareMap(m1, m2 map[gvkn.GroupVersionKindName]*unstructured.Unstructured) error {
	if len(m1) != len(m2) {
		keySet1 := []gvkn.GroupVersionKindName{}
		keySet2 := []gvkn.GroupVersionKindName{}
		for GVKn := range m1 {
			keySet1 = append(keySet1, GVKn)
		}
		for GVKn := range m1 {
			keySet2 = append(keySet2, GVKn)
		}
		return fmt.Errorf("maps has different number of entries: %#v doesn't equals %#v", keySet1, keySet2)
	}
	for GVKn, obj1 := range m1 {
		obj2, found := m2[GVKn]
		if !found {
			return fmt.Errorf("%#v doesn't exist in %#v", GVKn, m2)
		}
		if !reflect.DeepEqual(obj1, obj2) {
			return fmt.Errorf("%#v doesn't match %#v", obj1, obj2)
		}
	}
	return nil
}
