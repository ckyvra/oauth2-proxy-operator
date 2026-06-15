package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IngressSpec struct {
	Enabled          bool   `json:"enabled"`
	Host             string `json:"host,omitempty"`
	IngressClassName string `json:"ingressClassName,omitempty"`
	TLSSecretName    string `json:"tlsSecretName,omitempty"`
}

type OAuth2ProxySpec struct {
	Upstream string `json:"upstream"`
	Address  string `json:"address,omitempty"`
	ClientID string `json:"clientId"`

	// +optional
	ClientSecret *corev1.SecretReference `json:"clientSecret,omitempty"`

	Realm        string `json:"realm"`
	Issuer       string `json:"issuer"`
	Role         string `json:"role,omitempty"`
	KeycloakRole string `json:"keycloakRole,omitempty"`

	// +optional
	// +kubebuilder:default=1
	Replicas *int32 `json:"replicas,omitempty"`

	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Ingress *IngressSpec `json:"ingress,omitempty"`
}

type OAuth2ProxyStatus struct {
	Ready   bool   `json:"ready"`
	Reason  string `json:"reason,omitempty"`
	Phase   string `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=oauth2proxies,scope=Namespaced,shortName=o2p

type OAuth2Proxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OAuth2ProxySpec   `json:"spec,omitempty"`
	Status OAuth2ProxyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type OAuth2ProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OAuth2Proxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OAuth2Proxy{}, &OAuth2ProxyList{})
}
