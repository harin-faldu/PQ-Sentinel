// Helper module defining weak legacy hashing function
const crypto = require('crypto');

export function hashPassword(plainText) {
    // Weak legacy MD5 hash calculation
    return crypto.createHash('md5').update(plainText).digest('hex');
}

export function generateLegacyRSA() {
    return crypto.generateKeyPairSync('rsa', {
        modulusLength: 1024,
    });
}
