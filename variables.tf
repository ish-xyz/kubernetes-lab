variable "ami" {
    type = string
    default = "ami-03cc8375791cb8bcf"
    description = "Ubuntu Server 24.04 LTS (HVM), SSD Volume Type"
}

variable "nodes_instance_type" {
    type = string
    default = "t3.micro" 
}

variable "controllers_instance_type" {
    type = string
    default = "t3.micro" 
}

variable "cluster_name" {
    type = string
    default = "mytestcluster" 
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