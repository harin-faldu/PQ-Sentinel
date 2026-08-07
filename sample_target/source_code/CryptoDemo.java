package com.quantumvault.demo;

import javax.crypto.Cipher;
import java.security.KeyPairGenerator;
import java.security.MessageDigest;

public class CryptoDemo {
    public void demoLegacyCrypto() throws Exception {
        // 1. Weak MessageDigest
        MessageDigest md5 = MessageDigest.getInstance("MD5");
        MessageDigest sha1 = MessageDigest.getInstance("SHA-1");

        // 2. Quantum-vulnerable KeyPairGenerator
        KeyPairGenerator kpg = KeyPairGenerator.getInstance("RSA");
        kpg.initialize(2048);

        // 3. Classical Cipher Instance
        Cipher cipher = Cipher.getInstance("RSA/ECB/PKCS1Padding");
    }
}
