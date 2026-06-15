package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OAuth2ProxySpec defines the desired state of an OAuth2Proxy resource.
type OAuth2ProxySpec struct {
	// Upstream is the URL of the application to protect (e.g. http://app:8080).
	Upstream string `json:"upstream"`

	// Address is the oauth2-proxy listen address (default: :4180).
	Address string `json:"address,omitempty"`

	// ClientID is the Keycloak OIDC client ID.
	ClientID string `json:"clientId"`

	// ClientSecret references a Kubernetes Secret containing the 'client-secret' key.
	// +optional
	ClientSecret *corev1.SecretReference `json:"clientSecret,omitempty"`

	// Realm is the Keycloak realm name.
	Realm string `json:"realm"`

	// Issuer is the full Keycloak issuer URL
	// (e.g. https://keycloak.example.com/realms/myrealm).
	Issuer string `json:"issuer"`

	// Role is a Keycloak role required to access the application.
	Role string `json:"role,omitempty"`

	// KeycloakRole is an alternative or more specific Keycloak role.
	KeycloakRole string `json:"keycloakRole,omitempty"`

	// Replicas is the number of oauth2-proxy replicas (default: 1).
	// +optional
	// +kubebuilder:default=1
	Replicas *int32 `json:"replicas,omitempty"`

	// Image is the oauth2-proxy container image (default: quay.io/oauth2-proxy/oauth2-proxy:v7.8.1).
	// +optional
	Image string `json:"image,omitempty"`

	// Ingress configures an optional Ingress to expose oauth2-proxy externally.
	// +optional
	Ingress *IngressSpec `json:"ingress,omitempty"`
}

// IngressSpec configures a Kubernetes Ingress for the oauth2-proxy service.
type IngressSpec struct {
	// Enabled controls whether an Ingress is created.
	Enabled bool `json:"enabled"`

	// Host is the hostname for the Ingress rule.
	Host string `json:"host,omitempty"`

	// IngressClassName is the Ingress class name (e.g. traefik, nginx).
	IngressClassName string `json:"ingressClassName,omitempty"`

	// TLSSecretName references a Secret containing the TLS certificate.
	TLSSecretName string `json:"tlsSecretName,omitempty"`
}

// OAuth2ProxyStatus defines the observed state of an OAuth2Proxy resource.
type OAuth2ProxyStatus struct {
	// Ready indicates whether the deployment is ready.
	Ready bool `json:"ready"`

	// Reason provides details about the current status.
	Reason string `json:"reason,omitempty"`

	// Phase is the current lifecycle phase (Running, Error, etc.).
	Phase string `json:"phase,omitempty"`
}

// OAuth2Proxy is the Schema for the oauth2proxies API.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=oauth2proxies,scope=Namespaced,shortName=o2p
type OAuth2Proxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OAuth2ProxySpec   `json:"spec,omitempty"`
	Status OAuth2ProxyStatus `json:"status,omitempty"`
}

// OAuth2ProxyList contains a list of OAuth2Proxy.
// +kubebuilder:object:root=true
type OAuth2ProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OAuth2Proxy `json:"items"`
}
