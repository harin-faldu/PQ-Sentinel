# Python PQC Test File
import hashlib
from cryptography.hazmat.primitives.asymmetric import rsa, ec
from Crypto.Cipher import DES

# 1. Broken MD5 & SHA-1
def broken_hashes(data):
    h1 = hashlib.md5(data.encode()).hexdigest()
    h2 = hashlib.sha1(data.encode()).hexdigest()
    return h1, h2

# 2. Quantum-vulnerable RSA key generation
def generate_rsa():
    return rsa.generate_private_key(public_exponent=65537, key_size=2048)

# 3. Quantum-vulnerable ECDSA key generation
def generate_ecc():
    return ec.generate_private_key(ec.SECP256R1())

# 4. Obsolete DES Cipher
def legacy_des_encrypt(key, data):
    cipher = DES.new(key, DES.MODE_ECB)
    return cipher.encrypt(data)
