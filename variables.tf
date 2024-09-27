variable "domain" {
    type = string
    default = "compute.zone"
}

variable "cluster_domain" {
    type = string
    default = "cluster.local"
}

variable "cluster_dns_servers" {
    type = list
    default = []
}

variable "cluster_name" {
    type = string
    default = "mytestcluster" 
}

variable "pod_cidr" {
    type = string
    default = "10.200.0.0/16"
}

variable "service_cidr" {
    type = string
    default = "10.32.0.0/24"
}

variable "architecture" {
    type = string
    default = "amd64"
}

variable "helm_version" {
    type = string
    default = "v3.16.0"
}

variable "containerd_version" {
    type = string
    default = "1.7.16"
}

variable "runc_version" {
    type = string
    default = "v1.1.14"
}

variable "etcd_version" {
    type = string
    default = "v3.5.15"
}

variable "kube_version" {
    type = string
    default = "v1.31.0"
}

variable "etcd_full_version" {
    type = string
    default = "etcd-v3.5.15-linux-amd64"
}

variable "route53_zone_id" {
    type = string
    default = "Z0254661OBYGIWIKHFI3"
}

variable "aws_region" {
    type = string
    default = "eu-west-1"
}

variable "ami" {
    type = string
    default = "ami-03cc8375791cb8bcf"
    description = "Ubuntu Server 24.04 LTS (HVM), SSD Volume Type"
}

variable "subnet_id" {
    type = string
    default = "subnet-6c9a2b25"
}

variable "node_ports_range" {
    type = string
    default = "30000-32767"
}

variable "key_name" {
    type = string
    default = "capi-demo"
}

## Per component configs

variable "controllers_count" {
    type = number
    default = 3
}

variable "workers_count" {
    type = number
    default = 3
}

variable "load_balancers_count" {
    type = number
    default = 1
}

variable "load_balancers_instance_type" {
    type = string
    default = "t3.medium" 
}

variable "workers_instance_type" {
    type = string
    default = "t3.medium" 
}

variable "controllers_instance_type" {
    type = string
    default = "t3.medium" 
}

variable "etcd_key1" {
    type = string
    default = "ivV84gTtStZstvT3en7MVqNANfKKKU8vTFzl/N8MEM4="
}

variable "etcd_key2" {
    type = string
    default = "MZ5vNy7kCmfFAr7mnQj4yUV36d1qLnTCpSnK0NGGc0k="
}