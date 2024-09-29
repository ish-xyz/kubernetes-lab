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
              # ${md5(local.workers_cloud_init[each.value])}
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
  content = local.workers_cloud_init[each.key]
}

# Templates

## Systemd units templates

locals {
  workers_kubelet_systemd_unit = templatefile("${path.module}/templates/workers/service-kubelet.tftpl", {
    kube_config_dir = local.kube_config_dir
  })

  workers_containerd_systemd_unit = templatefile("${path.module}/templates/workers/service-containerd.tftpl", {})

  workers_kubelet_config = templatefile("${path.module}/templates/workers/kubelet-config.tftpl", {
    pod_cidr           = var.pod_cidr
#    cluster_domain     = var.cluster_domain
    kube_certs_dir     = local.kube_certs_dir
    cluster_dns_service_ip = cidrhost(var.service_cidr, 2)
  })

  workers_kubeconfig_kubelet = {
    for worker in local.workers_set :
    worker => templatefile("${path.module}/templates/workers/kubeconfig-kubelet.tftpl", {
      ca_crt              = base64encode(module.ca.ca_cert)
      node_name           = worker
      cluster_name        = var.cluster_name
      kubelet_crt         = base64encode(module.workers-kubelet[worker].cert)
      kubelet_key         = base64encode(module.workers-kubelet[worker].key)
      lb_apiserver_address = local.lb_apiserver_address
    })
  }

  workers_cloud_init = {
    for worker in local.workers_set :
    worker => templatefile("${path.module}/templates/workers/cloud-init.yaml.tftpl", {
      fqdn                = "${worker}.${var.domain}"
      domain              = var.domain
      kube_version        = var.kube_version
      runc_version        = var.runc_version
      containerd_version  = var.containerd_version
      kube_config_dir     = local.kube_config_dir
      kube_certs_dir      = local.kube_certs_dir
      arch                = var.architecture

      # install packages
      packages            = jsonencode([
        "socat",
        "conntrack",
        "ipset",
        "net-tools",
      ])

      # write files
      dns_config          = base64encode(local.resolved_config)
      containerd_config   = base64encode(file("${path.module}/files/containerd.toml"))
      ssh_public_key      = data.aws_key_pair.ssh_key.public_key
      systemd_units       = jsonencode([
        {
          name    = "containerd"
          content = base64encode(local.workers_containerd_systemd_unit)
        },
        {
          name    = "kubelet"
          content = base64encode(local.workers_kubelet_systemd_unit)
        },
      ])
      kube_configs        = jsonencode([
        {
          name    = "kubelet.kubeconfig"
          content = base64encode(local.workers_kubeconfig_kubelet[worker])
        },
        {
          name    = "kubelet-config.yaml"
          content = base64encode(local.workers_kubelet_config)
        }
      ])
      kube_certs          = jsonencode([
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
          content = base64encode(module.workers-kubelet[worker].cert)
        },
        {
          name    = "kubelet.key"
          content = base64encode(module.workers-kubelet[worker].key)
        }
      ])
    })
  }
}
