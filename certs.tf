## generate and import certificates into terraform
## for cloud init configuration

resource "local_file" "ca_config_file" {
  content  = templatefile("${path.module}/templates/certs-config/ca.conf.tftpl", {
    cluster_name = var.cluster_name
  })
  filename = "${path.module}/files/certs-config/ca.conf"
}

resource "local_file" "ca_ctrlplane_config_file" {
  content  = templatefile("${path.module}/templates/certs-config/ca-control-plane.conf.tftpl", {
    cluster_name = var.cluster_name
  })
  filename = "${path.module}/files/certs-config/ca-control-plane.conf"
}

resource "local_file" "ca_etcd_config_file" {
  content  = templatefile("${path.module}/templates/certs-config/ca-etcd.conf.tftpl", {
    etcd_nodes = local.etcd_nodes
    cluster_name = var.cluster_name
  })
  filename = "${path.module}/files/certs-config/ca-etcd.conf"
}


resource "local_file" "ca_nodes_config_file" {
  content  = templatefile("${path.module}/templates/certs-config/ca-nodes.conf.tftpl", {
    nodes = local.nodes
    cluster_name = var.cluster_name
  })
  filename = "${path.module}/files/certs-config/ca-nodes.conf"
}

#TODO: IF CA is regenerated, regen everything else too
resource "null_resource" "generate_ca" {
  triggers = {
    always_run = timestamp()
  }

  depends_on = [local_file.ca_config_file]

  provisioner "local-exec" {
    command = "${path.module}/scripts/generate-certs.sh ${path.module}/files/certs ${path.module}/files/certs-config/ca.conf ca"
  }
}

resource "null_resource" "generate_control_plane_certs" {
  triggers = {
    always_run = timestamp()
  }

  depends_on = [null_resource.generate_ca, local_file.ca_ctrlplane_config_file]

  provisioner "local-exec" {
    command = "${path.module}/scripts/generate-certs.sh ${path.module}/files/certs ${path.module}/files/certs-config/ca-control-plane.conf ${local.control_plane_components}"
  }
}

resource "null_resource" "generate_etcd_certs" {
  triggers = {
    always_run = timestamp()
  }

  depends_on = [null_resource.generate_ca, local_file.ca_etcd_config_file]

  provisioner "local-exec" {
    command = "${path.module}/scripts/generate-certs.sh ${path.module}/files/certs ${path.module}/files/certs-config/ca-etcd.conf ${local.etcd_certs_ids}"
  }
}

resource "null_resource" "generate_nodes_certs" {
  triggers = {
    always_run = timestamp()
  }

  depends_on = [null_resource.generate_ca, local_file.ca_nodes_config_file]

  provisioner "local-exec" {
    command = "${path.module}/scripts/generate-certs.sh ${path.module}/files/certs ${path.module}/files/certs-config/ca-nodes.conf ${local.nodes_string}"
  }
}

## Kube Admin Certs
data "local_file" "admin_crt" {
    filename = "${path.module}/files/certs/admin.crt"
    depends_on = [null_resource.generate_control_plane_certs]
}

data "local_file" "admin_csr" {
    filename = "${path.module}/files/certs/admin.csr"
    depends_on = [null_resource.generate_control_plane_certs]
}

data "local_file" "admin_key" {
    filename = "${path.module}/files/certs/admin.key"
    depends_on = [null_resource.generate_control_plane_certs]
}

## CA Certs
data "local_file" "ca_crt" {
    filename = "${path.module}/files/certs/ca.crt"
    depends_on = [null_resource.generate_control_plane_certs]
}

data "local_file" "ca_key" {
    filename = "${path.module}/files/certs/ca.key"
    depends_on = [null_resource.generate_control_plane_certs]
}

## Kube API Server Certs
data "local_file" "kube_api_server_crt" {
    filename = "${path.module}/files/certs/kube-api-server.crt"
    depends_on = [null_resource.generate_control_plane_certs]
}

data "local_file" "kube_api_server_csr" {
    filename = "${path.module}/files/certs/kube-api-server.csr"
    depends_on = [null_resource.generate_control_plane_certs]
}

data "local_file" "kube_api_server_key" {
    filename = "${path.module}/files/certs/kube-api-server.key"
    depends_on = [null_resource.generate_control_plane_certs]
}

## Kube Controller Manager Certs
data "local_file" "kube_controller_manager_crt" {
    filename = "${path.module}/files/certs/kube-controller-manager.crt"
    depends_on = [null_resource.generate_control_plane_certs]
}

data "local_file" "kube_controller_manager_csr" {
    filename = "${path.module}/files/certs/kube-controller-manager.csr"
    depends_on = [null_resource.generate_control_plane_certs]
}

data "local_file" "kube_controller_manager_key" {
    filename = "${path.module}/files/certs/kube-controller-manager.key"
    depends_on = [null_resource.generate_control_plane_certs]
}

## Kube Scheduler Certs
data "local_file" "kube_scheduler_crt" {
    filename = "${path.module}/files/certs/kube-scheduler.crt"
    depends_on = [null_resource.generate_control_plane_certs]
}

data "local_file" "kube_scheduler_csr" {
    filename = "${path.module}/files/certs/kube-scheduler.csr"
    depends_on = [null_resource.generate_control_plane_certs]
}

data "local_file" "kube_scheduler_key" {
    filename = "${path.module}/files/certs/kube-scheduler.key"
    depends_on = [null_resource.generate_control_plane_certs]
}

## Kube Scheduler Certs
data "local_file" "service_accounts_crt" {
    filename = "${path.module}/files/certs/service-accounts.crt"
    depends_on = [null_resource.generate_control_plane_certs]
}

data "local_file" "service_accounts_csr" {
    filename = "${path.module}/files/certs/service-accounts.csr"
    depends_on = [null_resource.generate_control_plane_certs]
}

data "local_file" "service_accounts_key" {
    filename = "${path.module}/files/certs/service-accounts.key"
    depends_on = [null_resource.generate_control_plane_certs]
}

## Kube Scheduler Certs
data "local_file" "kube_proxy_crt" {
    filename = "${path.module}/files/certs/kube-proxy.crt"
    depends_on = [null_resource.generate_control_plane_certs]
}

data "local_file" "kube_proxy_csr" {
    filename = "${path.module}/files/certs/kube-proxy.csr"
    depends_on = [null_resource.generate_control_plane_certs]
}

data "local_file" "kube_proxy_key" {
    filename = "${path.module}/files/certs/kube-proxy.key"
    depends_on = [null_resource.generate_control_plane_certs]
}

data "local_file" "etcd_peer_crt" {
    filename = "${path.module}/files/certs/etcd-peer.crt"
    depends_on = [null_resource.generate_etcd_certs]
}

data "local_file" "etcd_peer_key" {
    filename = "${path.module}/files/certs/etcd-peer.key"
    depends_on = [null_resource.generate_etcd_certs]
}

data "local_file" "etcd_client_crt" {
    filename = "${path.module}/files/certs/etcd-client.crt"
    depends_on = [null_resource.generate_etcd_certs]
}

data "local_file" "etcd_client_key" {
    filename = "${path.module}/files/certs/etcd-client.key"
    depends_on = [null_resource.generate_etcd_certs]
}
