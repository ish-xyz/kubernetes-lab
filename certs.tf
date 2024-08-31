## generate and import certificates into terraform
## for cloud init configuration

resource "local_file" "config_file" {
  content  = templatefile("${path.module}/templates/certs-config/ca-nodes.conf.tftpl", {
    nodes = [
        for i in range(var.nodes_count) : {
            name = "node-${i}-${var.cluster_name}"
        }
    ]
  })
  filename = "${path.module}/files/certs-config/ca-nodes.conf"
}

resource "null_resource" "generate_certs" {
  triggers = {
    always_run = timestamp()
  }

  depends_on = [local_file.config_file]

  provisioner "local-exec" {
    command = "${path.module}/scripts/generate-certs.sh ${path.module}/files/certs ${path.module}/files/certs-config node-1 node-2 node-3"
  }
}

## Kube Admin Certs
data "local_file" "admin_crt" {
    filename = "${path.module}/files/certs/admin.crt"
    depends_on = [null_resource.generate_certs]
}

data "local_file" "admin_csr" {
    filename = "${path.module}/files/certs/admin.csr"
    depends_on = [null_resource.generate_certs]
}

data "local_file" "admin_key" {
    filename = "${path.module}/files/certs/admin.key"
    depends_on = [null_resource.generate_certs]
}

## CA Certs
data "local_file" "ca_crt" {
    filename = "${path.module}/files/certs/ca.crt"
    depends_on = [null_resource.generate_certs]
}

data "local_file" "ca_key" {
    filename = "${path.module}/files/certs/ca.key"
    depends_on = [null_resource.generate_certs]
}

## Kube API Server Certs
data "local_file" "kube_api_server_crt" {
    filename = "${path.module}/files/certs/kube-api-server.crt"
    depends_on = [null_resource.generate_certs]
}

data "local_file" "kube_api_server_csr" {
    filename = "${path.module}/files/certs/kube-api-server.csr"
    depends_on = [null_resource.generate_certs]
}

data "local_file" "kube_api_server_key" {
    filename = "${path.module}/files/certs/kube-api-server.key"
    depends_on = [null_resource.generate_certs]
}

## Kube Controller Manager Certs
data "local_file" "kube_controller_manager_crt" {
    filename = "${path.module}/files/certs/kube-controller-manager.crt"
    depends_on = [null_resource.generate_certs]
}

data "local_file" "kube_controller_manager_csr" {
    filename = "${path.module}/files/certs/kube-controller-manager.csr"
    depends_on = [null_resource.generate_certs]
}

data "local_file" "kube_controller_manager_key" {
    filename = "${path.module}/files/certs/kube-controller-manager.key"
    depends_on = [null_resource.generate_certs]
}

## Kube Scheduler Certs
data "local_file" "kube_scheduler_crt" {
    filename = "${path.module}/files/certs/kube-scheduler.crt"
    depends_on = [null_resource.generate_certs]
}

data "local_file" "kube_scheduler_csr" {
    filename = "${path.module}/files/certs/kube-scheduler.csr"
    depends_on = [null_resource.generate_certs]
}

data "local_file" "kube_scheduler_key" {
    filename = "${path.module}/files/certs/kube-scheduler.key"
    depends_on = [null_resource.generate_certs]
}

## Kube Scheduler Certs
data "local_file" "service_accounts_crt" {
    filename = "${path.module}/files/certs/service-accounts.crt"
    depends_on = [null_resource.generate_certs]
}

data "local_file" "service_accounts_csr" {
    filename = "${path.module}/files/certs/service-accounts.csr"
    depends_on = [null_resource.generate_certs]
}

data "local_file" "service_accounts_key" {
    filename = "${path.module}/files/certs/service-accounts.key"
    depends_on = [null_resource.generate_certs]
}

## Kube Scheduler Certs
data "local_file" "kube_proxy_crt" {
    filename = "${path.module}/files/certs/kube-proxy.crt"
    depends_on = [null_resource.generate_certs]
}

data "local_file" "kube_proxy_csr" {
    filename = "${path.module}/files/certs/kube-proxy.csr"
    depends_on = [null_resource.generate_certs]
}

data "local_file" "kube_proxy_key" {
    filename = "${path.module}/files/certs/kube-proxy.key"
    depends_on = [null_resource.generate_certs]
}
