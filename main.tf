# DNS Config
data "aws_route53_zone" "compute_zone" {
  zone_id      = var.route53_zone_id
}

data "dns_a_record_set" "name_servers" {
  for_each = toset(data.aws_route53_zone.compute_zone.name_servers)
  host = each.value
}

# S3 for Cloud-init configs
resource "aws_s3_bucket" "config_bucket" {
  bucket = "cloud-init-configurations"
  force_destroy = true
}