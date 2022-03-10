module "eks" {
  source          = "terraform-aws-modules/eks/aws"
  version         = "18.9.0"
  cluster_name    = local.cluster_name
  cluster_version = "1.21"
  subnet_ids      = concat(sort(data.aws_subnet_ids.private.ids))

  tags = {
    GithubRepo = "terraform-aws-eks"
    GithubOrg  = "terraform-aws-modules"
  }

  vpc_id = module.vpc.vpc_id

  worker_groups = [
    {
      name          = "prow-worker-1"
      instance_type = "t3.medium"
      # additional_userdata           = "echo foo bar"
      additional_security_group_ids = [aws_security_group.worker_group_mgmt_two.id]
      asg_desired_capacity          = 3
    },
  ]
}

data "aws_eks_cluster" "cluster" {
  name = module.eks.cluster_id
}

data "aws_eks_cluster_auth" "cluster" {
  name = module.eks.cluster_id
}
