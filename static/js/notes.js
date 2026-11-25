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
      draggedId: null,
      orderChanged: false,

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

      // === Create \ Delete \ Switch ===
      createNote() {
         const id = crypto.randomUUID();
         const note = { id, title: "", content: "", version: 0, dirty: true };
         this.notes.push(note);
         this.switchNote(id);
         this.orderChanged = true;
      },

      async deleteNote() {
         const note = this.activeNote;
         if (!note) return;

         try {
            const res = await fetch(`/api/notes/${note.id}`, {
               method: "DELETE",
            });
            if (res.ok) {
               const prevNoteId =
                  this.notes.findIndex((n) => n.id === note.id) - 1;
               this.notes = this.notes.filter((n) => n.id !== note.id);

               this.switchNote(this.notes[prevNoteId]?.id);
            } else {
               this.showError("Failed to delete note.");
            }
         } catch (e) {
            this.showError("Failed to delete note.");
         }
         this.showDeleteModal = false;
         this.orderChanged = true;
      },

      switchNote(id) {
         this.activeId = id;
         const note = this.activeNote;
         this.activeTitle = note?.title || "";
         this.activeContent = note?.content || "";
         this.updateHighlight();
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
            } else if (res.status === 409) {
               this.showError(
                  "Sync conflict: this note was changed elsewhere. Reload to overwrite.",
               );
            } else {
               this.showError(
                  "Sync failed: " + res.status + " " + res.statusText,
               );
            }
         } catch (e) {
            this.showError("Sync failed: " + (e?.message ?? e));
         }
      },

      async retrySync() {
         for (const note of this.notes) {
            if (note.dirty) {
               await this.syncNote(note);
            }
         }
         this.updateSyncStatus();
      },

      async syncOrder() {
         if (!this.orderChanged) return;
         try {
            const res = await fetch("/api/notes/reorder", {
               method: "POST",
               headers: { "Content-Type": "application/json" },
               body: JSON.stringify({
                  order: this.notes.map((note) => note.id),
               }),
               keepalive,
            });
            if (!res.ok) {
               throw new Error();
            } else {
               this.orderChanged = false;
            }
         } catch (e) {
            this.showError("Failed to save order.");
            // await this.loadNotes();
         }
      },

      async forceSync() {
         // Save order
         if (this.orderChanged) {
            navigator.sendBeacon(
               `/api/notes/reorder`,
               new Blob(
                  [
                     JSON.stringify({
                        order: this.notes.map((note) => note.id),
                     }),
                  ],
                  { type: "application/json" },
               ),
            );
         }

         // Clear notes timeout
         if (this._timers) {
            for (const timerId of Object.values(this._timers)) {
               clearTimeout(timerId);
            }
         }

         // Save changed notes
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

      // === Drag & drop ===
      onDragStart(id, e) {
         this.draggedId = id;
         e.dataTransfer.effectAllowed = "move";
         e.dataTransfer.setData("text/plain", id);
      },
      onDragOver(e) {
         e.preventDefault();
      },
      onDrop(targetId, e) {
         e.preventDefault();
         const fromId = this.draggedId;
         this.draggedId = null;
         if (!fromId || fromId === targetId) return;

         const fromIdx = this.notes.findIndex((note) => note.id === fromId);
         const toIdx = this.notes.findIndex((note) => note.id === targetId);
         if (fromIdx === -1 || toIdx === -1) return;

         const [moved] = this.notes.splice(fromIdx, 1);
         this.notes.splice(toIdx, 0, moved);
         this.orderChanged = true;
         this.updateHighlight();
      },
      onDragEnd() {
         this.draggedId = null;
      },

      // === Highlight ===
      updateHighlight() {
         this.$nextTick(() => {
            const list = document.querySelector(".sidebar__list");
            if (!list) return;
            const active = list.querySelector(".note-item.active");
            list.style.setProperty(
               "--highlight-y",
               active ? `${active.offsetTop}px` : "-100px",
            );

            list.style.setProperty(
               "--highlight-x",
               active ? `${active.offsetLeft}px` : "-100px",
            );
            list.style.setProperty(
               "--highlight-w",
               active ? `${active.offsetWidth}px` : "0px",
            );
         });
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
