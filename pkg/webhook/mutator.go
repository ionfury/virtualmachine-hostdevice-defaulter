package webhook

import (
	kubevirtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var mutatorlog = logf.Log.WithName("mutator")

func SetupWebhookWithManager(v *kubevirtv1.VirtualMachine, mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(v).
		Complete()
}
