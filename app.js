async function encrypt(text) {
    const key = await crypto.subtle.generateKey(
        { name: "AES-GCM", length: 256 }, true, ["encrypt", "decrypt"]);
    const nonce = crypto.getRandomValues(new Uint8Array(12));

    const ciphertext = await crypto.subtle.encrypt(
        { name: "AES-GCM", iv: nonce }, key, new TextEncoder().encode(text));

    const blob = new Uint8Array(nonce.length + ciphertext.byteLength);
    blob.set(nonce, 0);
    blob.set(new Uint8Array(ciphertext), nonce.length);

    const rawKey = await crypto.subtle.exportKey("raw", key);
    return { key: toBase64url(rawKey), blob };
}

async function decrypt(blob, keyStr) {
    const key = await crypto.subtle.importKey("raw", fromBase64url(keyStr), { name: "AES-GCM" }, false, ["decrypt"]);

    const nonce = blob.slice(0, 12);
    const ciphertext = blob.slice(12);
    const plaintext = await crypto.subtle.decrypt(
        { name: "AES-GCM", iv: nonce }, key, ciphertext);

    return new TextDecoder().decode(plaintext);
}

function toBase64url(buf) {
    let bin = "";
    for (const b of new Uint8Array(buf)) bin += String.fromCharCode(b);
    return btoa(bin).replaceAll("+", "-").replaceAll("/", "_").replaceAll("=", "");
}

function fromBase64url(str) {
    const bin = atob(str.replaceAll("-", "+").replaceAll("_", "/"));
    const bytes = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
    return bytes;
}