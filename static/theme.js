// Theme toggle — persists preference in localStorage
(function() {
  const stored = localStorage.getItem('theme');
  if (stored) {
    document.documentElement.setAttribute('data-theme', stored);
  }

  window.toggleTheme = function() {
    const current = document.documentElement.getAttribute('data-theme');
    const next = current === 'dark' ? 'light' : 'dark';
    document.documentElement.setAttribute('data-theme', next);
    localStorage.setItem('theme', next);

    // Update toggle button text
    const btn = document.querySelector('.theme-toggle');
    if (btn) btn.textContent = next === 'dark' ? '☀️' : '🌙';
  };
})();
