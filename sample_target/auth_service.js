const crypto = require('crypto');
const jwt = require('jsonwebtoken');

// =========================================================
// 1. ACTIVE EXECUTABLE CRYPTO CODE (Should be detected)
// =========================================================
function generateAuthKeys() {
    return crypto.generateKeyPairSync('rsa', {
        modulusLength: 2048,
        publicKeyEncoding: { type: 'spki', format: 'pem' },
        privateKeyEncoding: { type: 'pkcs8', format: 'pem' }
    });
}

function signUserToken(user, privateKey) {
    return jwt.sign({ id: user.id, email: user.email }, privateKey, {
        algorithm: 'RS256',
        expiresIn: '2h'
    });
}

function encryptSessionData(data, key, iv) {
    const cipher = crypto.createCipheriv('aes-256-gcm', key, iv);
    return cipher.update(data, 'utf8', 'hex') + cipher.final('hex');
}

// =========================================================
// 2. COMMENTED-OUT CRYPTO CODE (AST Parser MUST IGNORE THESE!)
// =========================================================
// // const legacyCommentedRsa = crypto.generateKeyPairSync('rsa');
// /* const oldUnusedJwt = jwt.sign({ id: 1 }, key, { algorithm: 'RS256' }); */
// // crypto.createHash('sha1').update('dummy').digest('hex');

module.exports = { generateAuthKeys, signUserToken, encryptSessionData };
