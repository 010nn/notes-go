const NotesApp = function () {
   const DOM = {
      username: document.querySelector(".js-username"),
      logoutBtn: document.querySelector(".js-logout-btn"),
      syncStatus: document.querySelector(".js-sync-status"),
      noteDeleteBtn: document.querySelector(".js-note-delete-btn"),
      notePushBtn: document.querySelector(".js-note-push-btn"),
      noteCreateBtn: document.querySelector(".js-note-create-btn"),
      notesList: document.querySelector(".js-notes-list"),
      noteTitleInput: document.querySelector(".js-note-title-input"),
      noteContentInput: document.querySelector(".js-note-content-input"),
      errorModal: document.querySelector(".js-error-modal"),
      errorModalText: document.querySelector(".js-error-modal__text"),
      errorModalBtn: document.querySelector(".js-error-modal__btn"),
      deleteModal: document.querySelector(".js-delete-modal"),
      deleteModalText: document.querySelector(".js-delete-modal-text"),
      deleteModalCancelBtn: document.querySelector(".js-delete-modal-cancel"),
      deleteModalConfirmBtn: document.querySelector(".js-delete-modal-confirm"),
   };

   const state = {
      notes: new Map(),
      activeId: null,
   };

   class Note {
      constructor(id, title, content, version, dirty) {
         this.id = id;
         this.title = title;
         this.content = content;
         this.version = version;
         this.dirty = dirty;
         this.el = null;
         this.syncTimer = null;
      }

      scheduleSync() {
         updateSyncStatus();
         clearTimeout(this.syncTimer);
         this.syncTimer = setTimeout(() => {
            this.sync();
         }, 1500);
      }

      cancelSync() {
         clearTimeout(this.syncTimer);
         this.syncTimer = null;
      }

      async sync(keepalive = false) {
         if (!this.dirty) return true;

         this.cancelSync();

         try {
            const response = await fetch(`/api/notes/${this.id}`, {
               method: "POST",
               headers: { "Content-Type": "application/json" },
               body: JSON.stringify({
                  title: this.title,
                  content: this.content,
                  version: this.version,
               }),
               keepalive: keepalive,
            });

            if (response.ok) {
               const res = await response.json();
               this.version = res.version;
               this.dirty = false;
               updateSyncStatus();
               renderNote(this.id); // обновить маркер "•" в списке
               return true;
            }
         } catch (error) {
            if (state.activeId === this.id) {
               updateSyncStatus("Sync Failed");
            }
            showError("Sync failed. Please try again later.");
         }
         return false;
      }

      async delete() {
         try {
            const response = await fetch(`/api/notes/${this.id}`, {
               method: "DELETE",
            });

            if (response.ok) {
               this.cancelSync();
               state.notes.delete(this.id);
               this.el.parentNode.removeChild(this.el);
               renderNotes();
            } else {
               showError("Failed to delete note. Please try again later.");
            }
         } catch (error) {
            showError("Failed to delete note. Please try again later.");
         }
      }
   }

   function bindEvents() {
      DOM.logoutBtn.addEventListener("click", logout);

      // Sidebar btns
      DOM.noteDeleteBtn.addEventListener("click", () => {
         if (!state.activeId) return;
         DOM.deleteModal.hidden = false;
      });

      DOM.notePushBtn.addEventListener("click", async () => {
         const note = getNote(state.activeId);
         if (note) {
            await note.sync();
            renderNote(state.activeId);
         }
      });

      DOM.noteCreateBtn.addEventListener("click", () => {
         let id = createNote();
         renderNotes();
         switchNote(id);
      });

      // Note input
      DOM.noteTitleInput.addEventListener("input", updateRenderActiveNote);
      DOM.noteContentInput.addEventListener("input", updateRenderActiveNote);

      // Modal
      DOM.errorModalBtn.addEventListener("click", () => {
         DOM.errorModal.hidden = true;
      });

      DOM.deleteModalCancelBtn.addEventListener("click", () => {
         DOM.deleteModal.hidden = true;
      });

      DOM.deleteModalConfirmBtn.addEventListener("click", async () => {
         const note = getNote(state.activeId);
         if (!note) return;

         await note.delete();
         switchNoteAfterDeletion();
         DOM.deleteModal.hidden = true;
      });

      document.addEventListener("keydown", (event) => {
         if (event.key === "Escape") {
            if (!DOM.errorModal.hidden) DOM.errorModal.hidden = true;
            if (!DOM.deleteModal.hidden) DOM.deleteModal.hidden = true;
         }
      });

      // Sync before close tab
      window.addEventListener("beforeunload", () => {
         state.notes.forEach((note) => {
            if (note.dirty) {
               note.cancelSync();
               note.sync(true); // keepalive = true
            }
         });
      });
   }

   // === Transport ===

   async function logout() {
      try {
         const response = await fetch(`/api/auth/logout`, {
            method: "POST",
         });

         if (response.ok) {
            window.location.href = "/login";
         }
      } catch (error) {
         alert(error);
      }
   }

   async function loadNotes() {
      state.notes.clear();
      try {
         const response = await fetch(`/api/notes`);

         if (response.ok) {
            const res = await response.json();

            res.forEach((note) => {
               const newNote = new Note(
                  note.id,
                  note.title,
                  note.content,
                  note.version,
                  false,
               );
               state.notes.set(note.id, newNote);
            });
         } else {
            showError("Failed to load notes. Please try again later.");
         }
      } catch (error) {
         showError("Failed to load notes. Please try again later.");
      }
   }

   // === Service ===

   function getNote(id) {
      return state.notes.get(id);
   }

   function createNote() {
      let note = new Note(crypto.randomUUID(), "", "", 0, true);
      state.notes.set(note.id, note);
      return note.id;
   }

   function updateRenderActiveNote() {
      const note = getNote(state.activeId);
      if (!note) return;

      note.title = DOM.noteTitleInput.value;
      note.content = DOM.noteContentInput.value;
      note.dirty = true;

      renderNote(state.activeId);
      note.scheduleSync(); // Запускаем таймер именно у этой заметки
   }

   function switchNote(id) {
      state.activeId = id;
      renderNoteContent(id);
      renderNotesActiveStatus();
   }

   function switchNoteAfterDeletion() {
      if (state.notes.size === 0) {
         switchNote(null);
         return;
      }

      state.notes.forEach((_, id) => {
         if (state.activeId != id) {
            switchNote(id);
            return;
         }
      });
   }

   // === Render ===

   function renderNotes() {
      if (state.notes.size === 0) {
         DOM.notesList.innerHTML = "";
         return;
      }

      state.notes.forEach((note, id) => {
         if (note.el === null) {
            const newNote = document.createElement("a");
            newNote.className = "js-note-item note-item";
            newNote.addEventListener("click", () => switchNote(id));
            DOM.notesList.appendChild(newNote);
            note.el = newNote;
         }

         let title = note.title || "untitled";
         if (note.dirty) title += " •";
         if (note.el.textContent !== title) {
            note.el.textContent = title;
         }
      });

      renderNotesActiveStatus();
   }

   function renderNote(id) {
      const note = getNote(id);
      if (!note || !note.el) return;

      let title = note.title || "untitled";
      if (note.dirty) title += " •";

      if (note.el.textContent !== title) {
         note.el.textContent = title;
      }
   }

   function renderNotesActiveStatus() {
      state.notes.forEach((note) => {
         if (!note.el) return;
         if (note.id === state.activeId) {
            note.el.classList.add("active");
         } else {
            note.el.classList.remove("active");
         }
      });
   }

   function renderNoteContent(id) {
      const note = getNote(id);
      if (!note) {
         DOM.noteTitleInput.value = "";
         DOM.noteContentInput.value = "";
         return;
      }

      DOM.noteTitleInput.value = note.title;
      DOM.noteContentInput.value = note.content;
   }

   function updateSyncStatus(message = "") {
      if (message != "") {
         DOM.syncStatus.textContent = message;
         return;
      }

      let isThereDirty = false;
      state.notes.forEach((note) => {
         if (note.dirty) {
            isThereDirty = true;
            return;
         }
      });

      DOM.syncStatus.textContent = isThereDirty ? "Syncing..." : "Synced";
   }

   // === Modal for Confirmation ===

   function showError(message) {
      DOM.errorModalText.textContent = message;
      DOM.errorModal.hidden = false;
      DOM.errorModalBtn.focus();
   }

   // === Return ===

   return {
      init: () => {
         bindEvents();
         loadNotes().then(renderNotes);
      },
   };
};

NotesApp().init();
