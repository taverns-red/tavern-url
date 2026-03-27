// Tavern URL — Dashboard interactions
// Uses event delegation to avoid inline onclick (CSP-safe).

document.addEventListener('DOMContentLoaded', function () {
  // Open create-link modal
  document.addEventListener('click', function (e) {
    var btn = e.target.closest('[data-modal-open]');
    if (btn) {
      var modal = document.getElementById(btn.getAttribute('data-modal-open'));
      if (modal) modal.style.display = 'flex';
    }
  });

  // Close create-link modal
  document.addEventListener('click', function (e) {
    var btn = e.target.closest('[data-modal-close]');
    if (btn) {
      var modal = document.getElementById(btn.getAttribute('data-modal-close'));
      if (modal) modal.style.display = 'none';
    }
  });

  // Close modal on backdrop click
  document.addEventListener('click', function (e) {
    if (e.target.id === 'create-modal') {
      e.target.style.display = 'none';
    }
  });

  // Copy to clipboard
  document.addEventListener('click', function (e) {
    var btn = e.target.closest('[data-copy]');
    if (btn) {
      navigator.clipboard.writeText(btn.getAttribute('data-copy'));
      btn.textContent = 'Copied!';
      setTimeout(function () { btn.textContent = 'Copy'; }, 2000);
    }
  });

  // Close modal and reset form after successful HTMX link creation
  document.body.addEventListener('htmx:afterRequest', function (e) {
    if (e.detail.successful && e.detail.elt.closest('#create-modal')) {
      var modal = document.getElementById('create-modal');
      if (modal) modal.style.display = 'none';
      e.detail.elt.reset();
    }
  });
});
