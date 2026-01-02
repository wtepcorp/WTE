# Руководство по релизам

## Обзор процесса

```
┌─────────────────────────────────────────────────────────────────┐
│                     ПРОЦЕСС РЕЛИЗА                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   1. Внесение изменений → 2. Обновление версии → 3. Создание тега │
│                                │                                │
│                                ▼                                │
│   4. Пуш тега → 5. Сборка GitHub Actions → 6. Релиз готов       │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Семантическое версионирование

Используйте формат `vMAJOR.MINOR.PATCH`:

| Тип изменения | Когда увеличивать | Пример |
|---------------|-------------------|--------|
| **MAJOR** | Ломающие изменения API | v1.0.0 → v2.0.0 |
| **MINOR** | Новые функции (обратная совместимость) | v1.0.0 → v1.1.0 |
| **PATCH** | Исправление ошибок | v1.0.0 → v1.0.1 |

**Примеры:**
- Добавлена новая команда `wte backup` → **MINOR** (v1.1.0)
- Исправлена ошибка в `wte install` → **PATCH** (v1.0.1)
- Изменён формат конфига (старые конфиги не работают) → **MAJOR** (v2.0.0)

---

## Пошаговая инструкция релиза

### Шаг 1: Убедитесь, что код готов

```bash
# Убедитесь, что вы на ветке main
git checkout main
git pull origin main

# Запустите тесты
make test

# Проверьте, что собирается
make build

# Запустите линтер
make lint
```

### Шаг 2: Обновите версию в коде

Отредактируйте файл `internal/cli/root.go`:

```go
var (
    Version   = "1.1.0"  // ← Обновите версию
    BuildTime = "unknown"
    GitCommit = "unknown"
)
```

### Шаг 3: Обновите CHANGELOG (опционально)

Если ведёте CHANGELOG.md, добавьте секцию:

```markdown
## [1.1.0] - 2024-01-15

### Добавлено
- Новая команда `wte backup` для резервного копирования

### Исправлено
- Исправлена ошибка установки на ARM64

### Изменено
- Улучшен вывод `wte status`
```

### Шаг 4: Закоммитьте изменения

```bash
git add -A
git commit -m "chore: bump version to v1.1.0"
```

### Шаг 5: Создайте тег

```bash
# Создать аннотированный тег
git tag -a v1.1.0 -m "Release v1.1.0"

# Или с описанием изменений
git tag -a v1.1.0 -m "Release v1.1.0

- Добавлена команда wte backup
- Исправлена установка на ARM64
- Улучшен вывод статуса"
```

### Шаг 6: Запушьте коммит и тег

```bash
# Запушить коммит
git push origin main

# Запушить тег (запускает GitHub Actions)
git push origin v1.1.0

# Или запушить все теги
git push --tags
```

### Шаг 7: Проверьте релиз

1. Перейдите на GitHub → Releases
2. Проверьте, что workflow запустился (Actions → Release)
3. Дождитесь завершения (обычно 2-3 минуты)
4. Проверьте, что бинарники загружены

---

## Альтернатива: Использование GoReleaser

Если предпочитаете GoReleaser (более мощный инструмент):

### Установка GoReleaser

```bash
# macOS
brew install goreleaser

# Linux
go install github.com/goreleaser/goreleaser@latest

# Или скачать бинарник
curl -sfL https://goreleaser.com/static/run | bash
```

### Локальная сборка (без релиза)

```bash
# Собрать без публикации
goreleaser build --snapshot --clean
```

### Релиз через GoReleaser

```bash
# Создать тег
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0

# GoReleaser запустится автоматически через GitHub Actions
# Или вручную:
export GITHUB_TOKEN="your_github_token"
goreleaser release --clean
```

---

## Настройка репозитория на GitHub

### 1. Включите GitHub Actions

1. GitHub → Settings → Actions → General
2. Выберите "Allow all actions"
3. В секции "Workflow permissions" выберите "Read and write permissions"

### 2. Создайте первый релиз

```bash
git tag -a v1.0.0 -m "Initial release"
git push origin v1.0.0
```

---

## Структура релиза

После успешного релиза на GitHub будет:

```
Release v1.1.0
├── wte-linux-amd64.tar.gz      # Для стандартных серверов
├── wte-linux-arm64.tar.gz      # Для ARM серверов (AWS Graviton и т.д.)
├── wte-linux-armv7.tar.gz      # Для Raspberry Pi
└── checksums.txt                # SHA256 контрольные суммы
```

---

## Механизм обновления для пользователей

### Автоматическое обновление

Пользователи с установленным WTE могут обновиться командой:

```bash
# Проверить наличие обновлений
sudo wte update --check

# Обновить до последней версии
sudo wte update

# Принудительно переустановить текущую версию
sudo wte update --force
```

### Ручное обновление

```bash
# Скачать новую версию
wget https://github.com/wtepcorp/WTE/releases/latest/download/wte-linux-amd64.tar.gz

# Распаковать
tar -xzf wte-linux-amd64.tar.gz

# Заменить бинарник
sudo mv wte-linux-amd64 /usr/local/bin/wte
sudo chmod +x /usr/local/bin/wte

# Проверить версию
wte version
```

### Обновление через скрипт

Предоставьте пользователям скрипт:

```bash
curl -sfL https://raw.githubusercontent.com/wtepcorp/WTE/main/install.sh | sudo bash
```

---

## Чеклист перед релизом

- [ ] Все тесты проходят (`make test`)
- [ ] Код компилируется (`make build`)
- [ ] Линтер проходит (`make lint`)
- [ ] Версия обновлена в `internal/cli/root.go`
- [ ] CHANGELOG обновлён (если ведётся)
- [ ] Коммит запушен в main
- [ ] Тег создан и запушен
- [ ] GitHub Actions завершился успешно
- [ ] Бинарники доступны на странице Releases

---

## Откат релиза

Если что-то пошло не так:

```bash
# Удалить тег локально
git tag -d v1.1.0

# Удалить тег на GitHub
git push origin :refs/tags/v1.1.0

# Удалить релиз вручную на GitHub
# Settings → Releases → Delete
```

---

## Часто задаваемые вопросы

**В: Как часто выпускать релизы?**
О: По мере необходимости. Накапливайте мелкие исправления, критические баги выпускайте сразу.

**В: Нужно ли тестировать на всех платформах?**
О: Желательно протестировать на основной платформе (linux/amd64). ARM можно проверить в CI.

**В: Что если забыл обновить версию в коде?**
О: GitHub Actions подставит версию из тега через ldflags. Но лучше обновлять для консистентности.

**В: Можно ли выпустить пре-релиз?**
О: Да, используйте теги вида `v1.1.0-beta.1` или `v1.1.0-rc.1`.
