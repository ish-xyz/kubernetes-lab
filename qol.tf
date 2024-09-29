
resource "local_file" "admin_kubeconfig" {
  content  = local.controllers_kubeconfig_admin
  filename = "${path.module}/terraform-output-files/admin.kubeconfig"
}

resource "local_file" "hosts" {
  for_each = toset(local.load_balancers_set)
  content  = "${aws_instance.load_balancers[each.key].public_ip} kube-apiserver-${var.cluster_name}.${var.domain}"
  filename = "${path.module}/terraform-output-files/hosts-${each.key}"
}
