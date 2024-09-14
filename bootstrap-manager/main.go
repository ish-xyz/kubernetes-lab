package main

func main() {
	// load configuration file

	// check etcd instances health and quantity
	// get initialControlPlaneDriver

	/*
		initialDriver: systemd
		desiredDriver: kubernetes
		desiredEtcdCount: 3
		preSwitchDeployments:
		- chart: cilium
		  valuesB64: ""
		postSwitchDeployments:
		- chart: coredns
		- chart: cert-manager
		  valuesB64: ""
		- chart: argocd
		  valuesB64: ""
	*/
}
