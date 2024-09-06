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
    O               = "system:masters"
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

module "kube-apiserver" {
    source          = "./submodules/tls-generator"
    ca_cert		    = module.ca.ca_cert
    ca_key		    = module.ca.ca_key

    CN			    = "kubernetes"
    O			    = ""
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
        "api-server.kubernetes.local",
        "apiserver.kubernetes.local",
        "kube-apiserver-${var.cluster_name}.${var.domain}",
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
    dns_names = local.etcd_nodes

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
    dns_names = local.etcd_nodes

    validity_period	= 8760
}

module "controllers-kubelet" {
    source          = "./submodules/tls-generator"
    for_each        = toset(local.controllers_set)
    ca_cert		    = module.ca.ca_cert
    ca_key		    = module.ca.ca_key

    CN			    = "system:node:${each.value}"
    O			    = "system:nodes"
    OU			    = var.cluster_name
    C		        = "United Kingdom"
    ST              = "London"
    L		        = "London"

    ip_addresses = ["127.0.0.1"]
    dns_names = [each.value]

    validity_period	= 8760
}

module "workers-kubelet" {
    source          = "./submodules/tls-generator"
    for_each        = toset(local.workers_set)
    ca_cert		    = module.ca.ca_cert
    ca_key		    = module.ca.ca_key

    CN			    = "system:node:${each.value}"
    O			    = "system:nodes"
    OU			    = var.cluster_name
    C		        = "United Kingdom"
    ST              = "London"
    L		        = "London"

    ip_addresses = ["127.0.0.1"]
    dns_names = [each.value]

    validity_period	= 8760
}
