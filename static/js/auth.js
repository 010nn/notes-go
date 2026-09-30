
const DOM = {
   form: document.querySelector('.js-auth-form'),
   btnLogin: document.querySelector('.js-auth-form-login'),
   btnRegister: document.querySelector('.js-auth-form-register'),
   username: document.querySelector('.js-auth-form-username'),
   password: document.querySelector('.js-auth-form-password')
}

function bindEvents() {

   DOM.btnLogin.addEventListener('click', (event) => {
      login(event)
   })
   DOM.btnRegister.addEventListener('click', (event) => {
      register(event)
   })
}

async function login(event) {
   event.preventDefault();

   const data = {
      username: DOM.username.value,
      password: DOM.password.value
   }

   console.log("Login", data)

   if (data.username === '' || data.password === '') {
      throw new Error("Invalid username or password");
   }

   const response = await fetch('/api/login', {
      method: 'POST',
      body: JSON.stringify(data)
   });

   if (!response.ok) {
      throw new Error("Login error");
   }

   window.location.href("/")
}

async function register(event) {
   event.preventDefault();

   const data = {
      username: DOM.username.value,
      password: DOM.password.value
   }

   if (data.username === '' || data.password === '') {
      throw new Error("Invalid username or password");
   }

   const response = await fetch('/api/register', {
      method: 'POST',
      body: JSON.stringify(data)
   });

   if (!response.ok) {
      throw new Error(response.statusText);
   }

   alert("Successfully")
}


bindEvents();
