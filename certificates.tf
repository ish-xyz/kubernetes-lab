module "ca" {
    source          = "./submodules/tls-generator"
    CN			    = "CA"
    O               = ""
    OU			    = var.cluster_name
    C		        = "United Kingdom"
    ST              = "London"
    L		        = "London"
    validity_period	= 8760
}

module "admin" {
    source          = "./submodules/tls-generator"
    ca_cert		    = module.ca.ca_cert
    ca_key		    = module.ca.ca_key

    CN			    = "admin"
    O               = ""
    OU			    = var.cluster_name
    C		        = "United Kingdom"
    ST              = "London"
    L		        = "London"
    validity_period	= 8760
}

module "service-accounts" {
    source          = "./submodules/tls-generator"
    ca_cert		    = module.ca.ca_cert
    ca_key		    = module.ca.ca_key

    CN			    = "service-accounts"
    O               = ""
    OU			    = var.cluster_name
    C		        = "United Kingdom"
    ST              = "London"
    L		        = "London"
    validity_period	= 8760
}

module "kube-controller-manager" {
    source          = "./submodules/tls-generator"
    ca_cert		    = module.ca.ca_cert
    ca_key		    = module.ca.ca_key

    CN			    = "system:kube-controller-manager"
    O			    = "system:kube-controller-manager"
    OU			    = var.cluster_name
    C		        = "United Kingdom"
    ST              = "London"
    L		        = "London"

    validity_period	= 8760
}

module "kube-proxy" {
    source          = "./submodules/tls-generator"
    ca_cert		    = module.ca.ca_cert
    ca_key		    = module.ca.ca_key

    CN			    = "system:kube-proxy"
    O			    = "system:node-proxier"
    OU			    = var.cluster_name
    C		        = "United Kingdom"
    ST              = "London"
    L		        = "London"

    validity_period	= 8760
}

module "kube-scheduler" {
    source          = "./submodules/tls-generator"
    ca_cert		    = module.ca.ca_cert
    ca_key		    = module.ca.ca_key

    CN			    = "system:kube-scheduler"
    O               = "system:system:kube-scheduler"
    OU			    = var.cluster_name
    C		        = "United Kingdom"
    ST              = "London"
    L		        = "London"

    validity_period	= 8760
}

module "kube-api-server" {
    source          = "./submodules/tls-generator"
    ca_cert		    = module.ca.ca_cert
    ca_key		    = module.ca.ca_key

    CN			    = "system:kube-scheduler"
    O			    = "system:system:kube-scheduler"
    OU			    = var.cluster_name
    C		        = "United Kingdom"
    ST              = "London"
    L		        = "London"

    ip_addresses = [
        "127.0.0.1"
    ]
    dns_names = [
        "kubernetes",
        "kubernetes.default",
        "kubernetes.default.svc",
        "kubernetes.default.svc.cluster",
        "kubernetes.svc.cluster.local",
        "server.kubernetes.local",
        "api-server.kubernetes.local"
    ]

    validity_period	= 8760
}

module "etcd-peer" {
    source          = "./submodules/tls-generator"
    ca_cert		    = module.ca.ca_cert
    ca_key		    = module.ca.ca_key

    CN			    = "etcd-peer"
    O			    = ""
    OU			    = var.cluster_name
    C		        = "United Kingdom"
    ST              = "London"
    L		        = "London"

    ip_addresses = ["127.0.0.1"]
    dns_names = local.etcd_nodes_fqdns

    validity_period	= 8760

    #     cert_filename = "${path.root}/files/tmpcerts/etcd-client.crt"
    #     key_filename = "${path.root}/files/tmpcerts/etcd-client.key"
}


module "etcd-client" {
    source          = "./submodules/tls-generator"
    ca_cert		    = module.ca.ca_cert
    ca_key		    = module.ca.ca_key

    CN			    = "etcd-client"
    O			    = ""
    OU			    = var.cluster_name
    C		        = "United Kingdom"
    ST              = "London"
    L		        = "London"

    ip_addresses = ["127.0.0.1"]
    dns_names = local.etcd_nodes_fqdns

    validity_period	= 8760

    #     cert_filename = "${path.root}/files/tmpcerts/etcd-client.crt"
    #     key_filename = "${path.root}/files/tmpcerts/etcd-client.key"
}

# resource "local_file" "test-ca-file-key" {
# 	content         = module.ca.ca_key
# 	filename        = "${path.root}/files/tmpcerts/ca.key"
# 	file_permission = "0644"
# }

# resource "local_file" "test-ca-file-crt" {
# 	content         = module.ca.ca_cert
# 	filename        = "${path.root}/files/tmpcerts/ca.crt"
# 	file_permission = "0644"
# }
