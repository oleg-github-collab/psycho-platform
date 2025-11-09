// Simple Markdown parser
export function parseMarkdown(text) {
  if (!text) return '';

  // Escape HTML
  text = text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');

  // Bold: **text** or __text__
  text = text.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
  text = text.replace(/__(.+?)__/g, '<strong>$1</strong>');

  // Italic: *text* or _text_
  text = text.replace(/\*(.+?)\*/g, '<em>$1</em>');
  text = text.replace(/_(.+?)_/g, '<em>$1</em>');

  // Strikethrough: ~~text~~
  text = text.replace(/~~(.+?)~~/g, '<del>$1</del>');

  // Code: `code`
  text = text.replace(/`(.+?)`/g, '<code style="background: rgba(255,255,255,0.1); padding: 2px 6px; border-radius: 4px;">$1</code>');

  // Links: [text](url)
  text = text.replace(/\[(.+?)\]\((.+?)\)/g, '<a href="$2" target="_blank" style="color: var(--primary); text-decoration: underline;">$1</a>');

  // Line breaks
  text = text.replace(/\n/g, '<br>');

  return text;
}

// Markdown toolbar for textarea
export function createMarkdownToolbar(textareaId) {
  const toolbar = document.createElement('div');
  toolbar.className = 'markdown-toolbar';
  toolbar.style.cssText = `
    display: flex;
    gap: 0.5rem;
    padding: 0.5rem;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 8px 8px 0 0;
    border-bottom: 1px solid var(--glass-border);
  `;

  const buttons = [
    { icon: 'B', title: 'Bold', wrap: '**' },
    { icon: 'I', title: 'Italic', wrap: '*' },
    { icon: 'S', title: 'Strikethrough', wrap: '~~' },
    { icon: '<>', title: 'Code', wrap: '`' },
    { icon: '🔗', title: 'Link', link: true },
  ];

  buttons.forEach(btn => {
    const button = document.createElement('button');
    button.type = 'button';
    button.textContent = btn.icon;
    button.title = btn.title;
    button.className = 'markdown-btn';
    button.style.cssText = `
      padding: 0.25rem 0.5rem;
      background: rgba(255, 255, 255, 0.1);
      border: 1px solid rgba(255, 255, 255, 0.1);
      border-radius: 6px;
      color: var(--text-primary);
      cursor: pointer;
      transition: all 0.2s;
      font-weight: ${btn.icon === 'B' ? 'bold' : btn.icon === 'I' ? 'italic' : 'normal'};
    `;

    button.addEventListener('mouseover', () => {
      button.style.background = 'rgba(255, 255, 255, 0.2)';
    });

    button.addEventListener('mouseout', () => {
      button.style.background = 'rgba(255, 255, 255, 0.1)';
    });

    button.addEventListener('click', () => {
      const textarea = document.getElementById(textareaId);
      if (!textarea) return;

      const start = textarea.selectionStart;
      const end = textarea.selectionEnd;
      const selectedText = textarea.value.substring(start, end);

      let newText;
      if (btn.link) {
        const url = prompt('Enter URL:');
        if (url) {
          newText = `[${selectedText || 'link text'}](${url})`;
        } else {
          return;
        }
      } else {
        newText = `${btn.wrap}${selectedText || 'text'}${btn.wrap}`;
      }

      textarea.value = textarea.value.substring(0, start) + newText + textarea.value.substring(end);
      textarea.focus();
      textarea.selectionStart = start + btn.wrap.length;
      textarea.selectionEnd = start + newText.length - btn.wrap.length;
    });

    toolbar.appendChild(button);
  });

  // Preview button
  const previewBtn = document.createElement('button');
  previewBtn.type = 'button';
  previewBtn.textContent = '👁';
  previewBtn.title = 'Preview';
  previewBtn.className = 'markdown-btn';
  previewBtn.style.cssText = toolbar.firstChild.style.cssText;
  previewBtn.style.marginLeft = 'auto';

  let isPreview = false;
  previewBtn.addEventListener('click', () => {
    const textarea = document.getElementById(textareaId);
    if (!textarea) return;

    isPreview = !isPreview;
    if (isPreview) {
      const preview = document.createElement('div');
      preview.id = textareaId + '-preview';
      preview.className = 'markdown-preview';
      preview.style.cssText = textarea.style.cssText;
      preview.style.padding = '0.875rem 1rem';
      preview.innerHTML = parseMarkdown(textarea.value);
      textarea.parentNode.insertBefore(preview, textarea);
      textarea.style.display = 'none';
      previewBtn.style.background = 'var(--primary)';
    } else {
      const preview = document.getElementById(textareaId + '-preview');
      if (preview) preview.remove();
      textarea.style.display = '';
      previewBtn.style.background = 'rgba(255, 255, 255, 0.1)';
    }
  });

  toolbar.appendChild(previewBtn);

  return toolbar;
}
