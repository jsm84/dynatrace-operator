package capability

import (
	"context"
	"fmt"
	"testing"

	dynatracev1beta1 "github.com/Dynatrace/dynatrace-operator/api/v1beta1"
	"github.com/Dynatrace/dynatrace-operator/controllers/activegate/capability"
	"github.com/Dynatrace/dynatrace-operator/controllers/activegate/internal/consts"
	rsfs "github.com/Dynatrace/dynatrace-operator/controllers/activegate/reconciler/statefulset"
	"github.com/Dynatrace/dynatrace-operator/controllers/customproperties"
	"github.com/Dynatrace/dynatrace-operator/controllers/kubesystem"
	"github.com/Dynatrace/dynatrace-operator/logger"
	"github.com/Dynatrace/dynatrace-operator/scheme"
	"github.com/go-logr/logr"
	testlogr "github.com/go-logr/logr/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testValue     = "test-value"
	testUID       = "test-uid"
	testNamespace = "test-namespace"
)

var metricsCapability = capability.NewRoutingCapability(
	&dynatracev1beta1.DynaKube{
		Spec: dynatracev1beta1.DynaKubeSpec{
			Routing: dynatracev1beta1.RoutingSpec{
				Enabled: true,
			},
		},
	},
)

func TestNewReconiler(t *testing.T) {
	createDefaultReconciler(t)
}

func createDefaultReconciler(t *testing.T) *Reconciler {
	log := logger.NewDTLogger()
	clt := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: kubesystem.Namespace,
				UID:  testUID,
			},
		}).
		Build()
	instance := &dynatracev1beta1.DynaKube{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
		},
		Spec: dynatracev1beta1.DynaKubeSpec{
			APIURL: "https://testing.dev.dynatracelabs.com/api",
		},
	}
	r := NewReconciler(metricsCapability, clt, clt, scheme.Scheme, log, instance)
	require.NotNil(t, r)
	require.NotNil(t, r.Client)
	require.NotNil(t, r.Instance)

	return r
}

func TestReconcile(t *testing.T) {
	assertStatefulSetExists := func(r *Reconciler) *appsv1.StatefulSet {
		statefulSet := new(appsv1.StatefulSet)
		assert.NoError(t, r.Get(context.TODO(), client.ObjectKey{Name: r.calculateStatefulSetName(), Namespace: r.Instance.Namespace}, statefulSet))
		assert.NotNil(t, statefulSet)
		return statefulSet
	}
	assertServiceExists := func(r *Reconciler) *corev1.Service {
		svc := new(corev1.Service)
		assert.NoError(t, r.Get(context.TODO(), client.ObjectKey{Name: BuildServiceName(r.Instance.Name, r.ShortName()), Namespace: r.Instance.Namespace}, svc))
		assert.NotNil(t, svc)
		return svc
	}
	reconcileAndExpectUpdate := func(r *Reconciler, updateExpected bool) {
		update, err := r.Reconcile()
		assert.NoError(t, err)
		assert.Equal(t, updateExpected, update)
	}
	setStatsDFeatureFlags := func(r *Reconciler, enabled bool) {
		if r.Instance.Annotations == nil {
			r.Instance.Annotations = make(map[string]string)
		}
		r.Instance.Annotations["alpha.operator.dynatrace.com/feature-enable-statsd"] = fmt.Sprintf("%t", enabled)
		r.Instance.Annotations["alpha.operator.dynatrace.com/feature-use-activegate-image-for-statsd"] = fmt.Sprintf("%t", enabled)
	}

	agIngestServicePort := corev1.ServicePort{
		Name:       consts.ServicePortName,
		Protocol:   corev1.ProtocolTCP,
		Port:       consts.ServicePort,
		TargetPort: intstr.FromString(consts.ServiceTargetPort),
	}
	statsDIngestServicePort := corev1.ServicePort{
		Name:       consts.StatsDIngestPortName,
		Protocol:   corev1.ProtocolUDP,
		Port:       consts.StatsDIngestPort,
		TargetPort: intstr.FromString(consts.StatsDIngestTargetPort),
	}

	t.Run(`reconcile custom properties`, func(t *testing.T) {
		r := createDefaultReconciler(t)

		metricsCapability.Properties().CustomProperties = &dynatracev1beta1.DynaKubeValueSource{
			Value: testValue,
		}
		// Reconcile twice since service is created before the stateful set is
		reconcileAndExpectUpdate(r, true)
		reconcileAndExpectUpdate(r, true)

		var customProperties corev1.Secret
		err := r.Get(context.TODO(), client.ObjectKey{Name: r.Instance.Name + "-" + metricsCapability.ShortName() + "-" + customproperties.Suffix, Namespace: r.Instance.Namespace}, &customProperties)
		assert.NoError(t, err)
		assert.NotNil(t, customProperties)
		assert.Contains(t, customProperties.Data, customproperties.DataKey)
		assert.Equal(t, testValue, string(customProperties.Data[customproperties.DataKey]))
	})
	t.Run(`create stateful set`, func(t *testing.T) {
		r := createDefaultReconciler(t)
		// Reconcile twice since service is created before the stateful set is
		reconcileAndExpectUpdate(r, true)
		reconcileAndExpectUpdate(r, true)

		statefulSet := assertStatefulSetExists(r)
		assert.Contains(t, statefulSet.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  dtDNSEntryPoint,
			Value: buildDNSEntryPoint(r.Instance, r.ShortName()),
		})
	})
	t.Run(`update stateful set`, func(t *testing.T) {
		r := createDefaultReconciler(t)
		// Reconcile twice since service is created before the stateful set is
		reconcileAndExpectUpdate(r, true)
		reconcileAndExpectUpdate(r, true)
		{
			statefulSet := assertStatefulSetExists(r)
			assert.NotContains(t, statefulSet.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
				Name:  rsfs.DTInternalProxy,
				Value: testValue,
			})
		}

		r.Instance.Spec.Proxy = &dynatracev1beta1.DynaKubeProxy{Value: testValue}
		reconcileAndExpectUpdate(r, true)
		{
			statefulSet := assertStatefulSetExists(r)
			assert.Contains(t, statefulSet.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
				Name:  rsfs.DTInternalProxy,
				Value: testValue,
			})
		}
	})
	t.Run(`create service`, func(t *testing.T) {
		r := createDefaultReconciler(t)
		reconcileAndExpectUpdate(r, true)
		assertServiceExists(r)

		reconcileAndExpectUpdate(r, true)
		assertStatefulSetExists(r)
	})
	t.Run(`update service`, func(t *testing.T) {
		r := createDefaultReconciler(t)
		reconcileAndExpectUpdate(r, true)
		{
			service := assertServiceExists(r)
			assert.Len(t, service.Spec.Ports, 1)

			assert.Error(t, r.Get(context.TODO(), client.ObjectKey{Name: r.calculateStatefulSetName(), Namespace: r.Instance.Namespace}, &appsv1.StatefulSet{}))
		}

		reconcileAndExpectUpdate(r, true)
		{
			service := assertServiceExists(r)
			assert.Len(t, service.Spec.Ports, 1)
			assert.ElementsMatch(t, service.Spec.Ports, []corev1.ServicePort{
				agIngestServicePort,
			})

			statefulSet := assertStatefulSetExists(r)
			assert.Len(t, statefulSet.Spec.Template.Spec.Containers, 1)
		}
		reconcileAndExpectUpdate(r, false)

		setStatsDFeatureFlags(r, true)
		reconcileAndExpectUpdate(r, true)
		{
			service := assertServiceExists(r)
			assert.Len(t, service.Spec.Ports, 2)
			assert.ElementsMatch(t, service.Spec.Ports, []corev1.ServicePort{
				agIngestServicePort, statsDIngestServicePort,
			})

			statefulSet := assertStatefulSetExists(r)
			assert.Len(t, statefulSet.Spec.Template.Spec.Containers, 1)
		}

		reconcileAndExpectUpdate(r, true)
		{
			service := assertServiceExists(r)
			assert.ElementsMatch(t, service.Spec.Ports, []corev1.ServicePort{
				agIngestServicePort, statsDIngestServicePort,
			})

			statefulSet := assertStatefulSetExists(r)
			assert.Len(t, statefulSet.Spec.Template.Spec.Containers, 3)
		}
		reconcileAndExpectUpdate(r, false)
		reconcileAndExpectUpdate(r, false)

		setStatsDFeatureFlags(r, false)
		reconcileAndExpectUpdate(r, true)
		reconcileAndExpectUpdate(r, true)
		reconcileAndExpectUpdate(r, false)
		{
			service := assertServiceExists(r)
			assert.ElementsMatch(t, service.Spec.Ports, []corev1.ServicePort{
				agIngestServicePort,
			})

			statefulSet := assertStatefulSetExists(r)
			assert.Len(t, statefulSet.Spec.Template.Spec.Containers, 1)
		}
	})
}

