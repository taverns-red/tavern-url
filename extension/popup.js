// Tavern URL Browser Extension
document.addEventListener('DOMContentLoaded', async () => {
  const settings = await chrome.storage.local.get(['apiKey', 'baseUrl']);

  if (settings.apiKey && settings.baseUrl) {
    document.getElementById('settings').style.display = 'none';
    document.getElementById('shorten').style.display = 'block';

    // Get current tab URL
    const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
    document.getElementById('pageUrl').value = tab.url;
  }

  // Save settings
  document.getElementById('saveBtn').addEventListener('click', async () => {
    const apiKey = document.getElementById('apiKey').value.trim();
    const baseUrl = document.getElementById('baseUrl').value.trim();

    if (!apiKey || !baseUrl) {
      document.getElementById('status').textContent = 'Both fields required';
      return;
    }

    await chrome.storage.local.set({ apiKey, baseUrl });
    document.getElementById('status').textContent = 'Saved! Reopen popup to shorten.';
  });

  // Shorten URL
  document.getElementById('shortenBtn').addEventListener('click', async () => {
    const url = document.getElementById('pageUrl').value;
    const { apiKey, baseUrl } = await chrome.storage.local.get(['apiKey', 'baseUrl']);
    const status = document.getElementById('status');

    status.textContent = 'Shortening...';

    try {
      const res = await fetch(`${baseUrl}/api/v1/links`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${apiKey}`
        },
        body: JSON.stringify({ url })
      });

      if (!res.ok) throw new Error(`HTTP ${res.status}`);

      const data = await res.json();
      const result = document.getElementById('result');
      const shortUrl = document.getElementById('shortUrl');

      shortUrl.href = data.short_url;
      shortUrl.textContent = data.short_url;
      result.classList.add('show');

      // Copy to clipboard
      await navigator.clipboard.writeText(data.short_url);
      status.textContent = 'Copied to clipboard!';
    } catch (err) {
      status.textContent = `Error: ${err.message}`;
    }
  });
});
