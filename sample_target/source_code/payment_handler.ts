// TypeScript WebCrypto API scanner test module

export async function processPaymentHash(paymentData: ArrayBuffer): Promise<ArrayBuffer> {
    // WebCrypto subtle digest SHA-1 test call
    return await window.crypto.subtle.digest('SHA-1', paymentData);
}

export async function generatePaymentKey(): Promise<CryptoKeyPair> {
    // WebCrypto subtle generateKey RSA 1024-bit test call
    return await window.crypto.subtle.generateKey(
        {
            name: 'RSASSA-PKCS1-v1_5',
            modulusLength: 1024,
            publicExponent: new Uint8Array([1, 0, 1]),
            hash: 'SHA-256',
        },
        true,
        ['sign', 'verify']
    );
}
