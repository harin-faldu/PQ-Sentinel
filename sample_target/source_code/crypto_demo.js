// Node.js & WebCrypto PQC Test File
const crypto = require('crypto');
const jwt = require('jsonwebtoken');

// 1. Insecure MD5 Hash (CRITICAL Risk)
function legacyHash(password) {
    return crypto.createHash('md5').update(password).digest('hex');
}

// 2. Deprecated SHA-1 Hash (HIGH Risk)
function legacySignature(data) {
    return crypto.createHash('sha1').update(data).digest('hex');
}

// 3. Classical Quantum-Vulnerable RSA Key Pair Gen (HIGH Risk)
function generateRSAKey() {
    return crypto.generateKeyPairSync('rsa', { modulusLength: 2048 });
}

// 4. Vulnerable JWT RS256 Signing (HIGH Risk)
function signUserToken(payload) {
    return jwt.sign(payload, 'secret-key', { algorithm: 'RS256' });
}

// 5. Quantum-Safe AES-256-GCM (SAFE)
function encryptSafe(text, key, iv) {
    const cipher = crypto.createCipheriv('aes-256-gcm', key, iv);
    return cipher.update(text, 'utf8', 'hex') + cipher.final('hex');
}

module.exports = { legacyHash, legacySignature, generateRSAKey, signUserToken, encryptSafe };
