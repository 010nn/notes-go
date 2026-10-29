const PBKDF2_ITERATIONS = 600_000;
const SALT_BYTES = 16;

export class CryptoError extends Error {}

function base64ToBytes(b64) {
   try {
      return Uint8Array.from(atob(b64), (c) => c.charCodeAt(0));
   } catch {
      throw new CryptoError("Invalid salt from server");
   }
}

async function deriveKey(password, salt) {
   const enc = new TextEncoder();
   const keyMaterial = await crypto.subtle.importKey(
      "raw",
      enc.encode(password),
      "PBKDF2",
      false,
      ["deriveKey"],
   );

   return crypto.subtle.deriveKey(
      {
         name: "PBKDF2",
         salt,
         iterations: PBKDF2_ITERATIONS,
         hash: "SHA-256",
      },
      keyMaterial,
      { name: "AES-GCM", length: 256 },
      true,
      ["encrypt", "decrypt"],
   );
}

export async function createAndStoreNotesKey(password, saltBase64) {
   if (!saltBase64) {
      throw new CryptoError("Server did not return notesSalt");
   }

   const salt = base64ToBytes(saltBase64);
   if (salt.length !== SALT_BYTES) {
      throw new CryptoError("Unexpected salt length from server");
   }

   const key = await deriveKey(password, salt);
   const jwk = await crypto.subtle.exportKey("jwk", key);
   sessionStorage.setItem("notesKey", JSON.stringify(jwk));

   return key;
}

export async function restoreNotesKey() {
   const stored = sessionStorage.getItem("notesKey");
   if (!stored) return;
   try {
      const jwk = JSON.parse(stored);
      return await crypto.subtle.importKey(
         "jwk",
         jwk,
         { name: "AES-GCM", length: 256 },
         true,
         ["encrypt", "decrypt"],
      );
   } catch (e) {
      sessionStorage.removeItem("notesKey");
   }
   return null;
}

export async function encryptNote(note, notesKey) {
   if (!notesKey) throw "No key";

   const payload = JSON.stringify({
      title: note.title,
      content: note.content,
   });
   const iv = crypto.getRandomValues(new Uint8Array(12));

   const encrypted = await crypto.subtle.encrypt(
      { name: "AES-GCM", iv },
      notesKey,
      new TextEncoder().encode(payload),
   );

   const combined = new Uint8Array(12 + encrypted.byteLength);
   combined.set(iv, 0);
   combined.set(new Uint8Array(encrypted), 12);

   return btoa(String.fromCharCode(...combined));
}

export async function decryptNote(base64, notesKey) {
   if (!notesKey) throw "No key";

   const combined = Uint8Array.from(atob(base64), (c) => c.charCodeAt(0));
   const iv = combined.slice(0, 12);
   const data = combined.slice(12);

   const decrypted = await crypto.subtle.decrypt(
      { name: "AES-GCM", iv },
      notesKey,
      data,
   );

   return JSON.parse(new TextDecoder().decode(decrypted));
}
