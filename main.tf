resource "aws_instance" "controllers" {
  count             = var.controllers_count
  ami               = var.ami
  instance_type     = var.controllers_instance_type
  user_data         = data.template_file.cloud_init_controllers[count.index].rendered
  subnet_id         = var.subnet_id
  key_name = var.key_name

  tags = {
    Name = "controller-${count.index}-${var.cluster_name}"
    Role = "controller"
    Cluster = "${var.cluster_name}"
  }
}

data "template_file" "cloud_init_controllers" {
  count = var.controllers_count
  template = file("${path.module}/templates/cloud-init/controllers.yaml.tftpl")

  depends_on = [null_resource.generate_certs]
  
  vars = {
    instance_name = "controller-${count.index}-${var.cluster_name}"
    kube_certs = jsonencode([
      {
        name    = "admin.crt"
        content = base64encode(data.local_file.admin_crt.content)
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
      {
        name    = "kube-api-server.crt"
        content = base64encode(data.local_file.kube_api_server_crt.content)
      },
      {
        name    = "kube-api-server.key"
        content = base64encode(data.local_file.kube_api_server_key.content)
      },
      {
        name    = "service-accounts.crt"
        content = base64encode(data.local_file.service_accounts_crt.content)
      },
      {
        name    = "service-accounts.key"
        content = base64encode(data.local_file.service_accounts_key.content)
      }
    ])

    etcd_certs = jsonencode([
      {
        name    = "ca.crt"
        content = base64encode(data.local_file.ca_crt.content)
      },
      {
        name    = "ca.key"
        content = base64encode(data.local_file.ca_key.content)
      }
    ])
    etcd_systemd_unit = ""
  }
}
