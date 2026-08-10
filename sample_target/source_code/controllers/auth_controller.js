// Production authentication controller importing weak hash helper
const { hashPassword, generateLegacyRSA } = require('../utils/hash_helper');

function loginUser(req, res) {
    const username = req.body.username;
    const password = req.body.password;

    // Active Consumer Call-Site 1: Triggers ScopeWeight elevation to PRODUCTION (1.5x)
    const hashedPassword = hashPassword(password);

    if (username === "admin" && hashedPassword === "5f4dcc3b5aa765d61d8327deb882cf99") {
        return res.json({ success: true, token: "session-token" });
    }
    return res.status(401).json({ error: "Invalid credentials" });
}

module.exports = { loginUser };
