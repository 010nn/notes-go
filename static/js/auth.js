import { api, ApiError } from "./pkg/api.js";
import { createAndStoreNotesKey, CryptoError } from "./pkg/crypto.js";

document.addEventListener("alpine:init", () => {
   Alpine.data("authApp", authApp);
});

function authApp() {
   const USERNAME_MIN = 1;
   const USERNAME_MAX = 32;
   const PASSWORD_MIN = 3;
   const PASSWORD_MAX = 64;

   return {
      // === State ===
      username: "",
      password: "",
      loading: false,
      error: null,

      init() {
         this.$nextTick(() => this.$refs.username?.focus());
      },

      // public actions
      async submitLogin() {
         await this.submit(api.login);
      },

      async submitRegister() {
         await this.submit(api.register);
      },

      closeError() {
         this.error = null;
      },

      // internals
      async submit(request) {
         if (this.loading) return;

         const validationError = this.validate();
         if (validationError) {
            this.error = validationError;
            return;
         }

         this.loading = true;
         this.error = null;

         try {
            const { notesSalt } = await request(this.username, this.password);
            await createAndStoreNotesKey(this.password, notesSalt);
            sessionStorage.setItem("username", this.username);
            window.location.href = "/";
         } catch (e) {
            this.error =
               e instanceof ApiError || e instanceof CryptoError
                  ? e.message
                  : "Unexpected error: " + (e?.message ?? e);
         } finally {
            this.loading = false;
            this.$nextTick(() => this.$refs.errorOk?.focus());
         }
      },

      validate() {
         if (!this.username) return "Username is required";
         if (this.username.length < USERNAME_MIN)
            return `Username must be at least ${USERNAME_MIN} characters`;
         if (this.username.length > USERNAME_MAX)
            return `Username must be at most ${USERNAME_MAX} characters`;

         if (!this.password) return "Password is required";
         if (this.password.length < PASSWORD_MIN)
            return `Password must be at least ${PASSWORD_MIN} characters`;
         if (this.password.length > PASSWORD_MAX)
            return `Password must be at most ${PASSWORD_MAX} characters`;
         return null;
      },
   };
}
