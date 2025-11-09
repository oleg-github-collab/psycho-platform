// Comprehensive Admin Panel
import { apiCall } from './app-enhanced.js';

const ROLE_ORDER = ['super_admin', 'premium', 'basic'];
const ROLE_DETAILS = {
  super_admin: { label: 'Суперадмін', icon: '👑', color: '#6366f1' },
  premium: { label: 'Преміум', icon: '⭐️', color: '#f97316' },
  basic: { label: 'Базовий', icon: '👤', color: 'rgba(255,255,255,0.2)' },
};

const getRoleDetail = (role) => ROLE_DETAILS[role] || ROLE_DETAILS.basic;

const renderRoleBadge = (role) => {
  const detail = getRoleDetail(role);
  return `<span class="badge" style="background: ${detail.color}; color: #fff;">${detail.icon} ${detail.label}</span>`;
};

const renderRoleOptions = (selectedRole) =>
  ROLE_ORDER.map(role => {
    const detail = getRoleDetail(role);
    return `<option value="${role}" ${selectedRole === role ? 'selected' : ''}>${detail.icon} ${detail.label}</option>`;
  }).join('');

export class AdminPanel {
  constructor() {
    this.stats = null;
    this.users = [];
    this.currentView = 'dashboard';
  }

  async render() {
    const container = document.getElementById('admin-container');
    if (!container) return;

    container.innerHTML = `
      <div class="admin-panel">
        <div class="admin-sidebar glass">
          <h2 style="padding: 1rem; margin: 0;">⚙️ Адмін-панель</h2>
          <nav class="admin-nav">
            <button class="admin-nav-btn ${this.currentView === 'dashboard' ? 'active' : ''}" data-view="dashboard">
              📊 Дашборд
            </button>
            <button class="admin-nav-btn ${this.currentView === 'users' ? 'active' : ''}" data-view="users">
              👥 Користувачі
            </button>
            <button class="admin-nav-btn ${this.currentView === 'topics' ? 'active' : ''}" data-view="topics">
              💬 Теми
            </button>
            <button class="admin-nav-btn ${this.currentView === 'groups' ? 'active' : ''}" data-view="groups">
              🏘️ Групи
            </button>
            <button class="admin-nav-btn ${this.currentView === 'sessions' ? 'active' : ''}" data-view="sessions">
              🎥 Сесії
            </button>
            <button class="admin-nav-btn ${this.currentView === 'moderation' ? 'active' : ''}" data-view="moderation">
              🛡️ Модерація
            </button>
            <button class="admin-nav-btn ${this.currentView === 'settings' ? 'active' : ''}" data-view="settings">
              ⚙️ Налаштування
            </button>
          </nav>
        </div>
        <div class="admin-content">
          <div id="admin-view-content"></div>
        </div>
      </div>
    `;

    this.attachNavListeners();
    await this.renderView();
  }

  attachNavListeners() {
    document.querySelectorAll('.admin-nav-btn').forEach(btn => {
      btn.addEventListener('click', async (e) => {
        this.currentView = e.target.dataset.view;
        document.querySelectorAll('.admin-nav-btn').forEach(b => b.classList.remove('active'));
        e.target.classList.add('active');
        await this.renderView();
      });
    });
  }

  async renderView() {
    const content = document.getElementById('admin-view-content');
    if (!content) return;

    try {
      switch (this.currentView) {
        case 'dashboard':
          await this.renderDashboard(content);
          break;
        case 'users':
          await this.renderUsers(content);
          break;
        case 'topics':
          await this.renderTopics(content);
          break;
        case 'groups':
          await this.renderGroups(content);
          break;
        case 'sessions':
          await this.renderSessions(content);
          break;
        case 'moderation':
          await this.renderModeration(content);
          break;
        case 'settings':
          await this.renderSettings(content);
          break;
      }
    } catch (error) {
      content.innerHTML = `
        <div class="glass" style="padding: 2rem; text-align: center;">
          <p style="color: var(--danger);">❌ Помилка завантаження: ${error.message}</p>
          <button class="btn btn-primary" onclick="location.reload()">Перезавантажити</button>
        </div>
      `;
    }
  }

