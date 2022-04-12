############### VPC variables
variable "name" {
  type        = string
  default     = "prow-cncf-io-eks"
  description = "A name for this stack."
}

variable "region" {
  type        = string
  default     = "ap-southeast-2"
  description = "Region where this stack will be deployed."
}

variable "cidr_block" {
  type        = string
  default     = "10.0.0.0/16"
  description = "The CIDR block for the VPC."
}

variable "availability_zones" {
  default     = ["ap-southeast-2a", "ap-southeast-2b", "ap-southeast-2c"]
  description = "The availability zones to create subnets in"
}

variable "az_counts" {
  default = 3
}
