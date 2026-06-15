package controllers

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	oauth2proxyv1alpha1 "github.com/ckyvra/oauth2-proxy-operator/api/v1alpha1"
)

func createNamespace(name string) *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	Expect(k8sClient.Create(ctx, ns)).To(Succeed())
	return ns
}

func createSecret(name, ns, val string) *corev1.Secret {
	s := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		StringData: map[string]string{"client-secret": val},
	}
	Expect(k8sClient.Create(ctx, s)).To(Succeed())
	return s
}

func newOAuth2Proxy(ns string) *oauth2proxyv1alpha1.OAuth2Proxy {
	return &oauth2proxyv1alpha1.OAuth2Proxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-proxy",
			Namespace: ns,
		},
		Spec: oauth2proxyv1alpha1.OAuth2ProxySpec{
			Upstream: "http://my-app:8080",
			ClientID: "test-client",
			ClientSecret: &corev1.SecretReference{
				Name: "test-secret",
			},
			Realm:        "my-realm",
			Issuer:       "https://keycloak.example.com/realms/my-realm",
			Role:         "viewer",
			KeycloakRole: "admin",
			Replicas:     int32Ptr(1),
		},
	}
}

func int32Ptr(v int32) *int32 { return &v }

var _ = Describe("OAuth2Proxy controller", func() {
	var ns *corev1.Namespace

	BeforeEach(func() {
		ns = createNamespace("test-" + randStr())
		createSecret("test-secret", ns.Name, "supersecret")
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(ctx, ns)).To(Succeed())
	})

	It("should create a Deployment and Service", func() {
		cr := newOAuth2Proxy(ns.Name)
		Expect(k8sClient.Create(ctx, cr)).To(Succeed())

		dep := &appsv1.Deployment{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name: cr.Name, Namespace: ns.Name,
			}, dep)).To(Succeed())
		}, 5*time.Second).Should(Succeed())

		Expect(dep.Spec.Template.Spec.Containers).To(HaveLen(1))
		container := dep.Spec.Template.Spec.Containers[0]
		Expect(container.Name).To(Equal("oauth2-proxy"))
		Expect(container.Args).To(ContainElement("--provider=oidc"))
		Expect(container.Args).To(ContainElement("--client-id=test-client"))
		Expect(container.Args).To(ContainElement("--upstream=http://my-app:8080"))
		Expect(container.Args).To(ContainElement("--allowed-role=viewer"))
		Expect(container.Args).To(ContainElement("--allowed-role=admin"))

		svc := &corev1.Service{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name: cr.Name, Namespace: ns.Name,
			}, svc)).To(Succeed())
		}, 5*time.Second).Should(Succeed())

		Expect(svc.Spec.Ports).To(HaveLen(1))
		Expect(svc.Spec.Ports[0].Port).To(Equal(int32(4180)))
	})

	It("should create an Ingress when enabled", func() {
		cr := newOAuth2Proxy(ns.Name)
		cr.Spec.Ingress = &oauth2proxyv1alpha1.IngressSpec{
			Enabled:          true,
			Host:             "auth.example.com",
			IngressClassName: "traefik",
		}
		Expect(k8sClient.Create(ctx, cr)).To(Succeed())

		ing := &networkingv1.Ingress{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name: cr.Name, Namespace: ns.Name,
			}, ing)).To(Succeed())
		}, 5*time.Second).Should(Succeed())

		Expect(ing.Spec.Rules).To(HaveLen(1))
		Expect(ing.Spec.Rules[0].Host).To(Equal("auth.example.com"))
		Expect(*ing.Spec.IngressClassName).To(Equal("traefik"))
		Expect(ing.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name).To(Equal(cr.Name))
	})

	It("should not create Ingress when disabled", func() {
		cr := newOAuth2Proxy(ns.Name)
		cr.Spec.Ingress = &oauth2proxyv1alpha1.IngressSpec{
			Enabled: false,
		}
		Expect(k8sClient.Create(ctx, cr)).To(Succeed())

		Consistently(func(g Gomega) {
			ing := &networkingv1.Ingress{}
			err := k8sClient.Get(ctx, types.NamespacedName{
				Name: cr.Name, Namespace: ns.Name,
			}, ing)
			g.Expect(client.IgnoreNotFound(err)).To(Succeed())
		}, 2*time.Second).Should(Succeed())
	})

	It("should set owner references on created resources", func() {
		cr := newOAuth2Proxy(ns.Name)
		Expect(k8sClient.Create(ctx, cr)).To(Succeed())

		dep := &appsv1.Deployment{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name: cr.Name, Namespace: ns.Name,
			}, dep)).To(Succeed())
		}, 5*time.Second).Should(Succeed())

		Expect(dep.OwnerReferences).To(HaveLen(1))
		Expect(dep.OwnerReferences[0].Name).To(Equal(cr.Name))
		Expect(dep.OwnerReferences[0].Kind).To(Equal("OAuth2Proxy"))

		svc := &corev1.Service{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name: cr.Name, Namespace: ns.Name,
			}, svc)).To(Succeed())
		}, 5*time.Second).Should(Succeed())

		Expect(svc.OwnerReferences).To(HaveLen(1))
		Expect(svc.OwnerReferences[0].Name).To(Equal(cr.Name))
	})
})
