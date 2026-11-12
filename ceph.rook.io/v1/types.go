/*
Copyright 2020 The Rook Authors. All rights reserved.

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

// Type declarations copied verbatim from Rook's pkg/apis/ceph.rook.io/v1, at the release in
// ROOK_VERSION. `make check-upstream` compares each one with that release.

package v1

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CephObjectStoreUser represents a Ceph Object Store Gateway User
// +kubebuilder:resource:shortName=rcou;objectuser;cephosu
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:subresource:status
type CephObjectStoreUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ObjectStoreUserSpec `json:"spec"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	Status *ObjectStoreUserStatus `json:"status,omitempty"`
}

// ObjectStoreUserStatus represents the status Ceph Object Store Gateway User
type ObjectStoreUserStatus struct {
	// +optional
	Phase string `json:"phase,omitempty"`
	// +optional
	// +nullable
	Info map[string]string `json:"info,omitempty"`
	// ObservedGeneration is the latest generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// +optional
	// +nullable
	Keys []SecretReference `json:"keys,omitempty"`
}

type SecretReference struct {
	v1.SecretReference `json:",secretReference"`
	UID                types.UID `json:"uid,omitempty"`
	ResourceVersion    string    `json:"resourceVersion,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CephObjectStoreUserList represents a list Ceph Object Store Gateway Users
type CephObjectStoreUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []CephObjectStoreUser `json:"items"`
}

// ObjectStoreUserSpec represents the spec of a CephObjectStoreUser
// +kubebuilder:validation:XValidation:message="defaultStorageClass requires defaultPlacement",rule="!has(self.defaultStorageClass) || has(self.defaultPlacement)"
type ObjectStoreUserSpec struct {
	// The store the user will be created in
	// +optional
	Store string `json:"store,omitempty"`
	// The display name for the ceph user.
	// +optional
	DisplayName string `json:"displayName,omitempty"`
	// +optional
	// +nullable
	Capabilities *ObjectUserCapSpec `json:"capabilities,omitempty"`
	// +optional
	// +nullable
	Quotas *ObjectUserQuotaSpec `json:"quotas,omitempty"`
	// Allows specifying credentials for the user. If not provided, the operator
	// will generate them.
	// +optional
	Keys []ObjectUserKey `json:"keys,omitempty"`
	// The namespace where the parent CephCluster and CephObjectStore are found
	// +optional
	ClusterNamespace string `json:"clusterNamespace,omitempty"`
	// The op-mask of the user.
	// +optional
	// +kubebuilder:validation:MinItems=0
	// +kubebuilder:validation:MaxItems=3
	// +listType=set
	OpMask *[]ObjectUserOpMask `json:"opMask,omitempty"`
	// AccountRef is a reference to a CephObjectStoreAccount to associate this user with.
	// The referenced account must be in the same namespace as the user.
	// When set, the user is created as an account user with no default permissions,
	// and resources created by this user are owned by the account.
	// +optional
	// +kubebuilder:validation:XValidation:message="accountRef is immutable",rule="self == oldSelf"
	AccountRef ObjectStoreUserAccountRef `json:"accountRef,omitzero"`
	// DefaultPlacement sets the default pool placement target for buckets
	// created by this user. It must name a placement target known to the
	// zonegroup serving the referenced object store; RGW rejects unknown
	// targets. If this field is absent the controller does not manage the
	// user's placement: an existing value (set previously through this
	// field, or outside of Rook) is left in place.
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=2048
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9._-]+$`
	DefaultPlacement string `json:"defaultPlacement,omitempty"`
	// DefaultStorageClass sets the default storage class for objects created
	// by this user, within the placement set by DefaultPlacement (which must
	// also be set). The storage class must exist on that placement target;
	// RGW rejects unknown storage classes. If this field is absent the
	// controller does not manage the user's storage class: an existing value
	// is preserved, except when DefaultPlacement changes, which resets the
	// storage class to the new target's default (STANDARD).
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=2048
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9._-]+$`
	DefaultStorageClass string `json:"defaultStorageClass,omitempty"`
}

// ObjectStoreUserAccountRef is a reference to a CephObjectStoreAccount
type ObjectStoreUserAccountRef struct {
	// Name of the CephObjectStoreAccount CR
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=2048
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9 ._-]+$`
	Name string `json:"name,omitempty"`
}

// Internally, RGW labels "operations" on persistent state as `RGW_OP_TYPE_READ` (`read`), `RGW_OP_TYPE_WRITE` (`write`), or `RGW_OP_TYPE_DELETE` (`delete`). All RGW users have an "operation mask", which does not function as a mask or filter as is typically implied by the word "mask", but as a set of allowed or permissible "operation" types the user is able to perform. The "operation mask" is applied regardless of the bucket or IAM policy. For example, in order for an RGW user to be able to read an object from a bucket, that user must have **both** the `read` "op mask" bit and an IAM/bucket policy that allows `s3:GetObject`. The default operations allowed are `read`, `write`, and `delete`. Setting the value to `[]` (an empty YAML sequence) causes all "operations" in the mask to be removed, meaning that the user will not be able to perform any operations. These operation masks are supported: `read`, `write`, `delete`
// +enum
// +kubebuilder:validation:Enum=read;write;delete
type ObjectUserOpMask string

// Additional admin-level capabilities for the Ceph object store user
type ObjectUserCapSpec struct {
	// +optional
	// +kubebuilder:validation:Enum={"*","read","write","read, write"}
	// Admin capabilities to read/write Ceph object store users. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
	User string `json:"user,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum={"*","read","write","read, write"}
	// Admin capabilities to read/write Ceph object store users. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
	Users string `json:"users,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum={"*","read","write","read, write"}
	// Admin capabilities to read/write Ceph object store buckets. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
	Bucket string `json:"bucket,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum={"*","read","write","read, write"}
	// Admin capabilities to read/write Ceph object store buckets. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
	Buckets string `json:"buckets,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum={"*","read","write","read, write"}
	// Admin capabilities to read/write Ceph object store metadata. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
	MetaData string `json:"metadata,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum={"*","read","write","read, write"}
	// Admin capabilities to read/write Ceph object store usage. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
	Usage string `json:"usage,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum={"*","read","write","read, write"}
	// Admin capabilities to read/write Ceph object store zones. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
	Zone string `json:"zone,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum={"*","read","write","read, write"}
	// Admin capabilities to read/write roles for user. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
	Roles string `json:"roles,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum={"*","read","write","read, write"}
	// Admin capabilities to read/write information about the user. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
	Info string `json:"info,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum={"*","read","write","read, write"}
	// Add capabilities for user to send request to RGW Cache API header. Documented in https://docs.ceph.com/en/latest/radosgw/rgw-cache/#cache-api
	AMZCache string `json:"amz-cache,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum={"*","read","write","read, write"}
	// Add capabilities for user to change bucket index logging. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
	BiLog string `json:"bilog,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum={"*","read","write","read, write"}
	// Add capabilities for user to change metadata logging. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
	MdLog string `json:"mdlog,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum={"*","read","write","read, write"}
	// Add capabilities for user to change data logging. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
	DataLog string `json:"datalog,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum={"*","read","write","read, write"}
	// Add capabilities for user to change user policies. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
	UserPolicy string `json:"user-policy,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum={"*","read","write","read, write"}
	// Add capabilities for user to change oidc provider. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
	OidcProvider string `json:"oidc-provider,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum={"*","read","write","read, write"}
	// Add capabilities for user to set rate limiter for user and bucket. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
	RateLimit string `json:"ratelimit,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum={"*","read","write","read, write"}
	// Add capabilities for user to fetch user info without keys. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
	// Note: Only supported from Ceph Squid (v19.2.0) onwards
	UserInfoWithoutKeys string `json:"userInfoWithoutKeys,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum={"*","read","write","read, write"}
	// Add capabilities for user to manage accounts. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
	// Note: Only supported from Ceph Squid (v19.2.3) onwards
	Accounts string `json:"accounts,omitempty"`
}

// ObjectUserQuotaSpec can be used to set quotas for the object store user to limit their usage. See the [Ceph docs](https://docs.ceph.com/en/latest/radosgw/admin/?#quota-management) for more
type ObjectUserQuotaSpec struct {
	// Maximum bucket limit for the ceph user
	// +optional
	// +nullable
	MaxBuckets *int `json:"maxBuckets,omitempty"`
	// Maximum size limit of all objects across all the user's buckets
	// See https://pkg.go.dev/k8s.io/apimachinery/pkg/api/resource#Quantity for more info.
	// +optional
	// +nullable
	MaxSize *resource.Quantity `json:"maxSize,omitempty"`
	// Maximum number of objects across all the user's buckets
	// +optional
	// +nullable
	MaxObjects *int64 `json:"maxObjects,omitempty"`
}

// ObjectUserKey defines a set of rgw user access credentials to be retrieved
// from secret resources.
type ObjectUserKey struct {
	// Secret key selector for the access_key (commonly referred to as AWS_ACCESS_KEY_ID).
	AccessKeyRef *v1.SecretKeySelector `json:"accessKeyRef,omitempty"`
	// Secret key selector for the secret_key (commonly referred to as AWS_SECRET_ACCESS_KEY).
	SecretKeyRef *v1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}
