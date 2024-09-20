package executor

func (e *Executor) KubectlApply() {
	return
}

/*

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/yaml"
)

func main() {
	// Set up the client configuration
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}

	// Create a dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// Create a discovery client
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		panic(err)
	}

	// Create a RESTMapper
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient))

	// Read the YAML file
	yamlFile, err := os.ReadFile("path/to/your/yaml/file.yaml")
	if err != nil {
		panic(err)
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
			fmt.Printf("Error parsing resource: %v\n", err)
			continue
		}

		// Get the GVK (Group, Version, Kind) for the resource
		gvk := resource.GroupVersionKind()

		// Use the RESTMapper to get the GroupVersionResource
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			fmt.Printf("Error mapping resource: %v\n", err)
			continue
		}

		// Get the API resource
		gvr := schema.GroupVersionResource{
			Group:    mapping.Resource.Group,
			Version:  mapping.Resource.Version,
			Resource: mapping.Resource.Resource,
		}

		// Apply the resource
		_, err = dynamicClient.Resource(gvr).Namespace(resource.GetNamespace()).Create(context.TODO(), &resource, metav1.CreateOptions{})
		if err != nil {
			// Handle error (create update logic here)
			fmt.Printf("Error applying resource: %v\n", err)
		} else {
			fmt.Printf("Resource %s/%s applied\n", resource.GetKind(), resource.GetName())
		}
	}
}

*/
