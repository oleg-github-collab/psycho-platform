import { parseMarkdown, createMarkdownToolbar } from './markdown.js';
import { showEmojiPicker } from './emoji-picker.js';

// API Configuration
const API_URL = window.location.hostname === 'localhost'
  ? 'http://localhost:8080/api'
  : '/api';

const WS_URL = window.location.hostname === 'localhost'
  ? 'ws://localhost:8080/api/ws'
  : `wss://${window.location.host}/api/ws`;

// State management
const state = {
  user: null,
  token: localStorage.getItem('token'),
  currentView: 'login',
  currentTopic: null,
  currentConversation: null,
  topics: [],
  messages: [],
  groups: [],
  sessions: [],
  appointments: [],
  conversations: [],
  users: [],
  typingUsers: new Set(),
  ws: null,
};

// API helpers
async function apiCall(endpoint, options = {}) {
  const headers = {
    'Content-Type': 'application/json',
    ...(state.token && { Authorization: `Bearer ${state.token}` }),
  };

  const response = await fetch(`${API_URL}${endpoint}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Network error' }));
    throw new Error(error.error || 'Request failed');
  }

  return response.json();
}

// WebSocket connection
function connectWebSocket() {
  if (!state.token) return;

  state.ws = new WebSocket(`${WS_URL}?token=${state.token}`);

  state.ws.onopen = () => {
    console.log('WebSocket connected');
    // Set online status
    apiCall('/status/online?online=true', { method: 'POST' });

    // Join current room
    if (state.currentTopic) {
      joinRoom('topic_' + state.currentTopic);
    }
  };

  state.ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    handleWebSocketMessage(data);
  };

  state.ws.onerror = (error) => {
    console.error('WebSocket error:', error);
  };

  state.ws.onclose = () => {
    console.log('WebSocket disconnected');
    apiCall('/status/online?online=false', { method: 'POST' }).catch(() => {});
    setTimeout(connectWebSocket, 3000);
  };

  // Set offline on page unload
  window.addEventListener('beforeunload', () => {
    apiCall('/status/online?online=false', { method: 'POST' }).catch(() => {});
  });
}

function handleWebSocketMessage(data) {
  if (data.type === 'new_message') {
    state.messages.unshift(data.payload);
    render();
  } else if (data.type === 'new_dm') {
    if (state.currentView === 'conversations') {
      fetchConversations();
    }
  } else if (data.type === 'typing') {
    if (data.payload.is_typing) {
      state.typingUsers.add(data.payload.user_id);
    } else {
      state.typingUsers.delete(data.payload.user_id);
    }
    updateTypingIndicator();
  }
}

function joinRoom(roomID) {
  if (state.ws && state.ws.readyState === WebSocket.OPEN) {
    state.ws.send(JSON.stringify({
      type: 'join_room',
      room: roomID,
    }));
  }
}

function updateTypingIndicator() {
  const indicator = document.getElementById('typing-indicator');
  if (!indicator) return;

  if (state.typingUsers.size > 0) {
    indicator.textContent = `${state.typingUsers.size} ${state.typingUsers.size === 1 ? 'користувач друкує' : 'користувачів друкують'}...`;
    indicator.style.display = 'block';
  } else {
    indicator.style.display = 'none';
  }
}

let typingTimer;
function handleTyping(roomID) {
  clearTimeout(typingTimer);

  // Send typing start
  apiCall(`/messages/typing/start?room=${roomID}`, { method: 'POST' }).catch(() => {});

  // Auto-stop after 3 seconds
  typingTimer = setTimeout(() => {
    apiCall(`/messages/typing/stop?room=${roomID}`, { method: 'POST' }).catch(() => {});
  }, 3000);
}

// Auth functions
async function register(username, password, displayName) {
  const data = await apiCall('/auth/register', {
    method: 'POST',
    body: JSON.stringify({ username, password, display_name: displayName }),
  });

  state.token = data.token;
  state.user = data.user;
  localStorage.setItem('token', data.token);
  connectWebSocket();
  state.currentView = 'topics';
  render();
}

async function login(username, password) {
  const data = await apiCall('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });

  state.token = data.token;
  state.user = data.user;
  localStorage.setItem('token', data.token);
  connectWebSocket();
  state.currentView = 'topics';
  render();
}

function logout() {
  apiCall('/status/online?online=false', { method: 'POST' }).catch(() => {});
  state.token = null;
  state.user = null;
  state.currentView = 'login';
  localStorage.removeItem('token');
  if (state.ws) state.ws.close();
  render();
}

// Data fetching
async function fetchTopics() {
  state.topics = await apiCall('/topics');
  render();
}

async function fetchMessages(topicId, groupId) {
  const query = topicId ? `?topic_id=${topicId}` : groupId ? `?group_id=${groupId}` : '';
  state.messages = await apiCall(`/messages${query}`);
  render();
}

async function fetchConversations() {
  state.conversations = await apiCall('/conversations');
  render();
}

async function fetchUsers(query = '') {
  state.users = await apiCall(`/users/search?q=${query}`);
  render();
}

// Profile actions
async function updateProfile(displayName, bio, status) {
  await apiCall('/profile', {
    method: 'PATCH',
    body: JSON.stringify({ display_name: displayName, bio, status }),
  });

  // Refresh user data
  state.user = await apiCall('/auth/me');
  render();
}

async function blockUser(userId) {
  await apiCall(`/users/${userId}/block`, { method: 'POST' });
  alert('Користувача заблоковано');
}

// Message actions
async function sendMessage(content, topicId, groupId, quotedMessageId) {
  await apiCall('/messages', {
    method: 'POST',
    body: JSON.stringify({
      content,
      topic_id: topicId || null,
      group_id: groupId || null,
      quoted_message_id: quotedMessageId || null,
    }),
  });

  // Stop typing
  const roomID = topicId ? 'topic_' + topicId : 'group_' + groupId;
  apiCall(`/messages/typing/stop?room=${roomID}`, { method: 'POST' }).catch(() => {});
}

async function editMessage(messageId, newContent) {
  await apiCall(`/messages/${messageId}`, {
    method: 'PATCH',
    body: JSON.stringify({ content: newContent }),
  });

  await fetchMessages(state.currentTopic);
}

async function deleteMessage(messageId) {
  if (!confirm('Видалити повідомлення?')) return;

  await apiCall(`/messages/${messageId}`, { method: 'DELETE' });
  await fetchMessages(state.currentTopic);
}

// DM actions
async function sendDirectMessage(recipientId, content) {
  await apiCall('/conversations/send', {
    method: 'POST',
    body: JSON.stringify({ recipient_id: recipientId, content }),
  });

  await fetchConversations();
}

// Render functions
function render() {
  const app = document.getElementById('app');

  if (!state.token) {
    app.innerHTML = renderAuth();
  } else {
    app.innerHTML = `
      ${renderNavbar()}
      <div class="container">
        ${renderView()}
      </div>
    `;
  }

  attachEventListeners();
}

function attachEventListeners() {
  // Navigation
  document.querySelectorAll('[data-view]').forEach(link => {
    link.addEventListener('click', (e) => {
      e.preventDefault();
      state.currentView = e.target.dataset.view;
      render();
    });
  });

  // Auth
  const authBtn = document.getElementById('auth-btn');
  const toggleAuth = document.getElementById('toggle-auth');

  if (authBtn) {
    authBtn.addEventListener('click', async () => {
      const username = document.getElementById('username').value;
      const password = document.getElementById('password').value;

      try {
        const endpoint = toggleAuth && toggleAuth.textContent === 'Увійти' ? '/auth/register' : '/auth/login';
        const response = await fetch(`${API_BASE}${endpoint}`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username, password, display_name: username })
        });

        const data = await response.json();
        if (response.ok) {
          state.token = data.token;
          state.user = data.user;
          localStorage.setItem('token', data.token);
          connectWebSocket();
          render();
        } else {
          alert(data.error || 'Authentication failed');
        }
      } catch (error) {
        alert('Network error: ' + error.message);
      }
    });
  }

  if (toggleAuth) {
    toggleAuth.addEventListener('click', () => {
      const authBtn = document.getElementById('auth-btn');
      if (toggleAuth.textContent === 'Реєстрація') {
        toggleAuth.textContent = 'Увійти';
        authBtn.textContent = 'Зареєструватись';
      } else {
        toggleAuth.textContent = 'Реєстрація';
        authBtn.textContent = 'Увійти';
      }
    });
  }
}

function renderNavbar() {
  return `
    <nav class="navbar glass">
      <div class="nav-content">
        <a href="#" class="logo">🧠 Психологічна Платформа</a>
        <ul class="nav-links">
          <li><a href="#" class="nav-link ${state.currentView === 'topics' ? 'active' : ''}" data-view="topics">💬 Теми</a></li>
          <li><a href="#" class="nav-link ${state.currentView === 'conversations' ? 'active' : ''}" data-view="conversations">📨 Повідомлення</a></li>
          <li><a href="#" class="nav-link ${state.currentView === 'groups' ? 'active' : ''}" data-view="groups">👥 Групи</a></li>
          <li><a href="#" class="nav-link ${state.currentView === 'sessions' ? 'active' : ''}" data-view="sessions">🎥 Вебінари</a></li>
          <li><a href="#" class="nav-link ${state.currentView === 'users' ? 'active' : ''}" data-view="users">👤 Користувачі</a></li>
          <li><a href="#" class="nav-link ${state.currentView === 'profile' ? 'active' : ''}" data-view="profile">⚙️ Профіль</a></li>
        </ul>
        <div style="display: flex; align-items: center; gap: 1rem;">
          <div class="avatar">${state.user?.display_name?.charAt(0) || 'U'}</div>
          <button class="btn btn-secondary" onclick="logout()">Вийти</button>
        </div>
      </div>
    </nav>
  `;
}

function renderAuth() {
  return `
    <div class="container" style="max-width: 400px; margin-top: 10vh;">
      <div class="glass" style="padding: 2rem;">
        <h1 style="text-align: center; margin-bottom: 2rem;">🧠 Вітаємо</h1>
        <div id="auth-form">
          <div class="form-group">
            <label class="form-label">Логін</label>
            <input type="text" id="username" class="form-input" placeholder="Введіть логін">
          </div>
          <div class="form-group">
            <label class="form-label">Пароль</label>
            <input type="password" id="password" class="form-input" placeholder="Введіть пароль">
          </div>
          <div class="form-group" id="display-name-group" style="display: none;">
            <label class="form-label">Ім'я для відображення</label>
            <input type="text" id="display-name" class="form-input" placeholder="Як вас звати?">
          </div>
          <button class="btn btn-primary" style="width: 100%; margin-bottom: 1rem;" id="auth-btn">Увійти</button>
          <button class="btn btn-secondary" style="width: 100%;" id="toggle-auth">Реєстрація</button>
        </div>
      </div>
    </div>
  `;
}

function renderView() {
  switch (state.currentView) {
    case 'topics':
      return renderTopics();
    case 'topic-detail':
      return renderTopicDetail();
    case 'conversations':
      return renderConversations();
    case 'users':
      return renderUsers();
    case 'profile':
      return renderProfile();
    default:
      return '<div class="loading"><div class="spinner"></div>Завантаження...</div>';
  }
}

function renderProfile() {
  return `
    <div class="glass" style="max-width: 600px; margin: 2rem auto; padding: 2rem;">
      <h1 style="margin-bottom: 2rem;">⚙️ Налаштування профілю</h1>
      <div class="form-group">
        <label class="form-label">Ім'я для відображення</label>
        <input type="text" id="profile-name" class="form-input" value="${state.user?.display_name || ''}" placeholder="Ваше ім'я">
      </div>
      <div class="form-group">
        <label class="form-label">Біографія</label>
        <textarea id="profile-bio" class="form-input" rows="4" placeholder="Розкажіть про себе...">${state.user?.bio || ''}</textarea>
      </div>
      <div class="form-group">
        <label class="form-label">Статус</label>
        <input type="text" id="profile-status" class="form-input" value="${state.user?.status || ''}" placeholder="Ваш статус">
      </div>
      <button class="btn btn-primary" onclick="saveProfile()">Зберегти</button>
    </div>
  `;
}

function renderUsers() {
  return `
    <div>
      <h1 style="margin-bottom: 1rem;">👤 Користувачі</h1>
      <input type="text" class="form-input" style="margin-bottom: 2rem;" placeholder="🔍 Пошук..." onkeyup="searchUsers(this.value)">
      <div class="grid grid-3">
        ${state.users.map(user => `
          <div class="glass-card">
            <div style="display: flex; align-items: center; gap: 1rem; margin-bottom: 1rem;">
              <div class="avatar" style="position: relative;">
                ${user.display_name?.charAt(0) || 'U'}
                ${user.is_online ? '<span style="position: absolute; bottom: 0; right: 0; width: 12px; height: 12px; background: #10b981; border: 2px solid var(--darker); border-radius: 50%;"></span>' : ''}
              </div>
              <div>
                <strong>${user.display_name}</strong>
                <div style="font-size: 0.9rem; color: var(--text-secondary);">@${user.username}</div>
              </div>
            </div>
            ${user.bio ? `<p style="color: var(--text-secondary); margin-bottom: 1rem;">${user.bio}</p>` : ''}
            ${user.is_psychologist ? '<span class="badge badge-success">Психолог</span>' : ''}
            <button class="btn btn-primary" style="width: 100%; margin-top: 1rem;" onclick="startConversation('${user.id}')">💬 Написати</button>
          </div>
        `).join('') || '<p>Поки немає користувачів</p>'}
      </div>
    </div>
  `;
}

function renderTopicDetail() {
  const topic = state.topics.find(t => t.id === state.currentTopic);
  if (!topic) return renderTopics();

  return `
    <div>
      <button class="btn btn-secondary" onclick="backToTopics()" style="margin-bottom: 1rem;">← Назад</button>
      <div class="glass" style="padding: 2rem; margin-bottom: 2rem;">
        <h1>${topic.title}</h1>
        <p style="color: var(--text-secondary); margin: 1rem 0;">${topic.description || ''}</p>
        <div style="display: flex; gap: 1rem; align-items: center;">
          <span>${topic.messages_count} повідомлень</span>
          <button class="reaction" onclick="voteTopic('${topic.id}', 'up')">👍 ${topic.votes_count}</button>
        </div>
      </div>

      <div id="typing-indicator" style="display: none; color: var(--text-secondary); margin-bottom: 1rem; font-style: italic;"></div>

      <div id="messages-container" style="margin-bottom: 2rem;">
        ${state.messages.map(msg => `
          <div class="message ${msg.is_deleted ? 'deleted' : ''}" data-message-id="${msg.id}">
            <div class="message-header">
              <div style="display: flex; align-items: center; gap: 0.5rem;">
                <div class="avatar" style="width: 32px; height: 32px; font-size: 0.9rem;">${msg.user?.display_name?.charAt(0) || 'U'}</div>
                <span class="message-author">${msg.user?.display_name || 'Unknown'}</span>
                ${msg.is_edited ? '<span style="font-size: 0.8rem; color: var(--text-secondary);">(змінено)</span>' : ''}
              </div>
              <div style="display: flex; gap: 0.5rem; align-items: center;">
                <span class="message-time">${new Date(msg.created_at).toLocaleString('uk-UA')}</span>
                ${msg.user_id === state.user?.id && !msg.is_deleted ? `
                  <button class="btn-icon" onclick="editMsg('${msg.id}', '${msg.content.replace(/'/g, "\\'")}')">✏️</button>
                  <button class="btn-icon" onclick="deleteMsg('${msg.id}')">🗑️</button>
                ` : ''}
              </div>
            </div>
            <div class="message-content">${parseMarkdown(msg.content)}</div>
            ${msg.reactions?.length ? `
              <div class="reactions">
                ${msg.reactions.map(r => `<span class="reaction">${r.emoji} ${r.count || 1}</span>`).join('')}
              </div>
            ` : ''}
            <button class="btn-icon" onclick="showReactionPicker('${msg.id}')" style="margin-top: 0.5rem;">➕ Реакція</button>
          </div>
        `).join('')}
      </div>

      <div class="glass" style="padding: 1rem;">
        <div id="message-toolbar"></div>
        <textarea id="message-input" class="form-input" rows="3" placeholder="Введіть повідомлення... (підтримує Markdown)" onkeyup="handleTyping('topic_${state.currentTopic}')"></textarea>
        <div style="display: flex; gap: 0.5rem; margin-top: 0.5rem;">
          <button class="btn btn-secondary" onclick="showEmojiForMessage()">😊 Емодзі</button>
          <button class="btn btn-primary" style="flex: 1;" onclick="sendMsg()">Відправити</button>
        </div>
      </div>
    </div>
  `;
}

// Initialize
if (state.token) {
  apiCall('/auth/me')
    .then(user => {
      state.user = user;
      connectWebSocket();
      state.currentView = 'topics';
      loadViewData();
    })
    .catch(() => {
      logout();
    });
} else {
  render();
}

// Export functions to window for onclick handlers
window.logout = logout;
window.voteTopic = (id, type) => apiCall(`/topics/${id}/vote?type=${type}`, { method: 'POST' }).then(fetchTopics);
window.openTopic = (id) => {
  state.currentTopic = id;
  state.currentView = 'topic-detail';
  fetchMessages(id);
  joinRoom('topic_' + id);
};
window.backToTopics = () => {
  state.currentView = 'topics';
  state.currentTopic = null;
  render();
};
window.sendMsg = () => {
  const input = document.getElementById('message-input');
  if (input.value.trim()) {
    sendMessage(input.value, state.currentTopic);
    input.value = '';
  }
};
window.editMsg = (id, oldContent) => {
  const newContent = prompt('Редагувати повідомлення:', oldContent);
  if (newContent && newContent !== oldContent) {
    editMessage(id, newContent);
  }
};
window.deleteMsg = deleteMessage;
window.saveProfile = () => {
  const name = document.getElementById('profile-name').value;
  const bio = document.getElementById('profile-bio').value;
  const status = document.getElementById('profile-status').value;
  updateProfile(name, bio, status);
};
window.searchUsers = (query) => fetchUsers(query);
window.startConversation = (userId) => {
  const message = prompt('Повідомлення:');
  if (message) {
    sendDirectMessage(userId, message);
  }
};
