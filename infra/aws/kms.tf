resource "aws_kms_key" "eks_encryption" {
  description = "This key is used to encrypt Kubernetes secrets"
}

resource "aws_kms_alias" "prow-cncf-io-kms-key" {
  name          = "alias/prow-cncf-io-kms-key"
  target_key_id = aws_kms_key.eks_encryption.key_id
}
