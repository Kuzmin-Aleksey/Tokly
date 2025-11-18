package server

type Server struct {
	detector *DetectorServer
	groups   *GroupsServer
	metrics  *MetricServer
	images   *ImagesServer
}

func NewServer(
	detector *DetectorServer,
	groups *GroupsServer,
	metrics *MetricServer,
	images *ImagesServer,
) *Server {
	return &Server{
		detector: detector,
		groups:   groups,
		metrics:  metrics,
		images:   images,
	}
}
