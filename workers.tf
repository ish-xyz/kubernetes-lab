# Workers compute resources
resource "aws_instance" "workers" {
  for_each                    = toset(local.workers_set)
  ami                         = var.ami
  instance_type               = var.workers_instance_type
  subnet_id                   = var.subnet_id
  key_name                    = var.key_name
  iam_instance_profile        = aws_iam_instance_profile.kube_nodes.name
  user_data_replace_on_change = true
  user_data = <<-EOF
              #!/bin/bash
              # ${md5(data.template_file.workers_cloud_init[each.value].rendered)}
              set -euo pipefail
              snap install aws-cli --classic
              cloud_config_url=$(aws s3 presign s3://${aws_s3_bucket.config_bucket.bucket}/${var.cluster_name}/${each.value}-config.yaml --expires-in 3600)
              curl -L -o /etc/cloud/cloud.cfg.d/custom.cfg $cloud_config_url
              [[ ! -f /custom-cloud-init-done ]] && cloud-init clean --logs --reboot
              EOF
  tags = {
    Name = each.value
    FQDN = "${each.value}.${var.domain}"
    Role = "worker"
    Cluster = "${var.cluster_name}"
  }
}

# DNS Records for workers
resource "aws_route53_record" "workers" {
  for_each  = aws_instance.workers
  zone_id   = data.aws_route53_zone.compute_zone.zone_id
  name      = each.value.tags["Name"]
  type      = "A"
  ttl       = 300
  records   = [each.value.private_ip]
}

resource "aws_s3_object" "workers_cloud_init" {
  for_each = toset(local.workers_set)
  bucket = aws_s3_bucket.config_bucket.id
  key    = "${var.cluster_name}/${each.value}-config.yaml"
  content = data.template_file.workers_cloud_init[each.key].rendered
}

# Templates

## Systemd units templates

data "template_file" "workers_kubelet_systemd_unit" {
  template = file("${path.module}/templates/workers/service-kubelet.tftpl")
  vars = {
    kube_config_dir = local.kube_config_dir
  }
}

data "template_file" "workers_containerd_systemd_unit" {
  template = file("${path.module}/templates/workers/service-containerd.tftpl")
  vars = {}
}


## Components configurations

data "template_file" "workers_kubelet_config" {
  template = file("${path.module}/templates/workers/kubelet-config.tftpl")
  vars = {
    pod_cidr = var.pod_cidr
    cluster_domain = var.cluster_domain
    kube_certs_dir = local.kube_certs_dir
    cluster_dns_servers = jsonencode(var.cluster_dns_servers)
  }
}

## Kubeconfig files

data "template_file" "workers_kubeconfig_kubelet" {
    for_each = toset(local.workers_set)
    template = file("${path.module}/templates/workers/kubeconfig-kubelet.tftpl")
    vars = {
      ca_crt = base64encode(module.ca.ca_cert)
      node_name = each.value
      cluster_name = var.cluster_name
      kubelet_crt = base64encode(module.workers-kubelet[each.key].cert)
      kubelet_key = base64encode(module.workers-kubelet[each.key].key)
    }
}

## Workers Cloud Init Config
data "template_file" "workers_cloud_init" {
  for_each  = toset(local.workers_set)
  template  = file("${path.module}/templates/workers/cloud-init.yaml.tftpl")
  
  vars = {
    fqdn = "${each.value}.${var.domain}"
    domain = var.domain
    kube_version = var.kube_version
    runc_version = var.runc_version
    containerd_version = var.containerd_version
    kube_config_dir = local.kube_config_dir
    kube_certs_dir = local.kube_certs_dir
    arch = var.architecture

    # install packages
    packages = jsonencode([
      "socat", 
      "conntrack", 
      "ipset", 
      "net-tools",
    ])

    # write files
    dns_config = base64encode(data.template_file.resolved_config.rendered)
    containerd_config = base64encode(file("${path.module}/files/containerd.toml"))
    systemd_units = jsonencode([
      {
        name = "containerd"
        content = base64encode(data.template_file.workers_containerd_systemd_unit.rendered)
      },
      {
        name = "kubelet"
        content = base64encode(data.template_file.workers_kubelet_systemd_unit.rendered)
      },
    ])
    kube_configs = jsonencode([
      {
        name    = "kubelet.kubeconfig"
        content = base64encode(data.template_file.workers_kubeconfig_kubelet[each.key].rendered)
      },
      {
        name    = "kubelet-config.yaml"
        content = base64encode(data.template_file.workers_kubelet_config.rendered)
      }
    ])
    kube_certs = jsonencode([
      {
        name    = "ca.crt"
        content = base64encode(module.ca.ca_cert)
      },
      {
        name    = "ca.key"
        content = base64encode(module.ca.ca_key)
      },
      {
        name    = "kubelet.crt"
        content = base64encode(module.kubelet-workers[each.key].cert)
      },
      {
        name    = "kubelet.key"
        content = base64encode(module.kubelet-workers[each.key].key)
      }
    ])
  }
}
