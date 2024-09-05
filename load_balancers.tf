# Compute resources
resource "aws_instance" "load_balancers" {
  for_each                    = toset(local.load_balancers_set)
  ami                         = var.ami
  instance_type               = var.load_balancers_instance_type
  subnet_id                   = var.subnet_id
  key_name                    = var.key_name
  user_data_replace_on_change = true
  user_data                   = ""
  tags = {
    Name = each.value
    FQDN = "${each.value}.${var.domain}"
    Role = "load-balancer"
    Cluster = "${var.cluster_name}"
  }
}
