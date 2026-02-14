package compose

// ComposeFile represents a docker-compose.yml file
type ComposeFile struct {
	Version  string                       `yaml:"version"`
	Services map[string]*ComposeService   `yaml:"services"`
	Networks map[string]*ComposeNetwork   `yaml:"networks"`
	Volumes  map[string]*ComposeVolume    `yaml:"volumes,omitempty"`
	Secrets  map[string]*ComposeSecret    `yaml:"secrets,omitempty"`
}

// ComposeService represents a service in docker-compose
type ComposeService struct {
	Image         string                 `yaml:"image,omitempty"`
	Build         *ComposeBuild          `yaml:"build,omitempty"`
	ContainerName string                 `yaml:"container_name,omitempty"`
	Restart       string                 `yaml:"restart,omitempty"`
	Ports         []string               `yaml:"ports,omitempty"`
	Environment   map[string]string      `yaml:"environment,omitempty"`
	Volumes       []string               `yaml:"volumes,omitempty"`
	Networks      []string               `yaml:"networks,omitempty"`
	DependsOn     []string               `yaml:"depends_on,omitempty"`
	Secrets       []string               `yaml:"secrets,omitempty"`
	CapAdd        []string               `yaml:"cap_add,omitempty"`
	Devices       []string               `yaml:"devices,omitempty"`
	ExtraHosts    []string               `yaml:"extra_hosts,omitempty"`
	HealthCheck   *ComposeHealthCheck    `yaml:"healthcheck,omitempty"`
	Deploy        *ComposeDeploy         `yaml:"deploy,omitempty"`
	Profiles      []string               `yaml:"profiles,omitempty"`
}

// ComposeBuild represents build configuration
type ComposeBuild struct {
	Context    string `yaml:"context"`
	Dockerfile string `yaml:"dockerfile"`
}

// ComposeHealthCheck represents health check configuration
type ComposeHealthCheck struct {
	Test     []string `yaml:"test"`
	Interval string   `yaml:"interval"`
	Timeout  string   `yaml:"timeout"`
	Retries  int      `yaml:"retries"`
}

// ComposeDeploy represents deployment configuration
type ComposeDeploy struct {
	Resources *ComposeResources `yaml:"resources,omitempty"`
}

// ComposeResources represents resource limits
type ComposeResources struct {
	Limits       *ComposeResourceLimits `yaml:"limits,omitempty"`
	Reservations *ComposeResourceLimits `yaml:"reservations,omitempty"`
}

// ComposeResourceLimits represents resource limit values
type ComposeResourceLimits struct {
	Memory string `yaml:"memory,omitempty"`
	CPUs   string `yaml:"cpus,omitempty"`
}

// ComposeNetwork represents a network in docker-compose
type ComposeNetwork struct {
	Driver string `yaml:"driver,omitempty"`
}

// ComposeVolume represents a volume in docker-compose
type ComposeVolume struct {
	Driver string `yaml:"driver,omitempty"`
}

// ComposeSecret represents a secret in docker-compose
type ComposeSecret struct {
	File string `yaml:"file"`
}
