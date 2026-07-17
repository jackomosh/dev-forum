document.addEventListener("DOMContentLoaded", () => {
  // --- Responsive Navigation Operations ---
  const mobileMenuBtn = document.getElementById("mobile-menu-btn");
  const mobileNav = document.getElementById("mobile-nav");

  if (mobileMenuBtn && mobileNav) {
    mobileMenuBtn.addEventListener("click", () => {
      const isExpanded = mobileMenuBtn.getAttribute("aria-expanded") === "true";
      mobileMenuBtn.setAttribute("aria-expanded", !isExpanded);
      mobileNav.setAttribute("aria-hidden", isExpanded);
      mobileNav.classList.toggle("open");
      mobileMenuBtn.classList.toggle("open");
    });

    document.addEventListener("click", (event) => {
      if (!mobileMenuBtn.contains(event.target) && !mobileNav.contains(event.target)) {
        mobileMenuBtn.setAttribute("aria-expanded", "false");
        mobileNav.setAttribute("aria-hidden", "true");
        mobileNav.classList.remove("open");
        mobileMenuBtn.classList.remove("open");
      }
    });
  }

  // --- Rich Text WYSIWYG Form Binder Sync ---
  const publishForm = document.getElementById("publish-form");
  const wysiwygEditor = document.getElementById("editor-wysiwyg");
  const hiddenBodyInput = document.getElementById("post-body-input");

  if (publishForm && wysiwygEditor && hiddenBodyInput) {
    publishForm.addEventListener("submit", (e) => {
      const contents = wysiwygEditor.innerHTML.trim();
      if (contents === "" || contents === "<br>") {
        e.preventDefault();
        alert("Please write standard body content into the thread workspace before publishing.");
        return;
      }
      hiddenBodyInput.value = contents;
    });
  }
});

// --- Workspace Toggle Suite (Edit vs Live Preview Sandbox Views) ---
function switchMode(mode) {
  const editPanel = document.getElementById("workspace-edit-panel");
  const previewPanel = document.getElementById("workspace-preview-panel");
  const btnEdit = document.getElementById("btn-mode-edit");
  const btnPreview = document.getElementById("btn-mode-preview");
  const wysiwygEditor = document.getElementById("editor-wysiwyg");

  if (!editPanel || !previewPanel) return;

  if (mode === 'preview') {
    const titleVal = document.getElementById("post-title").value.trim() || "Untitled Structural Thread Document";
    const bodyVal = (wysiwygEditor && wysiwygEditor.innerHTML.trim() !== "") ? wysiwygEditor.innerHTML : "<i>No description layer provided.</i>";
    
    document.getElementById("preview-title").textContent = titleVal;
    document.getElementById("preview-body").innerHTML = bodyVal;

    const badgeWrapper = document.getElementById("preview-badges");
    if (badgeWrapper) {
      badgeWrapper.innerHTML = "";
      const checkedBoxes = document.querySelectorAll("input[name='categories']:checked");
      checkedBoxes.forEach(box => {
        const catName = box.getAttribute("data-category-name") || "Subforum";
        const badge = document.createElement("span");
        badge.className = "badge-subforum";
        badge.textContent = "#" + catName;
        badgeWrapper.appendChild(badge);
      });
    }

    editPanel.style.display = "none";
    previewPanel.style.display = "block";
    
    if (btnPreview) {
      btnPreview.className = "btn-comment-submit-round";
    }
    if (btnEdit) {
      btnEdit.className = "vote-btn";
      btnEdit.style.borderRadius = "30px";
    }
  } else {
    editPanel.style.display = "block";
    previewPanel.style.display = "none";
    
    if (btnEdit) {
      btnEdit.className = "btn-comment-submit-round";
    }
    if (btnPreview) {
      btnPreview.className = "vote-btn";
      btnPreview.style.borderRadius = "30px";
    }
  }
}

// --- Rich Text Context Execution Commands ---
function formatDoc(cmd, value = null) {
  const editor = document.getElementById("editor-wysiwyg");
  if (editor) {
    editor.focus();
    document.execCommand(cmd, false, value);
  }
}

function formatSize(size) {
  const editor = document.getElementById("editor-wysiwyg");
  if (editor) {
    editor.focus();
    // Translate numeric choices securely to standard font scaling factors
    document.execCommand('fontSize', false, size);
  }
}

function insertLink() {
  const url = prompt("Enter targeted destination hyperlink URL:");
  if (url) formatDoc("createLink", url);
}

function insertImage() {
  const url = prompt("Enter explicit static resource image URL link path:");
  if (url) formatDoc("insertImage", url);
}

function toggleEmojiPicker() {
  const panel = document.getElementById("emoji-picker-panel");
  if (panel) {
    if (panel.classList.contains("emoji-picker-hidden")) {
      panel.classList.remove("emoji-picker-hidden");
      panel.classList.add("emoji-picker-active");
    } else {
      panel.classList.remove("emoji-picker-active");
      panel.classList.add("emoji-picker-hidden");
    }
  }
}

// Fixed parameter ingestion loop to match manual string bindings
function insertEmoji(emoji) {
  formatDoc("insertHTML", emoji);
  toggleEmojiPicker();
}

// Generates dynamic raw structural inline micro-elements safely
function insertHashtag() {
  const tag = prompt("Specify custom index target tag term without standard hash characters:");
  if (tag) formatDoc("insertHTML", ` <span class='badge-subforum'>#${tag.trim()}</span> `);
}

function simulateAttachment() {
  alert("File attachment processing framework is available for authenticated high-karma authors.");
}