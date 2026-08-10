# Terraform AWS KMS Key Resources for QuantumVault CBOM Testing

resource "aws_kms_key" "payment_rsa_key" {
  description              = "Legacy RSA KMS Key for Payment Signatures"
  deletion_window_in_days  = 30
  enable_key_rotation      = false
  customer_master_key_spec = "RSA_2048"
  key_usage                = "SIGN_VERIFY"
}

resource "aws_kms_key" "db_symmetric_key" {
  description              = "Default Symmetric AES-256 KMS Key for Database Encryption"
  deletion_window_in_days  = 30
  enable_key_rotation      = true
  customer_master_key_spec = "SYMMETRIC_DEFAULT"
}
