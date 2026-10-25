const NotesApp = function () {
   const DOM = {
      form: document.querySelector(".js-auth-form"),
      btnLogin: document.querySelector(".js-auth-form-login"),
      btnRegister: document.querySelector(".js-auth-form-register"),
      username: document.querySelector(".js-auth-form-username"),
      password: document.querySelector(".js-auth-form-password"),
      errorModal: document.querySelector(".js-error-modal"),
      errorModalText: document.querySelector(".js-error-modal__text"),
      errorModalBtn: document.querySelector(".js-error-modal__btn"),
   };

   function bindEvents() {
      DOM.btnLogin.addEventListener("click", (event) => {
         login(event);
      });
      DOM.btnRegister.addEventListener("click", (event) => {
         register(event);
      });

      // Modal
      DOM.errorModalBtn.addEventListener("click", () => {
         DOM.errorModal.hidden = true;
      });
   }

   async function login(event) {
      event.preventDefault();

      const data = {
         username: DOM.username.value,
         password: DOM.password.value,
      };

      if (data.username === "" || data.password === "") {
         showError("Invalid username or password");
         return;
      }

      const response = await fetch("/api/auth/login", {
         method: "POST",
         body: JSON.stringify(data),
      });

      if (!response.ok) {
         const res = await response.text();
         showError(response.statusText + ": " + res);
         return;
      }

      await setNotesKey(data.password);
      window.location.href = "/";
   }

   async function register(event) {
      event.preventDefault();

      const data = {
         username: DOM.username.value,
         password: DOM.password.value,
      };

      if (data.username === "" || data.password === "") {
         showError("Invalid username or password");
         return;
      }

      const response = await fetch("/api/auth/register", {
         method: "POST",
         body: JSON.stringify(data),
      });

      if (!response.ok) {
         const res = await response.text();
         showError(response.statusText + ": " + res);
         return;
      }

      await setNotesKey(data.password);
      window.location.href = "/";
   }

   function showError(message) {
      DOM.errorModalText.textContent = message;
      DOM.errorModal.hidden = false;
      DOM.errorModalBtn.focus();
   }

   async function setNotesKey(password) {
      // 1. Получаем соль с сервера
      const response = await fetch("/api/auth/me", { method: "GET" });
      if (!response.ok) {
         const res = await response.text();
         showError(response.statusText + ": " + res);
         return null;
      }
      const res = await response.json();
      const saltBase64 = res.notesSalt;
      if (!saltBase64) {
         showError("Server did not return notesSalt");
         return null;
      }

      // 2. Деривация ключа
      const enc = new TextEncoder();
      const salt = Uint8Array.from(atob(saltBase64), (c) => c.charCodeAt(0));

      const keyMaterial = await crypto.subtle.importKey(
         "raw",
         enc.encode(password),
         "PBKDF2",
         false,
         ["deriveKey"],
      );

      const key = await crypto.subtle.deriveKey(
         {
            name: "PBKDF2",
            salt,
            iterations: 600000,
            hash: "SHA-256",
         },
         keyMaterial,
         { name: "AES-GCM", length: 256 },
         true, // ← было false: без этого экспорт невозможен
         ["encrypt", "decrypt"],
      );

      // 3. Экспортируем в JWK и сохраняем строку
      const jwk = await crypto.subtle.exportKey("jwk", key);
      sessionStorage.setItem("notesKey", JSON.stringify(jwk));

      return key;
   }

   return {
      init: () => {
         bindEvents();
      },
   };
};

NotesApp().init();
