package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

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
			return err
		} else {
			logrus.Infof("resource %s applied succesfully!", obj.GetName())
		}
	}
	return nil
}

func (e *Executor) KubectlApply(filePath string) error {

	var err error
	var payload []byte

	for retry := 0; retry <= 10; retry++ {
		payload, err = os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("cannot read manifest file %s: %v", filePath, err)
		}

		objects := getUnstructuredFromYAML(string(payload))
		err = e.apply(context.TODO(), objects)
		if err == nil {
			// no error exit early
			break
		}

		time.Sleep(5 * time.Second)
	}
	return err
}
