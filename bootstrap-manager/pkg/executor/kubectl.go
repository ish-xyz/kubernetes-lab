package executor

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (e *Executor) KubectlApply(filePath string) error {
	// Read the YAML file
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot read manifest file %s: %v", filePath, err)
	}

	// Split the YAML file into multiple resources
	resources := strings.Split(string(yamlFile), "---")

	for _, resourceYAML := range resources {
		// Skip empty resources
		if strings.TrimSpace(resourceYAML) == "" {
			continue
		}

		// Parse the YAML to an unstructured object
		var resource unstructured.Unstructured
		err = yaml.Unmarshal([]byte(resourceYAML), &resource)
		if err != nil {
			logrus.Warningf("Failed to unmarshal yaml: %v", err)
			continue
		}

		// Get the GVK (Group, Version, Kind) for the resource
		gvk := resource.GroupVersionKind()

		// Use the RESTMapper to get the GroupVersionResource
		mapping, err := e.RestMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			logrus.Warningf("Error mapping resource: %v\n", err)
			continue
		}

		// Get the API resource
		gvr := schema.GroupVersionResource{
			Group:    mapping.Resource.Group,
			Version:  mapping.Resource.Version,
			Resource: mapping.Resource.Resource,
		}

		// Apply the resource
		_, err = e.DynamicClient.Resource(gvr).Namespace(resource.GetNamespace()).Create(context.TODO(), &resource, metav1.CreateOptions{})
		if err != nil {
			logrus.Warningf("Error applying resource: %v\n", err)
		} else {
			logrus.Infof("Resource %s/%s applied\n", resource.GetKind(), resource.GetName())
		}
	}

	return nil
}
