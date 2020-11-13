package perfserver

import "github.com/epmd-edp/perf-operator/v2/pkg/apis/edp/v1alpha1"

type PerfServer struct {
	Name      string
	Available bool
}

func ConvertPerfServerToDto(server v1alpha1.PerfServer) PerfServer {
	return PerfServer{
		Name:      server.Name,
		Available: server.Status.Available,
	}
}
