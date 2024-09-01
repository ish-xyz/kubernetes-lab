locals {
  mock_instances = [
    {
      tags = {
        Name = "node-1"
      }
    },
    {
      tags = {
        Name = "node-2"
      }
    },
    {
      tags = {
        Name = "node-3"
      }
    }
  ]

  control_plane_components = "\"admin\" \"kube-proxy\" \"kube-scheduler\" \"kube-controller-manager\" \"kube-api-server\" \"service-accounts\""

  nodes = [
      for i in range(var.nodes_count) : {
          name = "node-${i}-${var.cluster_name}"
      }
  ]
  nodes_string = join(" ", [for node in local.nodes: node.name])


  controllers = [
    for i in range(var.nodes_count) : {
        name = "controller-${i}-${var.cluster_name}"
    }
  ]

  ### ETCD Config
  etcd_nodes = [
    for i in range(var.nodes_count) : {
        name = "controller-${i}-${var.cluster_name}"
    }
  ]
  etcd_nodes_alt_names = join(", ", [for i, node in local.etcd_nodes : "DNS.${i + 1}:${node.name}"])
  etcd_certs_ids = "etcd-client etcd-peer"
}