package controllers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	oauth2proxyv1alpha1 "github.com/ckyvra/oauth2-proxy-operator/api/v1alpha1"
)

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	var err error

	crdPath := findCRDYAML()

	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{crdPath},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())

	Expect(oauth2proxyv1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(appsv1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(corev1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(networkingv1.AddToScheme(scheme.Scheme)).To(Succeed())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).NotTo(HaveOccurred())

	Expect((&OAuth2ProxyReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)).To(Succeed())

	ctx, cancel = context.WithCancel(context.Background())
	go func() {
		defer GinkgoRecover()
		Expect(mgr.Start(ctx)).To(Succeed())
	}()
})

var _ = AfterSuite(func() {
	cancel()
	Expect(testEnv.Stop()).To(Succeed())
})

func findCRDYAML() string {
	candidates := []string{
		"../config/crd/oauth2proxy.kyvrakidis.com_oauth2proxies.yaml",
		"config/crd/oauth2proxy.kyvrakidis.com_oauth2proxies.yaml",
	}
	for _, p := range candidates {
		abs, err := filepath.Abs(p)
		if err == nil {
			if _, err := os.Stat(abs); err == nil {
				return filepath.Dir(abs)
			}
		}
	}
	panic(fmt.Sprintf("CRD YAML not found (cwd: %s)", mustAbs(".")))
}

func mustAbs(p string) string {
	a, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return a
}

var testCounter atomic.Int64

func randStr() string {
	n := testCounter.Add(1)
	return fmt.Sprintf("%d-%d", GinkgoRandomSeed(), n)
}
