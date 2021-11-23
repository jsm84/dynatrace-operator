package consts

const (
	ActiveGateContainerName = "dynatrace-operator" // TODO Does it make sense as the container name? Do we need a link to dynatracev1beta1.OperatorName?
	ServicePortName         = "https"
	ServicePort             = 443
	ServiceTargetPort       = "ag-https"

	EecContainerName = ActiveGateContainerName + "-eec"

	StatsDContainerName    = ActiveGateContainerName + "-statsd"
	StatsDIngestPortName   = "statsd"
	StatsDIngestPort       = 18125
	StatsDIngestTargetPort = "statsd-port"
)
