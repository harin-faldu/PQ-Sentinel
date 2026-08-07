// Sample JS file with hardcoded secrets and crypto calls for CBOM testing

const crypto = require('crypto');
const jwt = require('jsonwebtoken');

// Hardcoded AWS Access Key ID
const AWS_ACCESS_KEY = "AKIAIOSFODNN7EXAMPLE";
const AWS_SECRET_KEY = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY";

// Hardcoded JWT Signing Secret
const token = jwt.sign({ user: 'admin' }, 'super-secret-hardcoded-key-12345');

// MD5 Hash Function (Legacy / Broken)
function hashUserPassword(password) {
    return crypto.createHash('md5').update(password).digest('hex');
}

// RSA Key Pair Generation
function generateUserKeys() {
    return crypto.generateKeyPairSync('rsa', {
        modulusLength: 1024,
    });
}

module.exports = { hashUserPassword, generateUserKeys };
