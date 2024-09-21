package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kyaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

const (
	DEFAULT_NAMESPACE = "default"
	YAML_DELIMITER    = "---"
)

func (e *Executor) getRESTMapper() (meta.RESTMapper, error) {

	groupResources, err := restmapper.GetAPIGroupResources(e.DiscoveryClient)
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	return mapper, nil
}

func getNamespace(obj *unstructured.Unstructured, mapping *meta.RESTMapping) string {
	// Default to "default" namespace if not specified
	namespace := obj.GetNamespace()
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace && namespace == "" {
		namespace = DEFAULT_NAMESPACE
	}
	return namespace
}

func getUnstructuredFromYAML(payload string) []*unstructured.Unstructured {

	var objects []*unstructured.Unstructured
	var err error

	data := strings.Split(payload, YAML_DELIMITER)
	decUnstructured := kyaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	for _, manifest := range data {
		unstructObject := &unstructured.Unstructured{}
		_, _, err = decUnstructured.Decode([]byte(manifest), nil, unstructObject)
		if err != nil {
			logrus.Warningln("failed to parse/load manifest")
			logrus.Debugln("failed to parse/load manifest: \n", manifest)
			continue
		}
		objects = append(objects, unstructObject)
	}

	return objects
}

// Implementation of Kubernetes server side apply
func (e *Executor) apply(ctx context.Context, objects []*unstructured.Unstructured) error {

	var dr dynamic.ResourceInterface

	for _, obj := range objects {

		groupResources, err := restmapper.GetAPIGroupResources(e.DiscoveryClient)
		if err != nil {
			logrus.Warnf("failed to get rest mapper: %v", err)
			continue
		}

		mapper := restmapper.NewDiscoveryRESTMapper(groupResources)
		mapping, err := mapper.RESTMapping(schema.ParseGroupKind(obj.GroupVersionKind().GroupKind().String()))
		if err != nil {
			logrus.Warnf("failed to get rest mapping: %v", err)
			continue
		}

		namespace := getNamespace(obj, mapping)

		dr = e.DynamicClient.Resource(mapping.Resource)
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			dr = e.DynamicClient.Resource(mapping.Resource).Namespace(namespace)
		}

		// Check if namespace is empty and if resource is namespaced or not
		data, err := json.Marshal(obj)
		if err != nil {
			logrus.Warnln("failed to unmarshal object:", obj.GetName())
			continue
		}
		_, err = dr.Patch(
			ctx,
			obj.GetName(),
			types.ApplyPatchType,
			data,
			metav1.PatchOptions{
				FieldManager: "bootstrap-manager",
			},
		)
		if err != nil {
			logrus.Warnln("failed to apply resource:", obj.GetName(), err)
			continue
		} else {
			logrus.Infof("resource %s applied succesfully!", obj.GetName())
		}
	}

	return nil
}

func (e *Executor) KubectlApply(filePath string) error {

	yamlPayload, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot read manifest file %s: %v", filePath, err)
	}

	objects := getUnstructuredFromYAML(string(yamlPayload))
	return e.apply(context.TODO(), objects)
}
