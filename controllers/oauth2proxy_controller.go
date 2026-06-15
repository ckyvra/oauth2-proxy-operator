package controllers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	oauth2proxyv1alpha1 "github.com/ckyvra/oauth2-proxy-operator/api/v1alpha1"
)

const (
	oauth2ProxyFinalizer = "oauth2proxy.kyvrakidis.com/finalizer"
	defaultImage         = "quay.io/oauth2-proxy/oauth2-proxy:v7.8.1"
	defaultBackend       = ":4180"
	defaultReplicas      = int32(1)
)

type OAuth2ProxyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *OAuth2ProxyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	instance := &oauth2proxyv1alpha1.OAuth2Proxy{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(instance, oauth2ProxyFinalizer) {
			controllerutil.AddFinalizer(instance, oauth2ProxyFinalizer)
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(instance, oauth2ProxyFinalizer) {
			controllerutil.RemoveFinalizer(instance, oauth2ProxyFinalizer)
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	secret, err := r.getClientSecret(ctx, instance)
	if err != nil {
		logger.Error(err, "failed to get client secret")
		r.setStatus(ctx, instance, false, err.Error(), "Error")
		return ctrl.Result{}, err
	}

	dep, err := r.reconcileDeployment(ctx, instance, secret)
	if err != nil {
		logger.Error(err, "failed to reconcile deployment")
		r.setStatus(ctx, instance, false, err.Error(), "Error")
		return ctrl.Result{}, err
	}

	svc, err := r.reconcileService(ctx, instance)
	if err != nil {
		logger.Error(err, "failed to reconcile service")
		r.setStatus(ctx, instance, false, err.Error(), "Error")
		return ctrl.Result{}, err
	}

	ing, err := r.reconcileIngress(ctx, instance)
	if err != nil {
		logger.Error(err, "failed to reconcile ingress")
		r.setStatus(ctx, instance, false, err.Error(), "Error")
		return ctrl.Result{}, err
	}

	_ = dep
	_ = svc
	_ = ing

	r.setStatus(ctx, instance, true, "Deployment ready", "Running")
	return ctrl.Result{}, nil
}

func (r *OAuth2ProxyReconciler) getClientSecret(ctx context.Context, instance *oauth2proxyv1alpha1.OAuth2Proxy) (string, error) {
	if instance.Spec.ClientSecret == nil {
		return "", fmt.Errorf("clientSecret is required")
	}

	secret := &corev1.Secret{}
	ns := instance.Spec.ClientSecret.Namespace
	if ns == "" {
		ns = instance.Namespace
	}

	if err := r.Get(ctx, types.NamespacedName{
		Name:      instance.Spec.ClientSecret.Name,
		Namespace: ns,
	}, secret); err != nil {
		return "", fmt.Errorf("cannot read secret %s/%s: %w", ns, instance.Spec.ClientSecret.Name, err)
	}

	clientSecret, ok := secret.Data["client-secret"]
	if !ok {
		return "", fmt.Errorf("secret %s/%s does not contain key 'client-secret'", ns, instance.Spec.ClientSecret.Name)
	}

	return string(clientSecret), nil
}

func (r *OAuth2ProxyReconciler) reconcileDeployment(
	ctx context.Context,
	instance *oauth2proxyv1alpha1.OAuth2Proxy,
	secret string,
) (*appsv1.Deployment, error) {
	name := instance.Name
	ns := instance.Namespace
	labels := r.labels(instance)
	replicas := defaultReplicas
	if instance.Spec.Replicas != nil {
		replicas = *instance.Spec.Replicas
	}
	image := defaultImage
	if instance.Spec.Image != "" {
		image = instance.Spec.Image
	}
	address := defaultBackend
	if instance.Spec.Address != "" {
		address = instance.Spec.Address
	}

	args := []string{
		"--provider=oidc",
		fmt.Sprintf("--client-id=%s", instance.Spec.ClientID),
		fmt.Sprintf("--client-secret=%s", secret),
		fmt.Sprintf("--oidc-issuer-url=%s", instance.Spec.Issuer),
		fmt.Sprintf("--upstream=%s", instance.Spec.Upstream),
		fmt.Sprintf("--http-address=%s", address),
		"--email-domain=*",
		"--cookie-secure=false",
		"--skip-auth-regex=^/metrics",
	}

	if instance.Spec.Role != "" {
		args = append(args, fmt.Sprintf("--allowed-role=%s", instance.Spec.Role))
	}
	if instance.Spec.KeycloakRole != "" {
		args = append(args, fmt.Sprintf("--allowed-role=%s", instance.Spec.KeycloakRole))
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "oauth2-proxy",
							Image: image,
							Args:  args,
							Ports: []corev1.ContainerPort{
								{Name: "http", ContainerPort: 4180, Protocol: corev1.ProtocolTCP},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/ping",
										Port:   intstr.FromString("http"),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       10,
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/ping",
										Port:   intstr.FromString("http"),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       30,
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(instance, dep, r.Scheme); err != nil {
		return nil, err
	}

	logger := log.FromContext(ctx)

	existing := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, existing); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("creating deployment", "namespace", ns, "name", name)
			if err := r.Create(ctx, dep); err != nil {
				logger.Error(err, "failed to create deployment")
				return nil, err
			}
			logger.Info("deployment created")
			return dep, nil
		}
		return nil, err
	}

	logger.Info("updating existing deployment", "namespace", ns, "name", name)
	existing.Spec = dep.Spec
	if err := r.Update(ctx, existing); err != nil {
		logger.Error(err, "failed to update deployment")
		return nil, err
	}
	logger.Info("deployment updated")
	return dep, nil
}

func (r *OAuth2ProxyReconciler) reconcileService(
	ctx context.Context,
	instance *oauth2proxyv1alpha1.OAuth2Proxy,
) (*corev1.Service, error) {
	name := instance.Name
	ns := instance.Namespace
	labels := r.labels(instance)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       4180,
					TargetPort: intstr.FromInt(4180),
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(instance, svc, r.Scheme); err != nil {
		return nil, err
	}

	existing := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, existing); err != nil {
		if errors.IsNotFound(err) {
			return svc, r.Create(ctx, svc)
		}
		return nil, err
	}

	existing.Spec = svc.Spec
	return svc, r.Update(ctx, existing)
}

