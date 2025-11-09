# 🚂 Railway Deployment - Покрокова інструкція

## Крок 1: Логін в Railway

Виконайте в терміналі:

```bash
cd "/Users/olehkaminskyi/Desktop/Платформа"
railway login
```

Відкриється браузер для авторизації.

## Крок 2: Створення проекту

```bash
# Ініціалізація нового проекту
railway init

# Введіть назву проекту: psycho-platform
```

## Крок 3: Додавання баз даних

```bash
# Додати PostgreSQL
railway add --plugin postgresql

# Додати Redis
railway add --plugin redis
```

## Крок 4: Налаштування змінних оточення

```bash
# Встановити змінні через CLI
railway variables set JWT_SECRET=$(openssl rand -base64 32)
railway variables set ENVIRONMENT=production
railway variables set FRONTEND_URL=https://psycho-platform.up.railway.app
```

Або через веб-інтерфейс Railway:
1. Відкрийте https://railway.app/dashboard
2. Виберіть проект
3. Settings → Variables
4. Додайте:
   - `JWT_SECRET` = (згенеруйте: `openssl rand -base64 32`)
   - `ENVIRONMENT` = `production`
   - `FRONTEND_URL` = URL вашого додатку

## Крок 5: Деплой

```bash
# Деплой на Railway
railway up

# Або з Git
git push railway main
```

## Крок 6: Перегляд логів

```bash
# Дивитись логи
railway logs

# Дивитись останні 100 рядків
railway logs --tail 100
```

## Крок 7: Відкрити додаток

```bash
# Відкрити в браузері
railway open
```

## Альтернативний метод: Через GitHub

1. Створіть GitHub репозиторій
2. Запуште код:
   ```bash
   git remote add origin https://github.com/your-username/psycho-platform.git
   git push -u origin main
   ```
3. В Railway Dashboard:
   - New Project
   - Deploy from GitHub repo
   - Виберіть репозиторій
   - Railway автоматично додасть PostgreSQL та Redis

## Перевірка після деплою

```bash
# Перевірити health
curl https://your-app.up.railway.app/health

# Повинно повернути:
{
  "status": "healthy",
  "database": "healthy",
  "redis": "healthy"
}
```

## Налаштування домену (опціонально)

1. Railway Dashboard → Settings → Domains
2. Generate Domain або додайте Custom Domain
3. Оновіть `FRONTEND_URL` змінну

## Troubleshooting

### Помилка підключення до БД
- Перевірте що PostgreSQL plugin додано
- Railway автоматично встановлює `DATABASE_URL`

### Помилка підключення до Redis
- Перевірте що Redis plugin додано
- Railway автоматично встановлює `REDIS_URL`

### Додаток не запускається
```bash
railway logs --tail 100
```

Шукайте помилки в логах.

## Готово! 🎉

Ваша платформа тепер в production на Railway!
