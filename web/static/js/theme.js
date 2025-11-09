// Theme management
export class ThemeManager {
  constructor() {
    this.currentTheme = localStorage.getItem('theme') || 'dark';
    this.applyTheme();
  }

  toggle() {
    this.currentTheme = this.currentTheme === 'dark' ? 'light' : 'dark';
    localStorage.setItem('theme', this.currentTheme);
    this.applyTheme();
  }

  applyTheme() {
    const root = document.documentElement;

    if (this.currentTheme === 'light') {
      root.style.setProperty('--primary', '#5b21b6');
      root.style.setProperty('--primary-dark', '#4c1d95');
      root.style.setProperty('--secondary', '#7c3aed');
      root.style.setProperty('--dark', '#f3f4f6');
      root.style.setProperty('--darker', '#ffffff');
      root.style.setProperty('--light', '#1f2937');
      root.style.setProperty('--text-primary', '#111827');
      root.style.setProperty('--text-secondary', '#4b5563');
      root.style.setProperty('--glass-bg', 'rgba(255, 255, 255, 0.8)');
      root.style.setProperty('--glass-border', 'rgba(0, 0, 0, 0.1)');
      root.style.setProperty('--border', 'rgba(0, 0, 0, 0.1)');

      document.body.style.background = 'linear-gradient(135deg, #f3f4f6 0%, #e5e7eb 100%)';
      document.body.style.color = '#111827';
    } else {
      root.style.setProperty('--primary', '#6366f1');
      root.style.setProperty('--primary-dark', '#4f46e5');
      root.style.setProperty('--secondary', '#8b5cf6');
      root.style.setProperty('--dark', '#1f2937');
      root.style.setProperty('--darker', '#111827');
      root.style.setProperty('--light', '#f3f4f6');
      root.style.setProperty('--text-primary', '#f9fafb');
      root.style.setProperty('--text-secondary', '#d1d5db');
      root.style.setProperty('--glass-bg', 'rgba(255, 255, 255, 0.05)');
      root.style.setProperty('--glass-border', 'rgba(255, 255, 255, 0.1)');
      root.style.setProperty('--border', 'rgba(255, 255, 255, 0.1)');

      document.body.style.background = 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)';
      document.body.style.color = '#f9fafb';
    }
  }

  getCurrentTheme() {
    return this.currentTheme;
  }
}

export function createThemeToggle() {
  const toggle = document.createElement('button');
  toggle.className = 'theme-toggle btn btn-secondary';
  toggle.innerHTML = '🌓';
  toggle.title = 'Toggle theme';
  toggle.style.cssText = `
    position: fixed;
    bottom: 2rem;
    right: 2rem;
    width: 50px;
    height: 50px;
    border-radius: 50%;
    font-size: 1.5rem;
    z-index: 1000;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  `;

  const themeManager = new ThemeManager();

  toggle.addEventListener('click', () => {
    themeManager.toggle();
    toggle.innerHTML = themeManager.getCurrentTheme() === 'dark' ? '🌓' : '☀️';
  });

  toggle.innerHTML = themeManager.getCurrentTheme() === 'dark' ? '🌓' : '☀️';

  return { toggle, themeManager };
}