func (r *OAuth2ProxyReconciler) reconcileIngress(
	ctx context.Context,
	instance *oauth2proxyv1alpha1.OAuth2Proxy,
) (*networkingv1.Ingress, error) {
	ingressSpec := instance.Spec.Ingress
	if ingressSpec == nil || !ingressSpec.Enabled {
		return nil, r.cleanupIngress(ctx, instance)
	}

	name := instance.Name
	ns := instance.Namespace
	labels := r.labels(instance)

	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: strPtr(ingressSpec.IngressClassName),
			Rules: []networkingv1.IngressRule{
				{
					Host: ingressSpec.Host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: pathTypePtr(networkingv1.PathTypePrefix),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: name,
											Port: networkingv1.ServiceBackendPort{
												Number: 4180,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if ingressSpec.TLSSecretName != "" {
		ing.Spec.TLS = []networkingv1.IngressTLS{
			{
				Hosts:      []string{ingressSpec.Host},
				SecretName: ingressSpec.TLSSecretName,
			},
		}
	}

	if err := controllerutil.SetControllerReference(instance, ing, r.Scheme); err != nil {
		return nil, err
	}

	existing := &networkingv1.Ingress{}
	if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, existing); err != nil {
		if errors.IsNotFound(err) {
			return ing, r.Create(ctx, ing)
		}
		return nil, err
	}

	existing.Spec = ing.Spec
	existing.SetLabels(ing.Labels)
	return ing, r.Update(ctx, existing)
}

func (r *OAuth2ProxyReconciler) cleanupIngress(ctx context.Context, instance *oauth2proxyv1alpha1.OAuth2Proxy) error {
	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}
	if err := r.Delete(ctx, ing); client.IgnoreNotFound(err) != nil {
		return err
	}
	return nil
}

func (r *OAuth2ProxyReconciler) setStatus(
	ctx context.Context,
	instance *oauth2proxyv1alpha1.OAuth2Proxy,
	ready bool,
	reason, phase string,
) {
	if instance.Status.Ready == ready && instance.Status.Reason == reason && instance.Status.Phase == phase {
		return
	}

	patch := client.MergeFrom(instance.DeepCopy())
	instance.Status.Ready = ready
	instance.Status.Reason = reason
	instance.Status.Phase = phase
	if err := r.Status().Patch(ctx, instance, patch); err != nil {
		log.FromContext(ctx).Error(err, "failed to update status")
	}
}

func (r *OAuth2ProxyReconciler) labels(instance *oauth2proxyv1alpha1.OAuth2Proxy) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "oauth2-proxy",
		"app.kubernetes.io/instance":   instance.Name,
		"app.kubernetes.io/component":  "auth-proxy",
		"app.kubernetes.io/managed-by": "oauth2-proxy-operator",
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func pathTypePtr(pt networkingv1.PathType) *networkingv1.PathType {
	return &pt
}

func (r *OAuth2ProxyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&oauth2proxyv1alpha1.OAuth2Proxy{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&networkingv1.Ingress{}).
		Complete(r)
}
