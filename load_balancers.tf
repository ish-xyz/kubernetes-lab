# Compute resources
resource "aws_instance" "load_balancers" {
  for_each                    = toset(local.load_balancers_set)
  ami                         = var.ami
  instance_type               = var.load_balancers_instance_type
  subnet_id                   = var.subnet_id
  key_name                    = var.key_name
  user_data_replace_on_change = true
  user_data_base64            = base64gzip(local.load_balancer_cloud_init[each.key])
  tags = {
    Name = each.value
    FQDN = "${each.value}.${var.domain}"
    Role = "load-balancer"
    Cluster = "${var.cluster_name}"
  }
}

resource "aws_route53_record" "load_balancers" {
  for_each  = aws_instance.load_balancers
  zone_id   = data.aws_route53_zone.compute_zone.zone_id
  name      = each.value.tags["Name"]
  type      = "A"
  ttl       = 300
  records   = [each.value.private_ip]
}

resource "aws_route53_record" "kube_apiserver_external" {
  for_each = aws_route53_record.load_balancers

  zone_id = data.aws_route53_zone.compute_zone.zone_id
  name    = "${local.lb_apiserver_address}"
  type    = "CNAME"
  ttl     = 300
  records = ["${each.value.name}.${var.domain}"]

  weighted_routing_policy {
      weight = 100
  }
  set_identifier = "kube_apiserver_to_lb_${each.key}"
}

locals {
  load_balancer_haproxy_cfg = templatefile("${path.module}/templates/load-balancers/haproxy.cfg.tftpl", {
    backend_servers = jsonencode(local.controllers_set)
    domain          = var.domain
  })

  load_balancer_cloud_init = {
    for lb in local.load_balancers_set :
    lb => templatefile("${path.module}/templates/load-balancers/cloud-init.yaml.tftpl", {
      fqdn            = "${lb}.${var.domain}"
      packages        = jsonencode(["haproxy", "net-tools"])
      resolved_config = base64encode(local.resolved_config)
      ssh_public_key  = data.aws_key_pair.ssh_key.public_key
      haproxy_config  = base64encode(local.load_balancer_haproxy_cfg)
    })
  }
}
