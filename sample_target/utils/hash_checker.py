import hashlib

# =========================================================
# ACTIVE EXECUTABLE CODE (Should be detected as CRITICAL)
# =========================================================
def generate_legacy_file_hash(data_bytes):
    # Cryptographically broken SHA-1 digest
    sha1_hash = hashlib.sha1(data_bytes).hexdigest()
    
    # Cryptographically broken MD5 digest
    md5_hash = hashlib.md5(data_bytes).hexdigest()
    
    return sha1_hash, md5_hash

def compute_safe_sha256(data_bytes):
    return hashlib.sha256(data_bytes).hexdigest()

# =========================================================
# COMMENTED-OUT CODE (AST Parser MUST IGNORE THESE!)
# =========================================================
# # hashlib.sha1(deprecated_token_bytes).hexdigest()
# # hashlib.md5(old_password_hash).hexdigest()
