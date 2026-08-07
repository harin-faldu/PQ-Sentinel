// Crypto Testbed File for CBOM Scanner 2.0 Verification

const crypto = require('crypto');

// 1. Multi-line function call spanning line boundaries (KeySize: 1024)
const legacyRSAKey = crypto.generateKeyPairSync(
    'rsa',
    {
        modulusLength: 1024
    }
);

// 2. Active Legacy Hash (MD5)
const legacyHash = crypto.createHash('md5').update('data').digest('hex');

// 3. Active Safe PQC-ready symmetric cipher
const safeCipher = crypto.createCipheriv('aes-256-gcm', key, iv);

// 4. Log message describing crypto method — must NOT be flagged as active asset!
console.log("Deprecated legacy hash used previously: crypto.createHash('md5')");

// 5. Multi-line block comment containing commented-out crypto — must NOT be flagged!
/*
  const oldKey = crypto.generateKeyPairSync('rsa', { modulusLength: 512 });
  const oldCipher = crypto.createCipheriv('des-cbc', k, i);
*/

// 6. Control flow check — must NOT be flagged!
if (algorithm === 'ML-DSA') {
    operationsObj.sign = true;
    operationsObj.verify = true;
}
