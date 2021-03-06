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

package util

import (
	"bytes"
	"io"
	"sort"

	"github.com/ghodss/yaml"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/kubectl/pkg/kinflate/gvkn"
)

// Decode decodes a list of objects in byte array format.
// Decoded object will be inserted in `into` if it's not nil. Otherwise, it will
// construct a new map and return it.
func Decode(in []byte, into map[gvkn.GroupVersionKindName]*unstructured.Unstructured) (map[gvkn.GroupVersionKindName]*unstructured.Unstructured, error) {
	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(in), 1024)
	objs := []*unstructured.Unstructured{}

	var err error
	for {
		var out unstructured.Unstructured
		err = decoder.Decode(&out)
		if err != nil {
			break
		}
		objs = append(objs, &out)
	}
	if err != io.EOF {
		return nil, err
	}

	if into == nil {
		into = map[gvkn.GroupVersionKindName]*unstructured.Unstructured{}
	}
	for i := range objs {
		metaAccessor, err := meta.Accessor(objs[i])
		if err != nil {
			return nil, err
		}
		name := metaAccessor.GetName()
		typeAccessor, err := meta.TypeAccessor(objs[i])
		if err != nil {
			return nil, err
		}
		apiVersion := typeAccessor.GetAPIVersion()
		kind := typeAccessor.GetKind()
		gv, err := schema.ParseGroupVersion(apiVersion)
		if err != nil {
			return nil, err
		}
		gvk := gv.WithKind(kind)
		gvkn := gvkn.GroupVersionKindName{
			GVK:  gvk,
			Name: name,
		}
		into[gvkn] = objs[i]
	}
	return into, nil
}

// Encode encodes the map `in` and output the encoded objects separated by `---`.
func Encode(in map[gvkn.GroupVersionKindName]*unstructured.Unstructured) ([]byte, error) {
	gvknList := []gvkn.GroupVersionKindName{}
	for gvkn := range in {
		gvknList = append(gvknList, gvkn)
	}
	sort.Sort(ByGVKN(gvknList))

	firstObj := true
	var b []byte
	buf := bytes.NewBuffer(b)
	for _, gvkn := range gvknList {
		obj := in[gvkn]
		out, err := yaml.Marshal(obj)
		if err != nil {
			return nil, err
		}
		if !firstObj {
			_, err = buf.WriteString("---\n")
			if err != nil {
				return nil, err
			}
		}
		_, err = buf.Write(out)
		if err != nil {
			return nil, err
		}
		firstObj = false
	}
	return buf.Bytes(), nil
}
