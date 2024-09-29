## Systemd units templates

locals {
  controllers_bootstrap_manager_systemd_units = {
    for controller in local.controllers_set :
    controller => templatefile("${path.module}/templates/controllers/service-bootstrap-manager.tftpl", {})
  }

  controllers_etcd_systemd_units = {
    for controller in local.controllers_set :
    controller => templatefile("${path.module}/templates/controllers/service-etcd.tftpl", {
      etcd_certs_dir       = local.etcd_certs_dir
      etcd_cluster_token   = random_password.etcd_token.result
      etcd_name            = controller
      etcd_cluster_members = local.etcd_cluster_members
    })
  }

  controllers_kube_apiserver_systemd_unit = templatefile("${path.module}/templates/controllers/service-kube-apiserver.tftpl", {
    service_cidr       = var.service_cidr
    node_ports_range   = var.node_ports_range
    kube_certs_dir     = local.kube_certs_dir
    kube_config_dir    = local.kube_config_dir
    etcd_endpoints     = local.etcd_endpoints
    etcd_certs_dir     = local.etcd_certs_dir
  })

  controllers_kube_controller_manager_systemd_unit = templatefile("${path.module}/templates/controllers/service-kube-controller-manager.tftpl", {
    pod_cidr        = var.pod_cidr
    service_cidr    = var.service_cidr
    kube_certs_dir  = local.kube_certs_dir
    kube_config_dir = local.kube_config_dir
  })

  controllers_kube_scheduler_systemd_unit = templatefile("${path.module}/templates/controllers/service-kube-scheduler.tftpl", {
    kube_config_dir = local.kube_config_dir
  })

  controllers_kubelet_systemd_unit = templatefile("${path.module}/templates/controllers/service-kubelet.tftpl", {
    kube_config_dir = local.kube_config_dir
  })

  controllers_containerd_systemd_unit = templatefile("${path.module}/templates/controllers/service-containerd.tftpl", {})
}

