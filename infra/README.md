- [Dependecies](#sec-1)
- [Initial things](#sec-2)
  - [Configure aws cli](#sec-2-1)
- [Deploying](#sec-3)
  - [Init](#sec-3-1)
  - [Plan](#sec-3-2)
  - [Apply](#sec-3-3)
  - [Configure kubectl to use the new cluster](#sec-3-4)

Terraform configuration for prow.cncf.io

# Dependecies<a id="sec-1"></a>

-   terraform
-   aws cli

# Initial things<a id="sec-2"></a>

## Configure aws cli<a id="sec-2-1"></a>

Set up the aws cli with your aws account.

```shell
aws configure
```

# Deploying<a id="sec-3"></a>

## Init<a id="sec-3-1"></a>

Initialize terraform with the plugins that are required.

```shell
terraform init modules/aws-project
```

    [0m[1mInitializing modules...[0m
    Downloading terraform-aws-modules/eks/aws 12.2.0 for eks...
    - eks in .terraform/modules/eks/terraform-aws-eks-12.2.0
    - eks.node_groups in .terraform/modules/eks/terraform-aws-eks-12.2.0/modules/node_groups
    Downloading terraform-aws-modules/vpc/aws 2.6.0 for vpc...
    - vpc in .terraform/modules/vpc/terraform-aws-vpc-2.6.0
    
    [0m[1mInitializing the backend...[0m
    
    [0m[1mInitializing provider plugins...[0m
    - Checking for available provider plugins...
    - Downloading plugin for provider "random" (hashicorp/random) 2.3.0...
    - Downloading plugin for provider "local" (hashicorp/local) 1.4.0...
    - Downloading plugin for provider "null" (hashicorp/null) 2.1.2...
    - Downloading plugin for provider "template" (hashicorp/template) 2.1.2...
    - Downloading plugin for provider "kubernetes" (hashicorp/kubernetes) 1.12.0...
    - Downloading plugin for provider "aws" (hashicorp/aws) 3.0.0...
    
    [0m[1m[32mTerraform has been successfully initialized![0m[32m[0m
    [0m[32m
    You may now begin working with Terraform. Try running "terraform plan" to see
    any changes that are required for your infrastructure. All Terraform commands
    should now work.
    
    If you ever set or change modules or backend configuration for Terraform,
    rerun this command to reinitialize your working directory. If you forget, other
    commands will detect it and remind you to do so if necessary.[0m

## Plan<a id="sec-3-2"></a>

Using plan, verify that the actions performed will be the correct ones.

```shell
terraform plan modules/aws-project
```

## Apply<a id="sec-3-3"></a>

Create the infrastructure using apply.

```shell
terraform apply -auto-approve modules/aws-project
```

    [0m[1mrandom_string.suffix: Refreshing state... [id=1QQTdZBm][0m
    [0m[1mdata.aws_availability_zones.available: Refreshing state...[0m
    [0m[1mmodule.eks.data.aws_iam_policy_document.cluster_assume_role_policy: Refreshing state...[0m
    [0m[1mmodule.eks.data.aws_partition.current: Refreshing state...[0m
    [0m[1mmodule.eks.data.aws_ami.eks_worker: Refreshing state...[0m
    [0m[1mmodule.eks.data.aws_caller_identity.current: Refreshing state...[0m
    [0m[1mmodule.eks.data.aws_ami.eks_worker_windows: Refreshing state...[0m
    [0m[1mmodule.vpc.aws_vpc.this[0]: Refreshing state... [id=vpc-09d09edcefe600c80][0m
    [0m[1mmodule.eks.aws_iam_role.cluster[0]: Refreshing state... [id=prow-1QQTdZBm20200806034053065000000001][0m
    [0m[1mmodule.eks.data.aws_iam_policy_document.cluster_elb_sl_role_creation[0]: Refreshing state...[0m
    [0m[1mmodule.eks.data.aws_iam_policy_document.workers_assume_role_policy: Refreshing state...[0m
    [0m[1mmodule.vpc.aws_eip.nat[0]: Refreshing state... [id=eipalloc-03ecf40cc0a6ea2ec][0m
    [0m[1mmodule.eks.aws_iam_role_policy_attachment.cluster_AmazonEKSClusterPolicy[0]: Refreshing state... [id=prow-1QQTdZBm20200806034053065000000001-20200806034056216100000003][0m
    [0m[1mmodule.eks.aws_iam_role_policy_attachment.cluster_AmazonEKSServicePolicy[0]: Refreshing state... [id=prow-1QQTdZBm20200806034053065000000001-20200806034056260300000004][0m
    [0m[1mmodule.eks.aws_iam_role_policy.cluster_elb_sl_role_creation[0]: Refreshing state... [id=prow-1QQTdZBm20200806034053065000000001:prow-1QQTdZBm-elb-sl-role-creation20200806034055239100000002][0m
    [0m[1maws_security_group.worker_group_mgmt_one: Refreshing state... [id=sg-01688871b9149d976][0m
    [0m[1mmodule.eks.aws_security_group.cluster[0]: Refreshing state... [id=sg-0d5e8581a92fa2587][0m
    [0m[1maws_security_group.all_worker_mgmt: Refreshing state... [id=sg-04ec7dfd40f8c93c6][0m
    [0m[1maws_security_group.worker_group_mgmt_two: Refreshing state... [id=sg-02174dbc9ccc77b38][0m
    [0m[1mmodule.vpc.aws_route_table.public[0]: Refreshing state... [id=rtb-09633fa4fb997d997][0m
    [0m[1mmodule.vpc.aws_subnet.private[1]: Refreshing state... [id=subnet-02295f4d18e17ce34][0m
    [0m[1mmodule.eks.aws_security_group.workers[0]: Refreshing state... [id=sg-0c8752c7c87728331][0m
    [0m[1mmodule.vpc.aws_subnet.private[0]: Refreshing state... [id=subnet-07196814d21ec45ec][0m
    [0m[1mmodule.vpc.aws_subnet.private[2]: Refreshing state... [id=subnet-0cacfb79e8adee3ab][0m
    [0m[1mmodule.vpc.aws_internet_gateway.this[0]: Refreshing state... [id=igw-0c68c83539acbedab][0m
    [0m[1mmodule.vpc.aws_subnet.public[0]: Refreshing state... [id=subnet-084b685a63657c35d][0m
    [0m[1mmodule.vpc.aws_subnet.public[1]: Refreshing state... [id=subnet-0702883af1478a7d8][0m
    [0m[1mmodule.vpc.aws_subnet.public[2]: Refreshing state... [id=subnet-0127de1be4dc80d67][0m
    [0m[1mmodule.vpc.aws_route_table.private[0]: Refreshing state... [id=rtb-0b56391d70fc57067][0m
    [0m[1mmodule.vpc.aws_route.public_internet_gateway[0]: Refreshing state... [id=r-rtb-09633fa4fb997d9971080289494][0m
    [0m[1mmodule.eks.aws_security_group_rule.cluster_egress_internet[0]: Refreshing state... [id=sgrule-3279247683][0m
    [0m[1mmodule.eks.aws_security_group_rule.workers_egress_internet[0]: Refreshing state... [id=sgrule-1919393567][0m
    [0m[1mmodule.eks.aws_security_group_rule.cluster_https_worker_ingress[0]: Refreshing state... [id=sgrule-3296179297][0m
    [0m[1mmodule.eks.aws_security_group_rule.workers_ingress_cluster_https[0]: Refreshing state... [id=sgrule-2406842100][0m
    [0m[1mmodule.eks.aws_security_group_rule.workers_ingress_cluster[0]: Refreshing state... [id=sgrule-3971114604][0m
    [0m[1mmodule.eks.aws_security_group_rule.workers_ingress_self[0]: Refreshing state... [id=sgrule-1334406211][0m
    [0m[1mmodule.vpc.aws_route_table_association.private[1]: Refreshing state... [id=rtbassoc-0357a1307b25e86ed][0m
    [0m[1mmodule.vpc.aws_route_table_association.private[0]: Refreshing state... [id=rtbassoc-027165941fcf7e5e7][0m
    [0m[1mmodule.vpc.aws_route_table_association.private[2]: Refreshing state... [id=rtbassoc-0b527b8df38672fbb][0m
    [0m[1mmodule.vpc.aws_route_table_association.public[0]: Refreshing state... [id=rtbassoc-021586ab725bef8de][0m
    [0m[1mmodule.vpc.aws_route_table_association.public[1]: Refreshing state... [id=rtbassoc-0a0dc4c138dbc530b][0m
    [0m[1mmodule.vpc.aws_route_table_association.public[2]: Refreshing state... [id=rtbassoc-0a45668521e7a95ce][0m
    [0m[1mmodule.vpc.aws_nat_gateway.this[0]: Refreshing state... [id=nat-09ab5fd1401235f6f][0m
    [0m[1mmodule.eks.aws_eks_cluster.this[0]: Refreshing state... [id=prow-1QQTdZBm][0m
    [0m[1mmodule.vpc.aws_route.private_nat_gateway[0]: Refreshing state... [id=r-rtb-0b56391d70fc570671080289494][0m
    [0m[1mmodule.eks.null_resource.wait_for_cluster[0]: Refreshing state... [id=5629724694957061585][0m
    [0m[1mmodule.eks.aws_iam_role.workers[0]: Refreshing state... [id=prow-1QQTdZBm2020080603513072090000000a][0m
    [0m[1mdata.aws_eks_cluster_auth.cluster: Refreshing state...[0m
    [0m[1mmodule.eks.local_file.kubeconfig[0]: Refreshing state... [id=7ae36342cf476d389a6c2b489df08d1711f6f21f][0m
    [0m[1mdata.aws_eks_cluster.cluster: Refreshing state...[0m
    [0m[1mmodule.eks.data.template_file.userdata[0]: Refreshing state...[0m
    [0m[1mmodule.eks.aws_iam_role_policy_attachment.workers_AmazonEKSWorkerNodePolicy[0]: Refreshing state... [id=prow-1QQTdZBm2020080603513072090000000a-2020080603513349630000000d][0m
    [0m[1mmodule.eks.aws_iam_role_policy_attachment.workers_AmazonEKS_CNI_Policy[0]: Refreshing state... [id=prow-1QQTdZBm2020080603513072090000000a-2020080603513351970000000e][0m
    [0m[1mmodule.eks.aws_iam_instance_profile.workers[0]: Refreshing state... [id=prow-1QQTdZBm2020080603513273300000000b][0m
    [0m[1mmodule.eks.aws_iam_role_policy_attachment.workers_AmazonEC2ContainerRegistryReadOnly[0]: Refreshing state... [id=prow-1QQTdZBm2020080603513072090000000a-2020080603513348130000000c][0m
    [0m[1mmodule.eks.aws_launch_configuration.workers[0]: Refreshing state... [id=prow-1QQTdZBm-prow-worker-12020080603513632450000000f][0m
    [0m[1mmodule.eks.kubernetes_config_map.aws_auth[0]: Refreshing state... [id=kube-system/aws-auth][0m
    [0m[1mmodule.eks.random_pet.workers[0]: Refreshing state... [id=ruling-hornet][0m
    [0m[1mmodule.eks.aws_autoscaling_group.workers[0]: Refreshing state... [id=prow-1QQTdZBm-prow-worker-120200806035143147500000010][0m
    [0m[1m[32m
    Apply complete! Resources: 0 added, 0 changed, 0 destroyed.[0m
    [0m[1m[32m

## Configure kubectl to use the new cluster<a id="sec-3-4"></a>

Find the cluster name:

```shell
aws eks list-clusters
```

    ---------------------
    |   ListClusters    |
    +-------------------+
    ||    clusters     ||
    |+-----------------+|
    ||  prow-cluster   ||
    |+-----------------+|

Set current context to be the newly created cluster

```shell
aws eks --region ap-southeast-2 update-kubeconfig --name prow-cluster
```