  async renderDashboard(content) {
    this.stats = await apiCall('/admin/stats');

    content.innerHTML = `
      <h1 style="margin-bottom: 2rem;">📊 Статистика платформи</h1>

      <div class="stats-grid">
        <div class="stat-card glass-card">
          <div class="stat-icon">👥</div>
          <div class="stat-value">${this.stats.total_users || 0}</div>
          <div class="stat-label">Користувачів</div>
        </div>

        <div class="stat-card glass-card">
          <div class="stat-icon">💬</div>
          <div class="stat-value">${this.stats.total_topics || 0}</div>
          <div class="stat-label">Тем</div>
        </div>

        <div class="stat-card glass-card">
          <div class="stat-icon">📨</div>
          <div class="stat-value">${this.stats.total_messages || 0}</div>
          <div class="stat-label">Повідомлень</div>
        </div>

        <div class="stat-card glass-card">
          <div class="stat-icon">🏘️</div>
          <div class="stat-value">${this.stats.total_groups || 0}</div>
          <div class="stat-label">Груп</div>
        </div>

        <div class="stat-card glass-card">
          <div class="stat-icon">🎥</div>
          <div class="stat-value">${this.stats.total_sessions || 0}</div>
          <div class="stat-label">Сесій</div>
        </div>

        <div class="stat-card glass-card">
          <div class="stat-icon">👑</div>
          <div class="stat-value">${this.stats.total_super_admins || 0}</div>
          <div class="stat-label">Суперадмінів</div>
        </div>

        <div class="stat-card glass-card">
          <div class="stat-icon">⭐</div>
          <div class="stat-value">${this.stats.total_premium_users || 0}</div>
          <div class="stat-label">Преміум</div>
        </div>

        <div class="stat-card glass-card">
          <div class="stat-icon">👤</div>
          <div class="stat-value">${this.stats.total_basic_users || 0}</div>
          <div class="stat-label">Базових</div>
        </div>
      </div>

      <div class="glass" style="padding: 2rem; margin-top: 2rem;">
        <h2>📈 Швидкі дії</h2>
        <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; margin-top: 1rem;">
          <button class="btn btn-primary" onclick="adminPanel.currentView='users'; adminPanel.renderView()">
            Керувати користувачами
          </button>
          <button class="btn btn-primary" onclick="adminPanel.currentView='moderation'; adminPanel.renderView()">
            Модерація контенту
          </button>
          <button class="btn btn-secondary" onclick="adminPanel.exportData()">
            📥 Експорт даних
          </button>
          <button class="btn btn-secondary" onclick="adminPanel.viewLogs()">
            📋 Переглянути логи
          </button>
        </div>
      </div>
    `;
  }

  async renderUsers(content) {
    this.users = await apiCall('/admin/users');

    content.innerHTML = `
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem;">
        <h1>👥 Користувачі (${this.users.length})</h1>
        <input type="text" class="form-input" placeholder="🔍 Пошук..." style="max-width: 300px;"
               onkeyup="adminPanel.filterUsers(this.value)">
      </div>

      <div class="glass" style="padding: 1rem; overflow-x: auto;">
        <table class="admin-table" id="users-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Користувач</th>
              <th>Роль</th>
              <th>Статус</th>
              <th>Дата реєстрації</th>
              <th>Дії</th>
            </tr>
          </thead>
          <tbody>
            ${this.users.map(user => `
              <tr data-user-id="${user.id}">
                <td><code>${user.id.substring(0, 8)}</code></td>
                <td>
                  <div style="display: flex; align-items: center; gap: 0.5rem;">
                    <div class="avatar" style="width: 32px; height: 32px; font-size: 0.9rem;">
                      ${user.display_name?.charAt(0) || 'U'}
                    </div>
                    <div>
                      <div><strong>${user.display_name || 'N/A'}</strong></div>
                      <div style="font-size: 0.85rem; color: var(--text-secondary);">@${user.username}</div>
                    </div>
                  </div>
                </td>
                <td>${renderRoleBadge(user.role)}</td>
                <td>
                  <span class="badge ${user.is_active ? 'badge-success' : 'badge-warning'}">
                    ${user.is_active ? 'Активний' : 'Деактивований'}
                  </span>
                </td>
                <td>${new Date(user.created_at).toLocaleDateString('uk-UA')}</td>
                <td>
                  <div style="display: flex; gap: 0.25rem; align-items: center;">
                    <button class="btn-icon" onclick="adminPanel.toggleUserStatus('${user.id}', ${!user.is_active})"
                            title="${user.is_active ? 'Деактивувати' : 'Активувати'}">
                      ${user.is_active ? '🔒' : '🔓'}
                    </button>
                    <select class="form-input" style="padding: 0.15rem 0.35rem; font-size: 0.85rem;"
                            onchange="adminPanel.updateUserRole('${user.id}', this.value)">
                      ${renderRoleOptions(user.role)}
                    </select>
                    <button class="btn-icon" onclick="adminPanel.viewUserDetails('${user.id}')" title="Детальніше">
                      👁️
                    </button>
                  </div>
                </td>
              </tr>
            `).join('')}
          </tbody>
        </table>
      </div>
    `;
  }

