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
  topics: [],
  messages: [],
  groups: [],
  sessions: [],
  appointments: [],
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
    setTimeout(connectWebSocket, 3000);
  };
}

function handleWebSocketMessage(data) {
  if (data.type === 'new_message') {
    state.messages.unshift(data.payload);
    render();
  }
}

function joinRoom(roomId) {
  if (state.ws && state.ws.readyState === WebSocket.OPEN) {
    state.ws.send(JSON.stringify({
      type: 'join_room',
      room: roomId,
    }));
  }
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

async function fetchGroups() {
  state.groups = await apiCall('/groups');
  render();
}

async function fetchSessions() {
  state.sessions = await apiCall('/sessions');
  render();
}

async function fetchAppointments() {
  state.appointments = await apiCall('/appointments');
  render();
}

// Actions
async function createTopic(title, description, isPublic) {
  await apiCall('/topics', {
    method: 'POST',
    body: JSON.stringify({ title, description, is_public: isPublic }),
  });
  await fetchTopics();
}

async function voteTopic(topicId, voteType) {
  await apiCall(`/topics/${topicId}/vote?type=${voteType}`, { method: 'POST' });
  await fetchTopics();
}

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
}

async function addReaction(messageId, emoji) {
  await apiCall(`/messages/${messageId}/reactions`, {
    method: 'POST',
    body: JSON.stringify({ emoji }),
  });
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

function renderNavbar() {
  return `
    <nav class="navbar glass">
      <div class="nav-content">
        <a href="#" class="logo">Психологічна Платформа</a>
        <ul class="nav-links">
          <li><a href="#" class="nav-link ${state.currentView === 'topics' ? 'active' : ''}" data-view="topics">Теми</a></li>
          <li><a href="#" class="nav-link ${state.currentView === 'groups' ? 'active' : ''}" data-view="groups">Групи</a></li>
          <li><a href="#" class="nav-link ${state.currentView === 'sessions' ? 'active' : ''}" data-view="sessions">Вебінари</a></li>
          <li><a href="#" class="nav-link ${state.currentView === 'appointments' ? 'active' : ''}" data-view="appointments">Зустрічі</a></li>
          ${state.user?.role === 'admin' ? '<li><a href="#" class="nav-link" data-view="admin">Адмін</a></li>' : ''}
        </ul>
        <button class="btn btn-secondary" onclick="logout()">Вийти</button>
      </div>
    </nav>
  `;
}

function renderAuth() {
  return `
    <div class="container" style="max-width: 400px; margin-top: 10vh;">
      <div class="glass" style="padding: 2rem;">
        <h1 style="text-align: center; margin-bottom: 2rem;">Вітаємо</h1>
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
    case 'groups':
      return renderGroups();
    case 'sessions':
      return renderSessions();
    case 'appointments':
      return renderAppointments();
    default:
      return '<div class="loading">Завантаження...</div>';
  }
}

function renderTopics() {
  return `
    <div>
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem;">
        <h1>Теми для обговорення</h1>
        <button class="btn btn-primary" onclick="showCreateTopicModal()">+ Створити тему</button>
      </div>
      <div class="grid grid-2">
        ${state.topics.map(topic => `
          <div class="glass-card fade-in" onclick="openTopic('${topic.id}')">
            <h3 style="margin-bottom: 0.5rem;">${topic.title}</h3>
            <p style="color: var(--text-secondary); margin-bottom: 1rem;">${topic.description || ''}</p>
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <span class="badge">${topic.messages_count} повідомлень</span>
              <div style="display: flex; gap: 0.5rem; align-items: center;">
                <button class="reaction" onclick="event.stopPropagation(); voteTopic('${topic.id}', 'up')">👍 ${topic.votes_count}</button>
                ${topic.is_public ? '<span class="badge badge-success">Публічна</span>' : '<span class="badge badge-warning">Приватна</span>'}
              </div>
            </div>
          </div>
        `).join('') || '<p style="text-align: center; color: var(--text-secondary);">Поки немає тем. Створіть першу!</p>'}
      </div>
    </div>
  `;
}

function renderGroups() {
  return `
    <div>
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem;">
        <h1>Групи</h1>
        <button class="btn btn-primary">+ Створити групу</button>
      </div>
      <div class="grid grid-3">
        ${state.groups.map(group => `
          <div class="glass-card">
            <h3>${group.name}</h3>
            <p style="color: var(--text-secondary); margin: 0.5rem 0 1rem;">${group.description || ''}</p>
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <span>${group.members_count} учасників</span>
              <button class="btn ${group.is_member ? 'btn-secondary' : 'btn-primary'}">${group.is_member ? 'Вийти' : 'Приєднатись'}</button>
            </div>
          </div>
        `).join('') || '<p>Поки немає груп</p>'}
      </div>
    </div>
  `;
}

function renderSessions() {
  return `
    <div>
      <h1 style="margin-bottom: 2rem;">Вебінари та зустрічі</h1>
      <div class="grid grid-2">
        ${state.sessions.map(session => `
          <div class="glass-card">
            <h3>${session.title}</h3>
            <p style="color: var(--text-secondary); margin: 0.5rem 0;">${session.description || ''}</p>
            <p style="margin: 0.5rem 0;"><strong>Психолог:</strong> ${session.psychologist?.display_name || 'Невідомий'}</p>
            <p style="margin: 0.5rem 0;"><strong>Дата:</strong> ${new Date(session.scheduled_at).toLocaleString('uk-UA')}</p>
            <button class="btn btn-primary" style="margin-top: 1rem; width: 100%;">Приєднатись</button>
          </div>
        `).join('') || '<p>Поки немає запланованих сесій</p>'}
      </div>
    </div>
  `;
}

function renderAppointments() {
  return `
    <div>
      <h1 style="margin-bottom: 2rem;">Мої зустрічі</h1>
      <div class="grid grid-2">
        ${state.appointments.map(apt => `
          <div class="glass-card">
            <h3>${apt.title || 'Консультація'}</h3>
            <p><strong>З:</strong> ${apt.psychologist?.display_name || apt.client?.display_name}</p>
            <p><strong>Дата:</strong> ${new Date(apt.scheduled_at).toLocaleString('uk-UA')}</p>
            <p><strong>Тривалість:</strong> ${apt.duration_minutes} хв</p>
            <span class="badge badge-${apt.status === 'confirmed' ? 'success' : 'warning'}">${apt.status}</span>
          </div>
        `).join('') || '<p>Поки немає зустрічей</p>'}
      </div>
    </div>
  `;
}

// Event listeners
function attachEventListeners() {
  // Navigation
  document.querySelectorAll('[data-view]').forEach(link => {
    link.addEventListener('click', (e) => {
      e.preventDefault();
      state.currentView = e.target.dataset.view;
      loadViewData();
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
      const displayName = document.getElementById('display-name').value;

      try {
        if (toggleAuth.textContent === 'Увійти') {
          await register(username, password, displayName || username);
        } else {
          await login(username, password);
        }
      } catch (error) {
        alert(error.message);
      }
    });
  }

  if (toggleAuth) {
    toggleAuth.addEventListener('click', () => {
      const displayNameGroup = document.getElementById('display-name-group');
      if (toggleAuth.textContent === 'Реєстрація') {
        toggleAuth.textContent = 'Увійти';
        authBtn.textContent = 'Зареєструватись';
        displayNameGroup.style.display = 'block';
      } else {
        toggleAuth.textContent = 'Реєстрація';
        authBtn.textContent = 'Увійти';
        displayNameGroup.style.display = 'none';
      }
    });
  }
}

async function loadViewData() {
  switch (state.currentView) {
    case 'topics':
      await fetchTopics();
      break;
    case 'groups':
      await fetchGroups();
      break;
    case 'sessions':
      await fetchSessions();
      break;
    case 'appointments':
      await fetchAppointments();
      break;
  }
}

function showCreateTopicModal() {
  const modal = document.createElement('div');
  modal.className = 'modal';
  modal.innerHTML = `
    <div class="glass modal-content">
      <h2 style="margin-bottom: 1.5rem;">Створити нову тему</h2>
      <div class="form-group">
        <label class="form-label">Назва теми</label>
        <input type="text" id="topic-title" class="form-input" placeholder="Введіть назву теми">
      </div>
      <div class="form-group">
        <label class="form-label">Опис</label>
        <textarea id="topic-description" class="form-input" rows="4" placeholder="Опишіть тему"></textarea>
      </div>
      <div class="form-group">
        <label style="display: flex; align-items: center; gap: 0.5rem; cursor: pointer;">
          <input type="checkbox" id="topic-public" checked>
          <span>Публічна тема</span>
        </label>
      </div>
      <div style="display: flex; gap: 1rem;">
        <button class="btn btn-primary" style="flex: 1;" id="create-topic-btn">Створити</button>
        <button class="btn btn-secondary" style="flex: 1;" onclick="this.closest('.modal').remove()">Скасувати</button>
      </div>
    </div>
  `;

  document.body.appendChild(modal);

  document.getElementById('create-topic-btn').addEventListener('click', async () => {
    const title = document.getElementById('topic-title').value;
    const description = document.getElementById('topic-description').value;
    const isPublic = document.getElementById('topic-public').checked;

    try {
      await createTopic(title, description, isPublic);
      modal.remove();
    } catch (error) {
      alert(error.message);
    }
  });

  modal.addEventListener('click', (e) => {
    if (e.target === modal) modal.remove();
  });
}

function openTopic(topicId) {
  console.log('Opening topic:', topicId);
  // TODO: Implement topic view with messages
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