# Controllers compute resources
resource "aws_instance" "controllers" {
  for_each                    = toset(local.controllers_set)
  ami                         = var.ami
  instance_type               = var.controllers_instance_type
  subnet_id                   = var.subnet_id
  key_name                    = var.key_name
  iam_instance_profile        = aws_iam_instance_profile.kube_nodes.name
  user_data_replace_on_change = true
  tags = {
    Name = each.value
    FQDN = "${each.value}.${var.domain}"
    Role = "controller"
    Cluster = "${var.cluster_name}"
  }

  user_data = <<-EOF
              #!/bin/bash
              # ${md5(local.controllers_cloud_init[each.value])}
              set -euo pipefail
              snap install aws-cli --classic
              cloud_config_url=$(aws s3 presign s3://${aws_s3_bucket.config_bucket.bucket}/${var.cluster_name}/${each.value}-config.yaml --expires-in 3600)
              curl -L -o /etc/cloud/cloud.cfg.d/custom.cfg $cloud_config_url
              [[ ! -f /custom-cloud-init-done ]] && cloud-init clean --logs --reboot
              EOF
}


# DNS Records for controllers
resource "aws_route53_record" "controllers" {
  for_each  = aws_instance.controllers
  zone_id   = data.aws_route53_zone.compute_zone.zone_id
  name      = each.value.tags["Name"]
  type      = "A"
  ttl       = 300
  records   = [each.value.private_ip]
}

resource "aws_s3_object" "controllers_cloud_init" {
  for_each = toset(local.controllers_set)
  bucket = aws_s3_bucket.config_bucket.id
  key    = "${var.cluster_name}/${each.value}-config.yaml"
  content = local.controllers_cloud_init[each.key]
}


## Components configurations

# ETCD token
resource "random_password" "etcd_token" {
  length           = 16
  special          = false
  lifecycle {
    create_before_destroy = true
  }
}

locals {
  etcd_encryption_config = templatefile("${path.module}/templates/controllers/encryption-config.yaml.tftpl", {
    key_1 = var.etcd_key1
    key_2 = var.etcd_key2
  })

  controllers_kubelet_config = templatefile("${path.module}/templates/controllers/kubelet-config.tftpl", {
    pod_cidr          = var.pod_cidr
    cluster_domain    = var.cluster_domain
    kube_certs_dir    = local.kube_certs_dir
    cluster_dns_service_ip = cidrhost(var.service_cidr, 2)
  })

  controllers_kube_scheduler_config = templatefile("${path.module}/templates/controllers/kube-scheduler-config.tftpl", {
    kube_config_dir = local.kube_config_dir
  })

  controllers_bootstrap_manager_configs = {
    for controller in local.controllers_set :
    controller => templatefile("${path.module}/templates/controllers/bootstrap-manager-config.yaml.tftpl", {
      node_name           = controller
      node_count          = var.controllers_count
      kube_config_dir     = local.kube_config_dir
      kube_certs_dir      = local.kube_certs_dir
      lb_apiserver_address = local.lb_apiserver_address
      cluster_dns_service_ip = cidrhost(var.service_cidr, 2)
    })
  }

  controllers_kube_apiserver_manifest = templatefile("${path.module}/templates/controllers/manifest-kube-apiserver.yaml.tftpl", {
    service_cidr       = var.service_cidr
    node_ports_range   = var.node_ports_range
    kube_certs_dir     = local.kube_certs_dir
    kube_config_dir    = local.kube_config_dir
    etcd_endpoints     = local.etcd_endpoints
    etcd_certs_dir     = local.etcd_certs_dir
    controllers_count  = var.controllers_count
    kube_version       = var.kube_version
  })

  controllers_kubeconfig_admin = templatefile("${path.module}/templates/controllers/kubeconfig-admin.tftpl", {
    ca_crt             = base64encode(module.ca.ca_cert)
    cluster_name       = var.cluster_name
    admin_crt          = base64encode(module.admin.cert)
    admin_key          = base64encode(module.admin.key)
    lb_apiserver_address = local.lb_apiserver_address
  })

  controllers_kubeconfig_controller_manager = templatefile("${path.module}/templates/controllers/kubeconfig-kube-controller-manager.tftpl", {
    ca_crt                   = base64encode(module.ca.ca_cert)
    cluster_name             = var.cluster_name
    kube_controller_manager_crt = base64encode(module.kube-controller-manager.cert)
    kube_controller_manager_key = base64encode(module.kube-controller-manager.key)
  })

  controllers_kubeconfig_kube_scheduler = templatefile("${path.module}/templates/controllers/kubeconfig-kube-scheduler.tftpl", {
    ca_crt            = base64encode(module.ca.ca_cert)
    cluster_name      = var.cluster_name
    kube_scheduler_crt = base64encode(module.kube-scheduler.cert)
    kube_scheduler_key = base64encode(module.kube-scheduler.key)
  })

  controllers_kubeconfig_kubelet = {
    for controller in local.controllers_set :
    controller => templatefile("${path.module}/templates/controllers/kubeconfig-kubelet.tftpl", {
      ca_crt                = base64encode(module.ca.ca_cert)
      node_name             = controller
      cluster_name          = var.cluster_name
      lb_apiserver_address  = local.lb_apiserver_address
      kubelet_crt           = base64encode(module.controllers-kubelet[controller].cert)
      kubelet_key           = base64encode(module.controllers-kubelet[controller].key)
    })
  }

  controllers_cloud_init = {
    for controller in local.controllers_set :
    controller => templatefile("${path.module}/templates/controllers/cloud-init.yaml.tftpl", {
      fqdn                  = "${controller}.${var.domain}"
      domain                = var.domain
      etcd_full_version     = var.etcd_full_version
      etcd_version          = var.etcd_version
      kube_version          = var.kube_version
      runc_version          = var.runc_version
      helm_version          = var.helm_version
      containerd_version    = var.containerd_version
      kube_config_dir       = local.kube_config_dir
      kube_certs_dir        = local.kube_certs_dir
      arch                  = var.architecture

      # install packages
      packages              = jsonencode([
        "socat",
        "conntrack",
        "ipset",
        "net-tools",
      ])

      # write files
      dns_config            = base64encode(local.resolved_config)
      hosts_config          = base64encode(file("${path.module}/files/hosts"))
      containerd_config     = base64encode(file("${path.module}/files/containerd.toml"))
      bootstrap_manager_config = base64encode(local.controllers_bootstrap_manager_configs[controller])
      ssh_public_key        = data.aws_key_pair.ssh_key.public_key
      systemd_units         = jsonencode([
        {
          name    = "etcd"
          content = base64encode(local.controllers_etcd_systemd_units[controller])
        },
        {
          name    = "kube-apiserver"
          content = base64encode(local.controllers_kube_apiserver_systemd_unit)
        },
        {
          name    = "kube-controller-manager"
          content = base64encode(local.controllers_kube_controller_manager_systemd_unit)
        },
        {
          name    = "kube-scheduler"
          content = base64encode(local.controllers_kube_scheduler_systemd_unit)
        },
        {
          name    = "containerd"
          content = base64encode(local.controllers_containerd_systemd_unit)
        },
        {
          name    = "kubelet"
          content = base64encode(local.controllers_kubelet_systemd_unit)
        },
        {
          name    = "bootstrap-manager"
          content = base64encode(local.controllers_bootstrap_manager_systemd_units[controller])
        },
      ])
      kube_configs          = jsonencode([
        {
          name    = "kubelet.kubeconfig"
          content = base64encode(local.controllers_kubeconfig_kubelet[controller])
        },
        {
          name    = "admin.kubeconfig"
          content = base64encode(local.controllers_kubeconfig_admin)
        },
        {
          name    = "kube-controller-manager.kubeconfig"
          content = base64encode(local.controllers_kubeconfig_controller_manager)
        },
        {
          name    = "kube-scheduler.kubeconfig"
          content = base64encode(local.controllers_kubeconfig_kube_scheduler)
        },
        {
          name    = "kube-scheduler.yaml"
          content = base64encode(local.controllers_kube_scheduler_config)
        },
        {
          name    = "kubelet-config.yaml"
          content = base64encode(local.controllers_kubelet_config)
        },
        {
          name    = "encryption-config.yaml"
          content = base64encode(local.etcd_encryption_config)
        },
        {
          name = "manifests/kube-apiserver.yaml"
          content = base64encode(local.controllers_kube_apiserver_manifest)
        },
        {
          name = "manifests/default-roles.yaml"
          content = base64encode(file("${path.module}/files/default-roles.yaml"))
        }
      ])
      kube_certs            = jsonencode([
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
          name    = "kube-apiserver.crt"
          content = base64encode(module.kube-apiserver.cert)
        },
        {
          name    = "kube-apiserver.key"
          content = base64encode(module.kube-apiserver.key)
        },
        {
          name    = "service-accounts.crt"
          content = base64encode(module.service-accounts.cert)
        },
        {
          name    = "service-accounts.key"
          content = base64encode(module.service-accounts.key)
        },
        {
          name    = "kubelet.crt"
          content = base64encode(module.controllers-kubelet[controller].cert)
        },
        {
          name    = "kubelet.key"
          content = base64encode(module.controllers-kubelet[controller].key)
        }
      ])
      etcd_certs            = jsonencode([
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
    })
  }
}
