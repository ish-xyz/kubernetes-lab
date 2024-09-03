resource "aws_instance" "controllers" {
  for_each                    = toset(local.controllers_set)
  ami                         = var.ami
  instance_type               = var.controllers_instance_type
  subnet_id                   = var.subnet_id
  key_name                    = var.key_name
  iam_instance_profile        = aws_iam_instance_profile.kube_controllers_profile.name
  user_data_replace_on_change = true
  user_data = <<-EOF
              #!/bin/bash
              set -euo pipefail
              snap install aws-cli --classic
              cloud_config_url=$(aws s3 presign s3://${aws_s3_bucket.config_bucket.bucket}/${var.cluster_name}/${each.value}-config.yaml --expires-in 3600)
              curl -L -o /etc/cloud/cloud.cfg.d/custom.cfg $cloud_config_url
              [[ ! -f /custom-cloud-init-done ]] && cloud-init clean --logs --reboot
              EOF
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
    for_each = toset(local.controllers_set)
    template = file("${path.module}/templates/os-config/etcd.service.tftpl")
    vars = {
      etcd_cluster_token = "test" # todo change token
      etcd_name = each.value
      etcd_cluster_members = local.etcd_cluster_members
    }
}

data "dns_a_record_set" "name_servers" {
  for_each = toset(data.aws_route53_zone.compute_zone.name_servers)
  host = each.value
}

data "template_file" "resolved_config" {
    template = file("${path.module}/templates/os-config/resolved.conf.tftpl")
    vars = {
      domain = var.domain
      aws_region = var.aws_region
      nameservers_list = join(" ", [for _, ns in data.dns_a_record_set.name_servers: join(" ", [for _, ip in ns.addrs: ip])])
    }
}

data "template_file" "cloud_init_controllers" {
  for_each  = toset(local.controllers_set)
  template  = file("${path.module}/templates/cloud-init/controllers.yaml.tftpl")
  
  vars = {
    fqdn = "${each.value}.${var.domain}"
    domain = var.domain
    dns_config = base64encode(data.template_file.resolved_config.rendered)
    etcd_systemd_unit = base64encode(data.template_file.etcd_systemd_unit[each.key].rendered)
    etcd_full_version = var.etcd_full_version
    etcd_version = var.etcd_version
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
        name    = "ca.crt"
        content = base64encode(module.ca.ca_cert)
      },
      {
        name    = "ca.key"
        content = base64encode(module.ca.ca_key)
      },
      {
        name    = "etcd-client.crt"
        content = base64encode(module.etcd-client.cert)
      },
      {
        name    = "etcd-client.key"
        content = base64encode(module.etcd-client.key)
      },
      {
        name    = "etcd-peer.crt"
        content = base64encode(module.etcd-peer.cert)
      },
      {
        name    = "etcd-peer.key"
        content = base64encode(module.etcd-peer.key)
      },
    ])
  }
}

# Create S3 bucket
resource "aws_s3_bucket" "config_bucket" {
  bucket = "cloud-init-configurations"
  force_destroy = true
}

# Upload cloud-init config to S3
resource "aws_s3_object" "controllers_cloud_init_config" {
  for_each = toset(local.controllers_set)
  bucket = aws_s3_bucket.config_bucket.id
  key    = "${var.cluster_name}/${each.value}-config.yaml"
  content = data.template_file.cloud_init_controllers[each.key].rendered
}

resource "local_file" "tmp_cloud_init" {
  content  = data.template_file.cloud_init_controllers[local.controllers_set[0]].rendered
  filename = "${path.module}/tmp-cloud-init.yaml"
}

resource "local_file" "tmp_etcd_service" {
  content  = data.template_file.etcd_systemd_unit[local.controllers_set[0]].rendered
  filename = "${path.module}/tmp-etcd.service"
}