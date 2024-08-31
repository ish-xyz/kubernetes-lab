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

data "template_file" "controllers_cloud_init_config" {
  template = file("${path.module}/templates/cloud-init-controller.yaml.tftpl")

  depends_on = [null_resource.generate_certs]
  
  vars = {
    certs_json = jsonencode([
      {
        name    = "admin.crt"
        content = base64encode(data.local_file.admin_crt.content)
      },
      {
        name    = "admin.csr"
        content = base64encode(data.local_file.admin_csr.content)
      },
      {
        name    = "admin.key"
        content = base64encode(data.local_file.admin_key.content)
      },
      {
        name    = "ca.crt"
        content = base64encode(data.local_file.ca_crt.content)
      },
      {
        name    = "ca.key"
        content = base64encode(data.local_file.ca_key.content)
      },
    ])
  }
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

