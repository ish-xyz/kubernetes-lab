resource "aws_instance" "controllers" {
  for_each          = toset(local.controllers_set)
  ami               = var.ami
  instance_type     = var.controllers_instance_type
  user_data_base64  = base64gzip(data.template_file.cloud_init_controllers[each.key].rendered)
  subnet_id         = var.subnet_id
  key_name          = var.key_name

  tags = {
    Name = each.value
    FQDN = "${each.value}.${var.domain}"
    Role = "controller"
    Cluster = "${var.cluster_name}"
  }
}

data "aws_route53_zone" "compute_zone" {
  zone_id      = var.route53_zone_id
}

resource "aws_route53_record" "www" {
  for_each  = aws_instance.controllers
  zone_id   = data.aws_route53_zone.compute_zone.zone_id
  name      = each.value.tags["Name"]
  type      = "A"
  ttl       = 300
  records   = [each.value.private_ip]
}

data "template_file" "etcd_systemd_unit" {
    template = file("${path.module}/templates/os-config/etcd.service.tftpl")
    vars = {
      etcd_cluster_token = "test"
      etcd_cluster_members = local.etcd_cluster_members
    }
}

data "template_file" "resolved_config" {
    template = file("${path.module}/templates/os-config/resolved.conf.tftpl")
    vars = {
      domain = var.domain
      aws_region = var.aws_region
      dns_list = join(" ", [for _, ns in data.aws_route53_zone.compute_zone.name_servers: ns])
    }
}

data "template_file" "cloud_init_controllers" {
  for_each  = toset(local.controllers_set)
  template  = file("${path.module}/templates/cloud-init/controllers.yaml.tftpl")
  
  vars = {
    fqdn = "${each.value}.${var.domain}"
    dns_config = data.template_file.resolved_config.rendered
    kube_certs = jsonencode([
      {
        name    = "admin.crt"
        content = base64encode(module.admin.cert)
      },
      {
        name    = "admin.key"
        content = base64encode(module.admin.key)
      },
      {
        name    = "ca.crt"
        content = base64encode(module.ca.ca_cert)
      },
      {
        name    = "ca.key"
        content = base64encode(module.ca.ca_key)
      },
      {
        name    = "kube-api-server.crt"
        content = base64encode(module.kube-api-server.cert)
      },
      {
        name    = "kube-api-server.key"
        content = base64encode(module.kube-api-server.key)
      },
      {
        name    = "service-accounts.crt"
        content = base64encode(module.service-accounts.cert)
      },
      {
        name    = "service-accounts.key"
        content = base64encode(module.service-accounts.key)
      }
    ])
    etcd_certs = jsonencode([
      {
        name    = "etcd-client.crt"
        content = base64encode(module.ca.ca_cert)
      },
      {
        name    = "etcd-client.key"
        content = base64encode(module.ca.ca_key)
      },
      {
        name    = "etcd-peer.crt"
        content = base64encode(module.ca.ca_cert)
      },
      {
        name    = "etcd-peer.key"
        content = base64encode(module.ca.ca_key)
      },
    ])
  }
}

# resource "local_file" "tmp_cloud_init" {
#   content  = data.template_file.cloud_init_controllers[local.controllers_set[0]].rendered
#   filename = "${path.module}/tmp-cloud-init.yaml"
# }

resource "local_file" "tmp_etcd_service" {
  content  = data.template_file.etcd_systemd_unit.rendered
  filename = "${path.module}/tmp-etcd.service"
}