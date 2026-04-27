const DEFAULTS = {
  serverUrl: 'http://localhost:8998',
  interceptPrefix: 'go'
};

const serverUrlInput = document.getElementById('serverUrl');
const interceptPrefixInput = document.getElementById('interceptPrefix');
const saveBtn = document.getElementById('save');
const statusEl = document.getElementById('status');

function showStatus(msg, isError) {
  statusEl.textContent = msg;
  statusEl.className = 'status visible' + (isError ? ' error' : '');
  setTimeout(() => { statusEl.className = 'status'; }, 2000);
}

chrome.storage.sync.get(DEFAULTS, (settings) => {
  serverUrlInput.value = settings.serverUrl;
  interceptPrefixInput.value = settings.interceptPrefix;
});

saveBtn.addEventListener('click', () => {
  let serverUrl = serverUrlInput.value.trim().replace(/\/+$/, '');
  const interceptPrefix = interceptPrefixInput.value.trim().toLowerCase();

  if (!serverUrl.match(/^https?:\/\/.+/)) {
    showStatus('Server URL must start with http:// or https://', true);
    return;
  }

  if (!interceptPrefix.match(/^[a-zA-Z0-9]+$/)) {
    showStatus('Prefix must be alphanumeric only', true);
    return;
  }

  chrome.storage.sync.set({ serverUrl, interceptPrefix }, () => {
    showStatus('Saved');
  });
});
