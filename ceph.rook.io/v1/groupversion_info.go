// Package v1 holds copies of Rook's ceph.rook.io/v1 API types, with deepcopy functions and apply
// configurations generated for them.
// +kubebuilder:object:generate=true
// +kubebuilder:ac:generate=true
// +kubebuilder:ac:output:package=../../applyconfiguration
// +groupName=ceph.rook.io
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// SchemeGroupVersion is the group version of Rook's ceph.rook.io API.
	SchemeGroupVersion = schema.GroupVersion{Group: "ceph.rook.io", Version: "v1"}

	schemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme adds the copied types to the given scheme.
	AddToScheme = schemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion, &CephObjectStoreUser{}, &CephObjectStoreUserList{})
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
