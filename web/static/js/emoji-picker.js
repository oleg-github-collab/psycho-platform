// Emoji picker component
export const EMOJI_CATEGORIES = {
  smileys: ['😀', '😃', '😄', '😁', '😅', '😂', '🤣', '😊', '😇', '🙂', '🙃', '😉', '😌', '😍', '🥰', '😘', '😗', '😙', '😚', '😋', '😛', '😝', '😜', '🤪', '🤨', '🧐', '🤓', '😎', '🥸', '🤩', '🥳'],
  emotions: ['😭', '😢', '😥', '😰', '😨', '😱', '😖', '😣', '😞', '😓', '😩', '😫', '🥱', '😤', '😡', '😠', '🤬', '😈', '👿'],
  gestures: ['👍', '👎', '👌', '✌️', '🤞', '🤟', '🤘', '🤙', '👈', '👉', '👆', '👇', '☝️', '✋', '🤚', '🖐️', '🖖', '👋', '🤝', '🙏'],
  hearts: ['❤️', '🧡', '💛', '💚', '💙', '💜', '🖤', '🤍', '🤎', '💔', '❤️‍🔥', '❤️‍🩹', '💕', '💞', '💓', '💗', '💖', '💘', '💝'],
  symbols: ['✨', '⭐', '🌟', '💫', '🔥', '💥', '💢', '💯', '✅', '❌', '⚠️', '🎉', '🎊', '🎈'],
};

export function createEmojiPicker(onSelect) {
  const picker = document.createElement('div');
  picker.className = 'emoji-picker glass';
  picker.style.cssText = `
    position: absolute;
    bottom: 100%;
    right: 0;
    width: 320px;
    max-height: 400px;
    padding: 1rem;
    margin-bottom: 0.5rem;
    z-index: 1000;
    overflow-y: auto;
  `;

  // Tabs
  const tabs = document.createElement('div');
  tabs.style.cssText = `
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1rem;
    border-bottom: 1px solid var(--glass-border);
    padding-bottom: 0.5rem;
  `;

  const categoryIcons = {
    smileys: '😊',
    emotions: '😢',
    gestures: '👍',
    hearts: '❤️',
    symbols: '✨',
  };

  let activeCategory = 'smileys';

  Object.keys(EMOJI_CATEGORIES).forEach(cat => {
    const tab = document.createElement('button');
    tab.type = 'button';
    tab.textContent = categoryIcons[cat];
    tab.className = 'emoji-tab';
    tab.style.cssText = `
      padding: 0.5rem;
      background: ${cat === activeCategory ? 'rgba(99, 102, 241, 0.3)' : 'transparent'};
      border: none;
      border-radius: 8px;
      font-size: 1.2rem;
      cursor: pointer;
      transition: all 0.2s;
    `;

    tab.addEventListener('click', () => {
      activeCategory = cat;
      renderEmojis();
      document.querySelectorAll('.emoji-tab').forEach(t => {
        t.style.background = 'transparent';
      });
      tab.style.background = 'rgba(99, 102, 241, 0.3)';
    });

    tabs.appendChild(tab);
  });

  picker.appendChild(tabs);

  // Emoji grid
  const grid = document.createElement('div');
  grid.className = 'emoji-grid';
  grid.style.cssText = `
    display: grid;
    grid-template-columns: repeat(8, 1fr);
    gap: 0.5rem;
  `;

  function renderEmojis() {
    grid.innerHTML = '';
    EMOJI_CATEGORIES[activeCategory].forEach(emoji => {
      const btn = document.createElement('button');
      btn.type = 'button';
      btn.textContent = emoji;
      btn.className = 'emoji-btn';
      btn.style.cssText = `
        padding: 0.5rem;
        background: transparent;
        border: none;
        font-size: 1.5rem;
        cursor: pointer;
        border-radius: 8px;
        transition: all 0.2s;
      `;

      btn.addEventListener('mouseover', () => {
        btn.style.background = 'rgba(255, 255, 255, 0.1)';
        btn.style.transform = 'scale(1.2)';
      });

      btn.addEventListener('mouseout', () => {
        btn.style.background = 'transparent';
        btn.style.transform = 'scale(1)';
      });

      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        onSelect(emoji);
        picker.remove();
      });

      grid.appendChild(btn);
    });
  }

  renderEmojis();
  picker.appendChild(grid);

  // Search
  const search = document.createElement('input');
  search.type = 'text';
  search.placeholder = 'Пошук емодзі...';
  search.className = 'form-input';
  search.style.marginTop = '1rem';

  search.addEventListener('input', (e) => {
    const query = e.target.value.toLowerCase();
    if (!query) {
      renderEmojis();
      return;
    }

    grid.innerHTML = '';
    Object.values(EMOJI_CATEGORIES).flat().forEach(emoji => {
      const btn = document.createElement('button');
      btn.type = 'button';
      btn.textContent = emoji;
      btn.className = 'emoji-btn';
      btn.style.cssText = `
        padding: 0.5rem;
        background: transparent;
        border: none;
        font-size: 1.5rem;
        cursor: pointer;
        border-radius: 8px;
        transition: all 0.2s;
      `;

      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        onSelect(emoji);
        picker.remove();
      });

      grid.appendChild(btn);
    });
  });

  picker.appendChild(search);

  // Close on outside click
  setTimeout(() => {
    document.addEventListener('click', function closePickerHandler(e) {
      if (!picker.contains(e.target)) {
        picker.remove();
        document.removeEventListener('click', closePickerHandler);
      }
    });
  }, 100);

  return picker;
}

export function showEmojiPicker(targetElement, onSelect) {
  const picker = createEmojiPicker(onSelect);
  targetElement.style.position = 'relative';
  targetElement.appendChild(picker);
}
