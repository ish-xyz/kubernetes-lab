# resource "aws_instance" "controllers" {
#   count             = 1
#   ami               = var.ami
#   instance_type     = var.controllers_instance_type
#   user_data         = file("templates/cloud-init-controller.yaml")
#   subnet_id         = var.subnet_id
#   key_name = var.key_name

#   tags = {
#     Name = "controller-${var.cluster_name}-${count.index}"
#     Role = "controller"
#     Cluster = "${var.cluster_name}"
#   }
# }

resource "local_file" "config_file" {
  content  = templatefile("${path.module}/templates/ca-nodes.conf.tftpl", {
    nodes = [for instance in local.mock_instances : {
      name = instance.tags["Name"]
    }]
  })
  filename = "${path.module}/files/ca-nodes.conf"
}

resource "null_resource" "generate_certs" {
  provisioner "local-exec" {
    command = "${path.module}/generate-certs.sh"
  }
}

data "template_file" "controllers_cloud_init_config" {
  template = file("${path.module}/templates/cloud-init-controller.tftpl")

  depends_on = [null_resource.generate_certs]
  
  vars = {
    certs_json = jsonencode([
      {
        name    = "admin.crt"
        content = base64encode(file("${path.module}/files/certs/admin.crt"))
      },
      {
        name    = "admin.csr"
        content = base64encode(file("${path.module}/files/certs/admin.csr"))
      },
      {
        name    = "admin.key"
        content = base64encode(file("${path.module}/files/certs/admin.key"))
      },
    ])
  }
    # admin_crt = ""
    # admin_key = ""
    # ca_crt = ""
    # ca_key = ""
    # kube_api_server_crt = ""
    # kube_api_server_csr = ""
    # kube_api_server_key = ""
    # kube_controller_manager_crt = ""
    # kube_controller_manager_csr = ""
    # kube_controller_manager_key = ""
    # kube_proxy_crt = ""
    # kube_proxy_csr = ""
    # kube_proxy_key = ""
    # kube_scheduler_crt = ""
    # kube_scheduler_csr = ""
    # kube_scheduler_key = ""
    # service_accounts_crt = ""
    # service_accounts_csr = ""
    # service_accounts_key = ""
}

output "rendered" {
  value = "${data.template_file.controllers_cloud_init_config.rendered}"
}


# resource "aws_instance" "nodes" {
#   count         = 3
#   ami           = var.ami
#   instance_type = var.nodes_instance_type

#   tags = {
#     Name = "node-${var.cluster_name}-${count.index}"
#     Role = "node"
#     Cluster = "${var.cluster_name}"
#   }
# }

