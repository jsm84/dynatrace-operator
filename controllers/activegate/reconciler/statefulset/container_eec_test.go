package statefulset

import (
	"testing"

	logr "github.com/go-logr/logr/testing"
	"github.com/stretchr/testify/assert"
)

func TestExtensionController_BuildContainerAndVolumes(t *testing.T) {
	assertion := assert.New(t)

	instance := buildTestInstance()
	capabilityProperties := &instance.Spec.Routing.CapabilityProperties
	stsProperties := NewStatefulSetProperties(instance, capabilityProperties,
		"", "", "", "", "", "", "",
		nil, nil, nil, logr.TestLogger{T: t},
	)

	t.Run("happy path", func(t *testing.T) {
		eec := NewExtensionController(stsProperties)
		container := eec.BuildContainer()

		assertion.NotEmpty(container.ReadinessProbe, "Expected readiness probe is defined")
		assertion.Equal("/readyz", container.ReadinessProbe.Handler.HTTPGet.Path, "Expected there is a readiness probe at /readyz")
		assertion.Empty(container.LivenessProbe, "Expected there is no liveness probe (not implemented)")
		assertion.Empty(container.StartupProbe, "Expected there is no startup probe")

		for _, port := range []int32{eecIngestPort} {
			assertion.Truef(portIsIn(container.Ports, port), "Expected that EEC container defines port %d", port)
		}

		for _, mountPath := range []string{
			"/var/lib/dynatrace/gateway/config",
			"/mnt/dsexecargs",
			"/var/lib/dynatrace/remotepluginmodule/agent/runtime/datasources",
			"/var/lib/dynatrace/remotepluginmodule/agent/conf/runtime",
			"/opt/dynatrace/remotepluginmodule/agent/datasources/statsd",
		} {
			assertion.Truef(mountPathIsIn(container.VolumeMounts, mountPath), "Expected that EEC container defines mount point %s", mountPath)
		}

		for _, envVar := range []string{
			"TenantId", "ServerUrl", "EecIngestPort",
		} {
			assertion.Truef(envVarIsIn(container.Env, envVar), "Expected that EEC container defined environment variable %s", envVar)
		}
	})

	t.Run("volumes vs volume mounts", func(t *testing.T) {
		eec := NewExtensionController(stsProperties)
		statsd := NewStatsD(stsProperties)
		volumes := buildVolumes(stsProperties, []ContainerBuilder{eec, statsd})

		container := eec.BuildContainer()
		for _, volumeMount := range container.VolumeMounts {
			assertion.Truef(volumeIsDefined(volumes, volumeMount.Name), "Expected that volume mount %s has a predefined pod volume", volumeMount.Name)
		}
	})
}

func TestGetTenantId(t *testing.T) {
	testCases := []struct {
		apiUrl           string
		expectedTenantId string
		expectedError    string
	}{
		{
			apiUrl: "https://demo.dev.dynatracelabs.com/api", expectedTenantId: "demo",
			expectedError: "",
		},
		{
			apiUrl: "demo.dev.dynatracelabs.com/api", expectedTenantId: "",
			expectedError: "problem getting tenant id from fqdn ''",
		},
		{
			apiUrl: "https://google.com", expectedTenantId: "",
			expectedError: "api url https://google.com does not end with /api",
		},
		{
			apiUrl: "/api", expectedTenantId: "",
			expectedError: "problem getting tenant id from fqdn ''",
		},
	}

	for _, testCase := range testCases {
		actualTenantId, err := getTenantId(testCase.apiUrl)
		if len(testCase.expectedError) > 0 {
			assert.EqualErrorf(t, err, testCase.expectedError, "Expected that getting tenant id from '%s' will result in: '%v'",
				testCase.apiUrl, testCase.expectedError,
			)
		} else {
			assert.NoErrorf(t, err, "Expected that getting tenant id from '%s' will be successful", testCase.apiUrl)
		}
		assert.Equalf(t, testCase.expectedTenantId, actualTenantId, "Expected that tenant id of %s is %s, but found %s",
			testCase.apiUrl, testCase.expectedTenantId, actualTenantId,
		)
	}
}
