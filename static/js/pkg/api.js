export class ApiError extends Error {
   constructor(message, status) {
      super(message);
      this.name = "ApiError";
      this.status = status;
   }
}

async function postJSON(path, body) {
   let response;
   try {
      response = await fetch(path, {
         method: "POST",
         headers: { "Content-Type": "application/json" },
         body: JSON.stringify(body),
      });
   } catch (e) {
      throw new ApiError("Network error: " + e.message, 0);
   }

   if (!response.ok) {
      const text = await response.text().catch(() => "");
      throw new ApiError(
         `${response.status} ${response.statusText}${text ? ": " + text : ""}`,
         response.status,
      );
   }

   try {
      return await response.json();
   } catch {
      throw new ApiError("Server returned invalid JSON", response.status);
   }
}

export const api = {
   login: (username, password) =>
      postJSON("/api/auth/login", { username, password }),
   register: (username, password) =>
      postJSON("/api/auth/register", { username, password }),
};
