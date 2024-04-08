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

package converter

import (
	"fmt"

	"k8s.io/klog/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func convertCRD(Object *unstructured.Unstructured, toVersion string) (*unstructured.Unstructured, metav1.Status) {
	klog.V(2).Info("converting crd")

	convertedObject := Object.DeepCopy()
	fromVersion := Object.GetAPIVersion()

	if toVersion == fromVersion {
		return nil, statusErrorWithMessage("conversion from a version to itself should not call the webhook: %s", toVersion)
	}
	switch Object.GetAPIVersion() {
	case "gateway.kyma-project.io/v1beta1":
		switch toVersion {
		case "gateway.kyma-project.io/v1beta2":
			fmt.Printf("%s/%s: converting v1beta1->v1beta2\n", convertedObject.GetNamespace(), convertedObject.GetName())
			annotations := convertedObject.GetAnnotations()
			annotations["gateway.kyma-project.io/converted-to-v1beta2"] = "true"
			convertedObject.SetAnnotations(annotations)
			rules, _, err := unstructured.NestedFieldNoCopy(convertedObject.Object, "spec", "rules")
			if err != nil {
				return nil, statusErrorWithMessage("failed to get rules field")
			}
			rulesSlice, ok := rules.([]interface{})
			if !ok {
				return nil, statusErrorWithMessage("rules field is not a slice")
			}
			for _, rule := range rulesSlice {
				noAuth := false
				ruleMap, ok := rule.(map[string]interface{})
				if !ok {
					return nil, statusErrorWithMessage("rule field is not an object")
				}
				accessStrategies := ruleMap["accessStrategies"]
				if accessStrategies != nil {
					accessStrategiesSlice, ok := accessStrategies.([]interface{})
					if !ok {
						return nil, statusErrorWithMessage("accessStrategies field is not a slice")
					}
					for _, accessStrategy := range accessStrategiesSlice {
						accessStrategyMap, ok := accessStrategy.(map[string]interface{})
						if !ok {
							return nil, statusErrorWithMessage("accessStrategy is not a map")
						}
						if accessStrategyMap["handler"] == "no_auth" {
							noAuth = true
						}
					}
				}
				if noAuth {
					ruleMap["noAuth"] = true
				}
				delete(ruleMap, "accessStrategies")
			}
		default:
			return nil, statusErrorWithMessage("unexpected conversion version %q", toVersion)
		}
	case "gateway.kyma-project.io/v1beta2":
		switch toVersion {
		case "gateway.kyma-project.io/v1beta1":
			fmt.Printf("%s/%s: converting v1beta2->v1beta1\n", convertedObject.GetNamespace(), convertedObject.GetName())
			annotations := convertedObject.GetAnnotations()
			annotations["gateway.kyma-project.io/converted-to-v1beta1"] = "true"
			convertedObject.SetAnnotations(annotations)
			rules, _, err := unstructured.NestedFieldNoCopy(convertedObject.Object, "spec", "rules")
			if err != nil {
				return nil, statusErrorWithMessage("failed to get rules field")
			}
			rulesSlice, ok := rules.([]interface{})
			if !ok {
				return nil, statusErrorWithMessage("rules field is not a slice")
			}
			for _, rule := range rulesSlice {
				ruleMap, ok := rule.(map[string]interface{})
				if !ok {
					return nil, statusErrorWithMessage("rule field is not a map")
				}
				noAuth := ruleMap["noAuth"]
				if noAuth != nil {
					noAuthBool, ok := noAuth.(bool)
					if !ok {
						return nil, statusErrorWithMessage("noAuth field is not a boolean")
					}
					if noAuthBool {
						ruleMap["accessStrategies"] = []interface{}{
							map[string]interface{}{
								"handler": "no_auth",
							},
						}
					}
					delete(ruleMap, "noAuth")
				}
				accessStrategy := ruleMap["accessStrategy"]
				if accessStrategy != nil {
					accessStrategyMap, ok := accessStrategy.(map[string]interface{})
					if !ok {
						return nil, statusErrorWithMessage("accessStrategy field is not a map")
					}
					if accessStrategyMap["extAuth"] != nil || accessStrategyMap["jwt"] != nil {
						ruleMap["accessStrategies"] = []interface{}{
							map[string]interface{}{
								"handler": "allow",
							},
						}
					}
				}
			}
		default:
			return nil, statusErrorWithMessage("unexpected conversion version %q", toVersion)
		}
	default:
		return nil, statusErrorWithMessage("unexpected conversion version %q", fromVersion)
	}
	return convertedObject, statusSucceed()
}
