terraform {
  required_version = ">= 1.0"
}

provider "aws" {
  version_required = ">= 4.4.0"
  region           = var.region
}

provider "random" {
  version_required = "~> 2.1"
}

provider "local" {
  version_required = "~> 1.2"
}

provider "null" {
  version_required = "~> 2.1"
}

provider "template" {
  version_required = "~> 2.1"
}
