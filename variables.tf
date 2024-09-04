variable "domain" {
    type = string
    default = "compute.zone"
}

variable "cluster_name" {
    type = string
    default = "mytestcluster" 
}

variable "etcd_full_version" {
    type = string
    default = "etcd-v3.5.15-linux-amd64"
}

variable "architecture" {
    type = string
    default = "amd64"
}

variable "etcd_version" {
    type = string
    default = "v3.5.15"
}

variable "kube_version" {
    type = string
    default = "v1.31.0"
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

variable "nodes_instance_type" {
    type = string
    default = "t3.medium" 
}

variable "controllers_instance_type" {
    type = string
    default = "t3.medium" 
}

variable "subnet_id" {
    type = string
    default = "subnet-6c9a2b25"
}

variable "key_name" {
    type = string
    default = "capi-demo"
}

variable "controllers_count" {
    type = number
    default = 3
}

variable "nodes_count" {
    type = number
    default = 3
}
