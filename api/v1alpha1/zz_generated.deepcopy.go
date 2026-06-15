package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func (in *OAuth2Proxy) DeepCopyInto(out *OAuth2Proxy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

func (in *OAuth2Proxy) DeepCopy() *OAuth2Proxy {
	if in == nil {
		return nil
	}
	out := new(OAuth2Proxy)
	in.DeepCopyInto(out)
	return out
}

func (in *OAuth2Proxy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *OAuth2ProxyList) DeepCopyInto(out *OAuth2ProxyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OAuth2Proxy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *OAuth2ProxyList) DeepCopy() *OAuth2ProxyList {
	if in == nil {
		return nil
	}
	out := new(OAuth2ProxyList)
	in.DeepCopyInto(out)
	return out
}

func (in *OAuth2ProxyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *IngressSpec) DeepCopyInto(out *IngressSpec) {
	*out = *in
}

func (in *IngressSpec) DeepCopy() *IngressSpec {
	if in == nil {
		return nil
	}
	out := new(IngressSpec)
	in.DeepCopyInto(out)
	return out
}

func (in *OAuth2ProxySpec) DeepCopyInto(out *OAuth2ProxySpec) {
	*out = *in
	if in.ClientSecret != nil {
		in, out := &in.ClientSecret, &out.ClientSecret
		*out = new(corev1.SecretReference)
		**out = **in
	}
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.Ingress != nil {
		in, out := &in.Ingress, &out.Ingress
		*out = new(IngressSpec)
		**out = **in
	}
}

func (in *OAuth2ProxySpec) DeepCopy() *OAuth2ProxySpec {
	if in == nil {
		return nil
	}
	out := new(OAuth2ProxySpec)
	in.DeepCopyInto(out)
	return out
}

func (in *OAuth2ProxyStatus) DeepCopyInto(out *OAuth2ProxyStatus) {
	*out = *in
}

func (in *OAuth2ProxyStatus) DeepCopy() *OAuth2ProxyStatus {
	if in == nil {
		return nil
	}
	out := new(OAuth2ProxyStatus)
	in.DeepCopyInto(out)
	return out
}
