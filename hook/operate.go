package hook

import (
	"encoding/json"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)


// 1. modify securityContext
// path:/spec/securityContext
func createSpecSecurityContextPatch(availableAnnotations map[string]string, annotations map[string]string, availableLabels map[string]string, labels map[string]string) ([]byte, error) {
	var patch []patchOperation
	// update Annotation to set admissionWebhookAnnotationStatusKey: "mutated"
	patch = append(patch, updateAnnotation(availableAnnotations, annotations)...)
	// add labels
	patch = append(patch, updateLabels(availableLabels, labels)...)

	// update pod spec securityContext
	replaceSecurityContext := patchOperation{
		Op:    "replace",
		Path:  "/spec/securityContext",
		Value: &corev1.PodSecurityContext{
			RunAsUser: &securityContextValue,
			RunAsGroup: &securityContextValue,
			FSGroup: &securityContextValue,
		},
	}
	glog.Infof("modify  pod spec securityContext for value: %v", replaceSecurityContext)
	patch = append(patch, replaceSecurityContext)

	return json.Marshal(patch)
}


// 2. modify pod containers resources
// path:/spec/containers/0/image/resources
func createModifyContainersResourcesPatch(pod corev1.Pod, availableAnnotations map[string]string, annotations map[string]string, availableLabels map[string]string, labels map[string]string) ([]byte, error) {
	var patch []patchOperation

	patch = append(patch, updateAnnotation(availableAnnotations, annotations)...)
	patch = append(patch, updateLabels(availableLabels, labels)...)

	// update pod container resource
	replaceResourse := patchOperation{
		Op:    "replace",
		Path:  "/spec/containers/0/resources",
		Value: assembleResourceRequirements(pod),
	}
	glog.Infof("update  pod container resource for value: %v", replaceResourse)
	patch = append(patch, replaceResourse)

	return json.Marshal(patch)
}

// 3. remove pod all containers resources
// path:/spec/containers/*
func createRemoveContainersResourcesPatch(pod corev1.Pod, availableAnnotations map[string]string, annotations map[string]string, availableLabels map[string]string, labels map[string]string) ([]byte, error) {
	var patch []patchOperation
	patch = append(patch, updateAnnotation(availableAnnotations, annotations)...)
	patch = append(patch, updateLabels(availableLabels, labels)...)

	// pod containers size
	var size = len(pod.Spec.Containers)
	for i := 0; i < size; i++ {
		removeResourse := patchOperation{
			Op:    "replace",
			Path:  "/spec/containers/" + string(i) + "/resources",
			Value: nil,
		}
		glog.Infof("remove  pod container resource for value: %v", removeResourse)
		patch = append(patch, removeResourse)
	}

	return json.Marshal(patch)
}


func updateAnnotation(target map[string]string, added map[string]string) (patch []patchOperation) {
	for key, value := range added {
		if target == nil || target[key] == "" {
			target = map[string]string{}
			patch = append(patch, patchOperation{
				Op:   "add",
				Path: "/metadata/annotations",
				Value: map[string]string{
					key: value,
				},
			})
		} else {
			patch = append(patch, patchOperation{
				Op:    "replace",
				Path:  "/metadata/annotations/" + key,
				Value: value,
			})
		}
	}
	return patch
}

func updateLabels(target map[string]string, added map[string]string) (patch []patchOperation) {
	values := make(map[string]string)
	for key, value := range added {
		if target == nil || target[key] == "" {
			values[key] = value
		}
	}
	patch = append(patch, patchOperation{
		Op:    "add",
		Path:  "/metadata/labels",
		Value: values,
	})
	return patch
}


func assembleResourceRequirements(pod corev1.Pod) corev1.ResourceRequirements {
	res := corev1.ResourceRequirements{}
	limitResource := corev1.ResourceList{}
	requestResource := corev1.ResourceList{}
	limitResource[corev1.ResourceCPU] = resource.MustParse("200m")
	limitResource[corev1.ResourceMemory] = resource.MustParse("150Mi")

	requestResource[corev1.ResourceCPU] = resource.MustParse("200m")

	requestResource[corev1.ResourceMemory] = resource.MustParse("150Mi")

	res.Requests = requestResource
	res.Limits = limitResource
	// jiexun default limits have value, put the limits value to the requests value
	//res.Requests = pod.Spec.Containers[0].Resources.Limits
	//res.Limits = pod.Spec.Containers[0].Resources.Limits
	return res
}
