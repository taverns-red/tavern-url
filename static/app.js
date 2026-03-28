// Tavern URL — Dashboard interactions
// Uses event delegation to avoid inline onclick (CSP-safe).

document.addEventListener('DOMContentLoaded', function () {
  // ── Modal open ──
  document.addEventListener('click', function (e) {
    var btn = e.target.closest('[data-modal-open]');
    if (btn) {
      var modal = document.getElementById(btn.getAttribute('data-modal-open'));
      if (modal) modal.style.display = 'flex';
    }
  });

  // ── Modal close ──
  document.addEventListener('click', function (e) {
    var btn = e.target.closest('[data-modal-close]');
    if (btn) {
      var modal = document.getElementById(btn.getAttribute('data-modal-close'));
      if (modal) modal.style.display = 'none';
    }
  });

  // ── Close modal on backdrop click ──
  document.addEventListener('click', function (e) {
    if (e.target.id === 'create-modal' || e.target.id === 'edit-modal') {
      e.target.style.display = 'none';
    }
  });

  // ── Copy to clipboard ──
  document.addEventListener('click', function (e) {
    var btn = e.target.closest('[data-copy]');
    if (btn) {
      navigator.clipboard.writeText(btn.getAttribute('data-copy'));
      var orig = btn.textContent;
      btn.textContent = 'Copied!';
      setTimeout(function () { btn.textContent = orig; }, 2000);
    }
  });

  // ── Edit link modal ──
  document.addEventListener('click', function (e) {
    var btn = e.target.closest('[data-edit-link]');
    if (btn) {
      var modal = document.getElementById('edit-modal');
      var form = document.getElementById('edit-form');
      var urlInput = document.getElementById('edit-url');
      var slugInput = document.getElementById('edit-slug');
      if (modal && form && urlInput && slugInput) {
        urlInput.value = btn.getAttribute('data-edit-url') || '';
        slugInput.value = btn.getAttribute('data-edit-slug') || '';
        form.setAttribute('hx-put', '/api/v1/links/' + btn.getAttribute('data-edit-id'));
        // Re-process HTMX attributes on the updated form.
        if (window.htmx) htmx.process(form);
        modal.style.display = 'flex';
      }
    }
  });

  // ── Close modal and reset form after successful HTMX request ──
  document.body.addEventListener('htmx:afterRequest', function (e) {
    if (e.detail.successful) {
      var modal = e.detail.elt.closest('#create-modal') || e.detail.elt.closest('#edit-modal');
      if (modal) {
        modal.style.display = 'none';
        e.detail.elt.reset();
      }
    }
  });

  // ── Dark mode toggle ──
  var themeToggle = document.getElementById('theme-toggle');
  var html = document.documentElement;

  // Restore saved theme.
  var saved = localStorage.getItem('theme');
  if (saved) {
    html.setAttribute('data-theme', saved);
  } else if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
    html.setAttribute('data-theme', 'dark');
  }
  updateToggleIcon();

  if (themeToggle) {
    themeToggle.addEventListener('click', function () {
      var current = html.getAttribute('data-theme');
      var next = current === 'dark' ? 'light' : 'dark';
      html.setAttribute('data-theme', next);
      localStorage.setItem('theme', next);
      updateToggleIcon();
    });
  }

  function updateToggleIcon() {
    var toggles = document.querySelectorAll('#theme-toggle');
    var isDark = html.getAttribute('data-theme') === 'dark';
    toggles.forEach(function (t) {
      t.textContent = isDark ? '☀️' : '🌙';
      t.title = isDark ? 'Switch to light mode' : 'Switch to dark mode';
    });
  }
});
