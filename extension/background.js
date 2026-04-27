const DEFAULTS = {
  serverUrl: 'http://localhost:8998',
  interceptPrefix: 'go'
};

async function getSettings() {
  const settings = await chrome.storage.sync.get(DEFAULTS);
  settings.serverUrl = settings.serverUrl.replace(/\/+$/, '');
  return settings;
}

function escapeXml(str) {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

// --- Omnibox ---

let debounceTimer = null;

chrome.omnibox.onInputChanged.addListener((text, suggest) => {
  clearTimeout(debounceTimer);

  chrome.omnibox.setDefaultSuggestion({
    description: `Navigate to <match>go/${escapeXml(text)}</match>`
  });

  if (!text.trim()) return;

  debounceTimer = setTimeout(async () => {
    try {
      const settings = await getSettings();
      const resp = await fetch(`${settings.serverUrl}/api/v1/search?qs=${encodeURIComponent(text.trim())}`);
      if (!resp.ok) return;

      const links = await resp.json();
      const suggestions = links.map(link => {
        const isWildcard = !!link.matched_query;
        const displayKeyword = isWildcard ? link.matched_query : link.keyword;
        const navigateKeyword = isWildcard ? link.matched_query : link.keyword;
        const patternNote = isWildcard ? ` <dim>(via ${escapeXml(link.keyword)})</dim>` : '';

        return {
          content: `${settings.serverUrl}/${navigateKeyword}`,
          description: `<match>go/${escapeXml(displayKeyword)}</match>${patternNote} <dim>- ${escapeXml(link.description || link.destination)}</dim>`
        };
      });
      suggest(suggestions);
    } catch (_) {
      // Server unreachable — silently fail
    }
  }, 200);
});

chrome.omnibox.onInputEntered.addListener(async (text, disposition) => {
  const settings = await getSettings();
  let url;

  if (text.startsWith('http://') || text.startsWith('https://')) {
    url = text;
  } else {
    url = `${settings.serverUrl}/${text}`;
  }

  switch (disposition) {
    case 'currentTab':
      chrome.tabs.update({ url });
      break;
    case 'newForegroundTab':
      chrome.tabs.create({ url });
      break;
    case 'newBackgroundTab':
      chrome.tabs.create({ url, active: false });
      break;
  }
});

// --- URL Interception via declarativeNetRequest ---

const RULE_ID = 1;

async function updateRedirectRules() {
  const settings = await getSettings();
  const prefix = settings.interceptPrefix;
  const serverUrl = settings.serverUrl;

  await chrome.declarativeNetRequest.updateDynamicRules({
    removeRuleIds: [RULE_ID],
    addRules: [
      {
        id: RULE_ID,
        priority: 1,
        action: {
          type: 'redirect',
          redirect: {
            regexSubstitution: `${serverUrl}\\1`
          }
        },
        condition: {
          regexFilter: `^https?://${prefix}(/.*)$`,
          resourceTypes: ['main_frame']
        }
      }
    ]
  });
}

updateRedirectRules();

chrome.storage.onChanged.addListener((changes, area) => {
  if (area === 'sync' && (changes.serverUrl || changes.interceptPrefix)) {
    updateRedirectRules();
  }
});
