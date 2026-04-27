const DEFAULTS = {
  serverUrl: 'http://localhost:8998',
  interceptPrefix: 'go'
};

const searchInput = document.getElementById('search');
const resultsEl = document.getElementById('results');
const placeholderEl = document.getElementById('placeholder');
const dashboardLink = document.getElementById('dashboard');
const settingsLink = document.getElementById('settings');

let debounceTimer = null;
let serverUrl = DEFAULTS.serverUrl;
let prefix = DEFAULTS.interceptPrefix;

chrome.storage.sync.get(DEFAULTS, (settings) => {
  serverUrl = settings.serverUrl.replace(/\/+$/, '');
  prefix = settings.interceptPrefix;
});

dashboardLink.addEventListener('click', () => {
  chrome.tabs.create({ url: `${serverUrl}/` });
});

settingsLink.addEventListener('click', () => {
  chrome.runtime.openOptionsPage();
});

searchInput.addEventListener('input', () => {
  clearTimeout(debounceTimer);
  const query = searchInput.value.trim();

  if (!query) {
    resultsEl.innerHTML = '';
    resultsEl.appendChild(placeholderEl);
    placeholderEl.style.display = '';
    return;
  }

  debounceTimer = setTimeout(async () => {
    try {
      const resp = await fetch(`${serverUrl}/api/v1/search?qs=${encodeURIComponent(query)}`);
      if (!resp.ok) {
        showEmpty('Server error');
        return;
      }
      const links = await resp.json();
      renderResults(links);
    } catch (_) {
      showEmpty('Cannot reach server');
    }
  }, 250);
});

function showEmpty(msg) {
  resultsEl.innerHTML = `<div class="empty">${msg}</div>`;
}

function renderResults(links) {
  if (!links || links.length === 0) {
    showEmpty('No results found');
    return;
  }

  resultsEl.innerHTML = '';
  for (const link of links) {
    const isWildcard = !!link.matched_query;
    const displayKeyword = isWildcard ? link.matched_query : link.keyword;
    const navigateKeyword = isWildcard ? link.matched_query : link.keyword;

    const item = document.createElement('a');
    item.className = 'result-item';
    item.href = '#';
    item.innerHTML = `
      <div class="result-keyword">${prefix}/${esc(displayKeyword)}${isWildcard ? ` <span class="wildcard-badge">via ${esc(link.keyword)}</span>` : ''}</div>
      <div class="result-desc">${esc(link.description || link.destination)}</div>
    `;
    item.addEventListener('click', (e) => {
      e.preventDefault();
      chrome.tabs.create({ url: `${serverUrl}/${navigateKeyword}` });
    });
    resultsEl.appendChild(item);
  }
}

function esc(str) {
  const el = document.createElement('span');
  el.textContent = str;
  return el.innerHTML;
}
