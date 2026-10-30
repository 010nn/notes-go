import { restoreNotesKey, encryptNote, decryptNote } from "./pkg/crypto.js";

document.addEventListener("alpine:init", () => {
   Alpine.data("notesApp", notesApp);
});

function notesApp() {
   return {
      // === State ===
      notes: [],
      activeId: null,
      notesKey: null,
      syncStatus: "Synced",
      username: "",

      // Active note fileds
      activeTitle: "",
      activeContent: "",

      // Modals
      errorMessage: "",
      showErrorModal: false,
      showDeleteModal: false,

      // === Init ===
      async init() {
         this.notesKey = await restoreNotesKey();
         if (!this.notesKey) {
            await this.logout();
            return;
         }
         this.username = sessionStorage.getItem("username");
         await this.loadNotes();

         window.addEventListener("beforeunload", async () => {
            await this.forceSync();
         });
      },

      // === Computed ===
      get activeNote() {
         return this.notes.find((n) => n.id === this.activeId);
      },

      // === First load ===
      async loadNotes() {
         try {
            const res = await fetch("/api/notes");
            if (!res.ok) throw new Error();
            const data = await res.json();

            this.notes = await Promise.all(
               data.map(async (n) => {
                  const d = await decryptNote(n.content, this.notesKey);
                  return {
                     id: n.id,
                     title: d.title,
                     content: d.content,
                     version: n.version,
                     dirty: false,
                  };
               }),
            );
         } catch (e) {
            this.showError("Failed to load notes.");
         }
      },

      // === Create \ Delete ===
      createNote() {
         const id = crypto.randomUUID();
         const note = { id, title: "", content: "", version: 0, dirty: true };
         this.notes.push(note);
         this.switchNote(id);
         this.scheduleSync(note);
      },

      async deleteNote() {
         const note = this.activeNote;
         if (!note) return;

         try {
            const res = await fetch(`/api/notes/${note.id}`, {
               method: "DELETE",
            });
            if (res.ok) {
               this.notes = this.notes.filter((n) => n.id !== note.id);
               this.switchNoteAfterDeletion();
            } else {
               this.showError("Failed to delete note.");
            }
         } catch (e) {
            this.showError("Failed to delete note.");
         }
         this.showDeleteModal = false;
      },

      // === Switch notes ===
      switchNote(id) {
         this.activeId = id;
         const note = this.activeNote;
         this.activeTitle = note?.title || "";
         this.activeContent = note?.content || "";
      },

      switchNoteAfterDeletion() {
         if (this.notes.length === 0) {
            this.switchNote(null);
         } else {
            this.switchNote(this.notes[0].id);
         }
      },

      // === Input ===
      onInput() {
         const note = this.activeNote;
         if (!note) return;

         note.title = this.activeTitle;
         note.content = this.activeContent;
         note.dirty = true;

         this.scheduleSync(note);
      },

      // === Sync ===
      scheduleSync(note) {
         this.syncStatus = "Syncing...";

         if (!this._timers) this._timers = {};
         clearTimeout(this._timers[note.id]);

         this._timers[note.id] = setTimeout(async () => {
            await this.syncNote(note);
         }, 1500);
      },

      async syncNote(note) {
         if (!note.dirty) return;

         try {
            const res = await fetch(`/api/notes/${note.id}`, {
               method: "POST",
               headers: { "Content-Type": "application/json" },
               body: JSON.stringify({
                  content: await encryptNote(note, this.notesKey),
                  version: note.version,
               }),
            });

            if (res.ok) {
               const data = await res.json();
               note.version = data.version;
               note.dirty = false;
               this.updateSyncStatus();
            } else {
               this.syncStatus = "Sync Failed";
            }
         } catch (e) {
            this.syncStatus = "Sync Failed";
         }
      },

      async forceSync() {
         if (this._timers) {
            for (const timerId of Object.values(this._timers)) {
               clearTimeout(timerId);
            }
         }

         for (const note of this.notes) {
            if (!note.dirty) continue;

            note.dirty = false;
            const encryptedContent = await encryptNote(note, this.notesKey);

            const blob = new Blob(
               [
                  JSON.stringify({
                     content: encryptedContent,
                     version: note.version,
                  }),
               ],
               { type: "application/json" },
            );

            navigator.sendBeacon(`/api/notes/${note.id}`, blob);
         }
      },

      updateSyncStatus() {
         const anyDirty = this.notes.some((n) => n.dirty);
         this.syncStatus = anyDirty ? "Syncing..." : "Synced";
      },

      // === Error modal ===
      showError(msg) {
         this.errorMessage = msg;
         this.showErrorModal = true;
      },

      // === Auth ===
      async logout() {
         await this.forceSync();
         try {
            const response = await fetch(`/api/auth/logout`, {
               method: "POST",
            });

            if (response.ok) {
               this.notesKey = null;
               sessionStorage.removeItem("notesKey");
               sessionStorage.removeItem("username");
               window.location.href = "/login";
            }
         } catch (error) {
            this.showError("Failed to logout: " + error);
         }
      },
   };
}
