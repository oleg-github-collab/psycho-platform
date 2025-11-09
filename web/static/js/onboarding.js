// Onboarding tutorial
export class OnboardingTour {
  constructor() {
    this.steps = [
      {
        target: '.logo',
        title: '🧠 Вітаємо!',
        content: 'Ласкаво просимо на психологічну платформу. Давайте швидко ознайомимося з основними функціями.',
        position: 'bottom',
      },
      {
        target: '[data-view="topics"]',
        title: '💬 Теми',
        content: 'Тут ви можете обговорювати різні теми з іншими учасниками. Голосуйте за цікаві теми!',
        position: 'bottom',
      },
      {
        target: '[data-view="conversations"]',
        title: '📨 Повідомлення',
        content: 'Особисті повідомлення для приватного спілкування з користувачами.',
        position: 'bottom',
      },
      {
        target: '[data-view="groups"]',
        title: '👥 Групи',
        content: 'Приєднуйтесь до груп за інтересами або створюйте власні.',
        position: 'bottom',
      },
      {
        target: '[data-view="sessions"]',
        title: '🎥 Вебінари',
        content: 'Онлайн-зустрічі та вебінари з психологами в режимі реального часу.',
        position: 'bottom',
      },
      {
        target: '[data-view="users"]',
        title: '👤 Користувачі',
        content: 'Шукайте психологів та інших учасників платформи.',
        position: 'bottom',
      },
      {
        target: '[data-view="profile"]',
        title: '⚙️ Профіль',
        content: 'Налаштуйте свій профіль, додайте біографію та статус.',
        position: 'bottom',
      },
      {
        target: '.theme-toggle',
        title: '🌓 Тема',
        content: 'Перемикайте між темною та світлою темою для комфорту.',
        position: 'left',
      },
    ];

    this.currentStep = 0;
    this.overlay = null;
    this.tooltip = null;
  }

  start() {
    if (localStorage.getItem('onboarding_completed') === 'true') {
      return;
    }

    this.showStep(0);
  }

  showStep(index) {
    if (index >= this.steps.length) {
      this.complete();
      return;
    }

    this.currentStep = index;
    const step = this.steps[index];

    // Remove existing overlay and tooltip
    this.cleanup();

    // Create overlay
    this.overlay = document.createElement('div');
    this.overlay.className = 'onboarding-overlay';
    this.overlay.style.cssText = `
      position: fixed;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      background: rgba(0, 0, 0, 0.7);
      z-index: 9998;
      backdrop-filter: blur(3px);
    `;

    // Find target element
    const target = document.querySelector(step.target);
    if (!target) {
      this.showStep(index + 1);
      return;
    }

    // Highlight target
    const rect = target.getBoundingClientRect();
    const highlight = document.createElement('div');
    highlight.className = 'onboarding-highlight';
    highlight.style.cssText = `
      position: fixed;
      top: ${rect.top - 4}px;
      left: ${rect.left - 4}px;
      width: ${rect.width + 8}px;
      height: ${rect.height + 8}px;
      border: 3px solid var(--primary);
      border-radius: 12px;
      z-index: 9999;
      pointer-events: none;
      box-shadow: 0 0 0 9999px rgba(0, 0, 0, 0.7);
    `;

    // Create tooltip
    this.tooltip = document.createElement('div');
    this.tooltip.className = 'onboarding-tooltip glass';
    this.tooltip.style.cssText = `
      position: fixed;
      z-index: 10000;
      max-width: 350px;
      padding: 1.5rem;
    `;

    this.tooltip.innerHTML = `
      <h3 style="margin: 0 0 1rem 0; font-size: 1.25rem;">${step.title}</h3>
      <p style="margin: 0 0 1.5rem 0; color: var(--text-secondary);">${step.content}</p>
      <div style="display: flex; justify-content: space-between; align-items: center;">
        <span style="color: var(--text-secondary); font-size: 0.9rem;">${index + 1} / ${this.steps.length}</span>
        <div style="display: flex; gap: 0.5rem;">
          ${index > 0 ? '<button class="btn btn-secondary" id="onboarding-prev">← Назад</button>' : ''}
          ${index < this.steps.length - 1
            ? '<button class="btn btn-primary" id="onboarding-next">Далі →</button>'
            : '<button class="btn btn-primary" id="onboarding-finish">Завершити</button>'}
        </div>
      </div>
      <button style="position: absolute; top: 0.5rem; right: 0.5rem; background: transparent; border: none; cursor: pointer; font-size: 1.5rem; opacity: 0.5;" id="onboarding-skip">×</button>
    `;

    // Position tooltip
    this.positionTooltip(this.tooltip, rect, step.position);

    document.body.appendChild(highlight);
    document.body.appendChild(this.tooltip);
    this.overlay.appendChild(highlight);

    // Event listeners
    const nextBtn = document.getElementById('onboarding-next');
    const prevBtn = document.getElementById('onboarding-prev');
    const finishBtn = document.getElementById('onboarding-finish');
    const skipBtn = document.getElementById('onboarding-skip');

    if (nextBtn) {
      nextBtn.addEventListener('click', () => this.showStep(index + 1));
    }

    if (prevBtn) {
      prevBtn.addEventListener('click', () => this.showStep(index - 1));
    }

    if (finishBtn) {
      finishBtn.addEventListener('click', () => this.complete());
    }

    if (skipBtn) {
      skipBtn.addEventListener('click', () => this.complete());
    }
  }

  positionTooltip(tooltip, rect, position) {
    const margin = 20;

    switch (position) {
      case 'top':
        tooltip.style.left = `${rect.left + rect.width / 2}px`;
        tooltip.style.bottom = `${window.innerHeight - rect.top + margin}px`;
        tooltip.style.transform = 'translateX(-50%)';
        break;
      case 'bottom':
        tooltip.style.left = `${rect.left + rect.width / 2}px`;
        tooltip.style.top = `${rect.bottom + margin}px`;
        tooltip.style.transform = 'translateX(-50%)';
        break;
      case 'left':
        tooltip.style.right = `${window.innerWidth - rect.left + margin}px`;
        tooltip.style.top = `${rect.top + rect.height / 2}px`;
        tooltip.style.transform = 'translateY(-50%)';
        break;
      case 'right':
        tooltip.style.left = `${rect.right + margin}px`;
        tooltip.style.top = `${rect.top + rect.height / 2}px`;
        tooltip.style.transform = 'translateY(-50%)';
        break;
    }
  }

  cleanup() {
    if (this.overlay) {
      this.overlay.remove();
      this.overlay = null;
    }
    if (this.tooltip) {
      this.tooltip.remove();
      this.tooltip = null;
    }
    document.querySelectorAll('.onboarding-highlight').forEach(el => el.remove());
  }

  complete() {
    this.cleanup();
    localStorage.setItem('onboarding_completed', 'true');
  }

  reset() {
    localStorage.removeItem('onboarding_completed');
    this.currentStep = 0;
  }
}

export function startOnboarding() {
  setTimeout(() => {
    const tour = new OnboardingTour();
    tour.start();
  }, 1000);
}
