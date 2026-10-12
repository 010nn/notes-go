const NotesApp = (function () {

   const DOM = {
      username: document.querySelector('.js-username'),
      logoutBtn: document.querySelector('.js-logout-btn'),
      syncStatus: document.querySelector('.js-sync-status'),
      noteDeleteBtn: document.querySelector('.js-note-delete-btn'),
      notePushBtn: document.querySelector('.js-note-push-btn'),
      noteCreateBtn: document.querySelector('.js-note-create-btn'),
      notesList: document.querySelector('.js-notes-list'),
      noteTitleInput: document.querySelector('.js-note-title-input'),
      noteContentInput: document.querySelector('.js-note-content-input'),
   }

   class Note {
      constructor(id, title, content, version, dirty) {
         this.id = id
         this.title = title
         this.content = content
         this.version = version
         this.dirty = dirty
         this.el = null
      }
   }

   const state = {
      notes: new Map(),
      activeId: null
   }

   function bindEvents() {

      DOM.logoutBtn.addEventListener('click', logout)
      DOM.notePushBtn.addEventListener('click', async () => {
         if (await pushNote(state.activeId)) {
            renderNote(state.activeId)
         }
      })
      DOM.noteCreateBtn.addEventListener('click', () => {
         let id = createNote()
         renderNotes()
         switchNote(id)
      })
      DOM.noteTitleInput.addEventListener('input', updateRenderActiveNote)
      DOM.noteContentInput.addEventListener('input', updateRenderActiveNote)
   }

   // === Transport ===

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

   async function loadNotes() {
      state.notes.clear()
      try {
         const response = await fetch(`/api/notes`)

         if (response.ok) {
            res = await response.json()

            res.forEach(note => {
               newNote = new Note(
                  note.id, note.title, note.content, note.version, false
               )
               state.notes.set(note.id, newNote)
               console.log("Load: ", note, newNote)
            })
         }
      } catch (error) {
         alert(error)
      }
   }

   async function pushNote(id) {
      const note = getNote(id)
      if (!note) return
      console.log("Push note: ", note)
      try {
         const response = await fetch(`/api/notes/${note.id}`, {
            method: 'POST',
            headers: {
               'Content-Type': 'application/json'
            },
            body: JSON.stringify({
               title: note.title,
               content: note.content,
               version: note.version,
            })
         })

         if (response.ok) {
            res = await response.json()
            note.dirty = false
            note.version = res.version
            console.log("Push result:", res)
            return true
         }
      } catch (error) {
         alert(error)
      }
      return false
   }

   // === Service ===

   function getNote(id) {
      return state.notes.get(id)
   }

   function createNote() {
      let note = new Note(
         crypto.randomUUID(), '', '', 0, true
      )
      state.notes.set(note.id, note)
      return note.id
   }

   function updateRenderActiveNote() {
      updateNoteData(
         state.activeId,
         DOM.noteTitleInput.value,
         DOM.noteContentInput.value,
         true
      )
      renderNote(state.activeId)
   }

   function updateNoteData(id, title, content, dirty) {
      const note = getNote(id)
      if (!note) return

      note.title = title
      note.content = content
      note.dirty = dirty
   }

   function switchNote(id) {
      state.activeId = id
      renderNoteContent(id)
      renderNotesActiveStatus()
   }

   // === Render ===

   function renderNotes() {
      if (state.notes.size === 0) {
         DOM.notesList.innerHTML = `Empty`
         return
      }

      state.notes.forEach((note, id) => {
         if (note.el === null) {
            const newNote = document.createElement('a')
            newNote.className = 'js-note-item note-item'
            newNote.textContent = note.title
            // newNote.dataset.id = id
            newNote.addEventListener('click', () => {
               switchNote(id)
            })
            DOM.notesList.appendChild(newNote)
            note.el = newNote
         } else {
            if (note.el.textContent !== note.title) {
               note.el.textContent = note.title
            }
         }
         if (note.dirty) {
            note.el.textContent += " •"
         }
      })

      renderNotesActiveStatus()
   }

   function renderNote(id) {
      const note = getNote(id)
      if (!note) return

      note.el.textContent = note.title
      if (note.dirty) {
         note.el.textContent += " •"
      }
   }

   function renderNotesActiveStatus() {
      state.notes.forEach(note => {
         if (note.id === state.activeId) {
            note.el.classList.add('active')
         } else {
            note.el.classList.remove('active')
         }
      })
   }

   function renderNoteContent(id) {
      const note = getNote(id)
      if (!note) return

      DOM.noteTitleInput.value = note.title
      DOM.noteContentInput.value = note.content
   }

   // === Return ===

   return {
      init: () => {
         bindEvents()
         loadNotes().then(renderNotes)
      }
   }

})

NotesApp().init()
