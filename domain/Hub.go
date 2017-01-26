package domain

// Hub is the primary root-aggregate within the domain that permits other connected
// domain components to communicate when connected to this root.
type Hub struct {
	API     *APIService
	Gateway *GatewayService
	Cluster *ClusterService
}