  filterUsers(query) {
    const rows = document.querySelectorAll('#users-table tbody tr');
    rows.forEach(row => {
      const text = row.textContent.toLowerCase();
      row.style.display = text.includes(query.toLowerCase()) ? '' : 'none';
    });
  }

  async toggleUserStatus(userId, activate) {
    try {
      await apiCall(`/admin/users/${userId}/status?action=${activate ? 'activate' : 'deactivate'}`, {
        method: 'PATCH',
      });
      await this.renderUsers(document.getElementById('admin-view-content'));
    } catch (error) {
      alert('Помилка: ' + error.message);
    }
  }

  async updateUserRole(userId, role) {
    try {
      await apiCall(`/admin/users/${userId}/role`, {
        method: 'PATCH',
        body: JSON.stringify({ role }),
      });
      await this.renderUsers(document.getElementById('admin-view-content'));
    } catch (error) {
      alert('Помилка: ' + error.message);
    }
  }

  async renderModeration(content) {
    content.innerHTML = `
      <h1 style="margin-bottom: 2rem;">🛡️ Модерація</h1>

      <div class="glass" style="padding: 2rem;">
        <h2>Останні повідомлення</h2>
        <p style="color: var(--text-secondary); margin: 1rem 0;">
          Тут відображаються останні повідомлення для модерації
        </p>
        <div id="moderation-messages">
          <div class="loading"><div class="spinner"></div>Завантаження...</div>
        </div>
      </div>

      <div class="glass" style="padding: 2rem; margin-top: 2rem;">
        <h2>Звіти користувачів</h2>
        <p style="color: var(--text-secondary);">
          Функція звітів буде доступна в наступній версії
        </p>
      </div>
    `;
  }

  async renderSettings(content) {
    content.innerHTML = `
      <h1 style="margin-bottom: 2rem;">⚙️ Налаштування</h1>

      <div class="glass" style="padding: 2rem;">
        <h2>Загальні налаштування</h2>

        <div class="form-group">
          <label class="form-label">Назва платформи</label>
          <input type="text" class="form-input" value="Психологічна Платформа">
        </div>

        <div class="form-group">
          <label class="form-label">Макс. розмір файлу (MB)</label>
          <input type="number" class="form-input" value="50">
        </div>

        <div class="form-group">
          <label class="form-label">
            <input type="checkbox" checked> Дозволити реєстрацію
          </label>
        </div>

        <div class="form-group">
          <label class="form-label">
            <input type="checkbox" checked> Модерація повідомлень
          </label>
        </div>

        <button class="btn btn-primary">Зберегти налаштування</button>
      </div>
    `;
  }

  renderTopics(content) {
    content.innerHTML = `
      <h1>💬 Управління темами</h1>
      <div class="glass" style="padding: 2rem; margin-top: 1rem;">
        <p>Список тем для адміністрування</p>
      </div>
    `;
  }

  renderGroups(content) {
    content.innerHTML = `
      <h1>🏘️ Управління групами</h1>
      <div class="glass" style="padding: 2rem; margin-top: 1rem;">
        <p>Список груп для адміністрування</p>
      </div>
    `;
  }

  renderSessions(content) {
    content.innerHTML = `
      <h1>🎥 Управління сесіями</h1>
      <div class="glass" style="padding: 2rem; margin-top: 1rem;">
        <p>Список сесій та вебінарів</p>
      </div>
    `;
  }

  exportData() {
    alert('Експорт даних у розробці');
  }

  viewLogs() {
    alert('Перегляд логів у розробці');
  }

  viewUserDetails(userId) {
    alert('Деталі користувача: ' + userId);
  }
}

// Export for global use
window.AdminPanel = AdminPanel;
window.adminPanel = new AdminPanel();
