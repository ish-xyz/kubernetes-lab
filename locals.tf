locals {

  controllers_set = [for i in range(var.controllers_count) : "controller-${i}-${var.cluster_name}"]

  ### ETCD Config
  etcd_nodes = [for i in range(var.nodes_count) : "controller-${i}-${var.cluster_name}"]
  etcd_nodes_fqdns = [for _, node in local.etcd_nodes : "${node}.${var.domain}"]
  etcd_cluster_members = join(",", [for _, node in local.etcd_nodes : "${node}=https://${node}:2380"])
  etcd_certs_ids = "etcd-client etcd-peer"

  kube_config_dir = "/etc/kubernetes"
  kube_certs_dir = "${local.kube_config_dir}/ssl"
}