func TestSetReadinessProbePort(t *testing.T) {
	r := createDefaultReconciler(t)
	stsProps := rsfs.NewStatefulSetProperties(r.Instance, metricsCapability.Properties(), "", "", "", "", "",
		nil, nil, nil, testlogr.TestLogger{T: t},
	)
	sts, err := rsfs.CreateStatefulSet(stsProps)

	assert.NoError(t, err)
	assert.NotNil(t, sts)

	setReadinessProbePort()(sts)

	assert.NotEmpty(t, sts.Spec.Template.Spec.Containers)
	assert.NotNil(t, sts.Spec.Template.Spec.Containers[0].ReadinessProbe)
	assert.NotNil(t, sts.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet)
	assert.NotNil(t, sts.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Port)
	assert.Equal(t, consts.ServiceTargetPort, sts.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Port.String())
}

func TestReconciler_calculateStatefulSetName(t *testing.T) {
	type fields struct {
		Reconciler *rsfs.Reconciler
		log        logr.Logger
		Capability *capability.RoutingCapability
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "instance and module names are defined",
			fields: fields{
				Reconciler: &rsfs.Reconciler{
					Instance: &dynatracev1beta1.DynaKube{
						ObjectMeta: metav1.ObjectMeta{
							Name: "instanceName",
						},
					},
				},
				Capability: metricsCapability,
			},
			want: "instanceName-routing",
		},
		{
			name: "empty instance name",
			fields: fields{
				Reconciler: &rsfs.Reconciler{
					Instance: &dynatracev1beta1.DynaKube{
						ObjectMeta: metav1.ObjectMeta{
							Name: "",
						},
					},
				},
				Capability: metricsCapability,
			},
			want: "-" + metricsCapability.ShortName(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reconciler{
				Reconciler: tt.fields.Reconciler,
				log:        tt.fields.log,
				Capability: tt.fields.Capability,
			}
			if got := r.calculateStatefulSetName(); got != tt.want {
				t.Errorf("Reconciler.calculateStatefulSetName() = %v, want %v", got, tt.want)
			}
		})
	}
}
