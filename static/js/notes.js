const NotesApp = (function () {

   const DOM = {
      username: document.querySelector('.js-username'),
      logoutBtn: document.querySelector('.js-logout-btn'),
      syncStatus: document.querySelector('.js-sync-status'),
      noteDeleteBtn: document.querySelector('.js-note-delete-btn'),
      noteAddBtn: document.querySelector('.js-note-add-btn'),
      notesList: document.querySelector('.js-notes-list'),
      noteTitleInput: document.querySelector('.js-note-title-input'),
      noteContentInput: document.querySelector('.js-note-content-input'),
   }


   function bindEvents() {

      DOM.logoutBtn.addEventListener('click', () => {
         logout();
      })

   }


   async function logout() {

      try {
         const response = await fetch(`/api/auth/logout`, {
            method: 'POST'
         })

         if (response.ok) {
            window.location.href = "/login";
         }

      } catch (error) {
         alert(error)
      }

   }

   return {
      init: () => {
         bindEvents()
      }
   }

})

NotesApp().init()
