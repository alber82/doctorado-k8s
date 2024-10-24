package common

type Event string

const (
	Normal  Event = "Normal"
	Warning Event = "Warning"
)

type Reason string

const (
	Created          Reason = "Created"
	Updated          Reason = "Updated"
	Ready            Reason = "Ready"
	Unhealthy        Reason = "Unhealthy"
	Config           Reason = "Config"
	Misconfiguration Reason = "Misconfiguration"
	Unknown          Reason = "Unknown"
)
