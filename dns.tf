# DNS Config
data "aws_route53_zone" "compute_zone" {
  zone_id      = var.route53_zone_id
}

resource "aws_route53_record" "controllers" {
  for_each  = aws_instance.controllers
  zone_id   = data.aws_route53_zone.compute_zone.zone_id
  name      = each.value.tags["Name"]
  type      = "A"
  ttl       = 300
  records   = [each.value.private_ip]
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
  name    = "kube-apiserver-${var.cluster_name}"
  type    = "CNAME"
  ttl     = 300
  records = [each.value.name]

  weighted_routing_policy {
    weight = 100
  }
  set_identifier = "kube_apiserver_to_lb_${each.key}"
}

data "dns_a_record_set" "name_servers" {
  for_each = toset(data.aws_route53_zone.compute_zone.name_servers)
  host = each.value
}
