# WTE

[![CI](https://github.com/wtepcorp/WTE/actions/workflows/ci.yml/badge.svg)](https://github.com/wtepcorp/WTE/actions/workflows/ci.yml)
[![Release](https://github.com/wtepcorp/WTE/actions/workflows/release.yml/badge.svg)](https://github.com/wtepcorp/WTE/actions/workflows/release.yml)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

**Window to Europe** — CLI-инструмент для управления прокси-инфраструктурой на базе GOST. Позволяет легко устанавливать, настраивать и управлять HTTP, HTTPS и Shadowsocks прокси-серверами на Linux.

## Возможности

- **HTTP Proxy** — Классический прокси для браузеров и приложений
- **HTTPS Proxy** — Прокси с TLS-шифрованием
- **Shadowsocks** — Протокол для обхода ограничений
- **Гибкая настройка** — Включение/отключение аутентификации, выбор портов
- **Автоматическая настройка** — Файрвол, systemd, генерация паролей
- **Простое управление** — start, stop, restart, status, logs

---

## Быстрая установка

```bash
curl -sfL https://raw.githubusercontent.com/wtepcorp/WTE/main/install.sh | sudo bash
```

## Ручная установка

### Вариант 1: Скачать бинарник

```bash
# Скачать последнюю версию
wget https://github.com/wtepcorp/WTE/releases/latest/download/wte-linux-amd64.tar.gz

# Распаковать
tar -xzf wte-linux-amd64.tar.gz

# Установить
sudo mv wte-linux-amd64 /usr/local/bin/wte
sudo chmod +x /usr/local/bin/wte

# Проверить
wte version
```

### Вариант 2: Сборка из исходников

```bash
# Установить Go (если не установлен)
sudo apt update && sudo apt install -y golang-go

# Клонировать репозиторий
git clone https://github.com/wtepcorp/WTE.git
cd WTE

# Собрать и установить
make build
sudo make install

# Проверить
wte version
```

---

## Быстрый старт

### Установка прокси-сервера

```bash
# Базовая установка (HTTP + Shadowsocks с автогенерацией паролей)
sudo wte install

# Только HTTP прокси (без Shadowsocks)
sudo wte install --ss-enabled=false

# Установка без аутентификации
sudo wte install --http-no-auth

# Пользовательские настройки
sudo wte install \
    --http-port 3128 \
    --http-user admin \
    --http-pass mypassword \
    --ss-port 8388
```

После установки будут показаны данные для подключения.

### Управление сервисом

```bash
# Проверить статус
sudo wte status

# Остановить
sudo wte stop

# Запустить
sudo wte start

# Перезапустить
sudo wte restart

# Просмотр логов
sudo wte logs

# Следить за логами в реальном времени
sudo wte logs -f
```

### Просмотр учётных данных

```bash
# Показать все данные для подключения
sudo wte credentials

# Показать только Shadowsocks URI (для импорта)
sudo wte credentials --uri

# Перегенерировать пароли
sudo wte credentials --regenerate
```

### Управление конфигурацией

```bash
# Показать текущую конфигурацию
wte config show

# Изменить порт HTTP прокси
sudo wte config set http.port 3128

# Отключить аутентификацию
sudo wte config set http.auth.enabled false

# Включить Shadowsocks
sudo wte config set shadowsocks.enabled true

# Применить изменения (перегенерировать конфиг и перезапустить)
sudo wte config apply

# Открыть конфиг в редакторе
sudo wte config edit

# Сбросить к настройкам по умолчанию
sudo wte config reset
```

### Обновление WTE

```bash
# Проверить наличие обновлений
sudo wte update --check

# Обновить до последней версии
sudo wte update

# Принудительно переустановить
sudo wte update --force
```

### Удаление

```bash
# Полное удаление
sudo wte uninstall

# Удаление без подтверждения
sudo wte uninstall --force

# Удалить, но сохранить файл с учётными данными
sudo wte uninstall --keep-creds
```

---

## Параметры установки

| Флаг | Описание | По умолчанию |
|------|----------|--------------|
| `--http-port` | Порт HTTP прокси | 8080 |
| `--http-user` | Имя пользователя | proxyuser |
| `--http-pass` | Пароль (автогенерация если пусто) | — |
| `--http-no-auth` | Отключить аутентификацию | false |
| `--ss-enabled` | Включить Shadowsocks | true |
| `--ss-port` | Порт Shadowsocks | 9500 |
| `--ss-password` | Пароль SS (автогенерация если пусто) | — |
| `--ss-method` | Метод шифрования | aes-128-gcm |
| `--https-enabled` | Включить HTTPS прокси | false |
| `--https-port` | Порт HTTPS прокси | 8443 |
| `--skip-firewall` | Не настраивать файрвол | false |
| `--gost-version` | Версия GOST | 3.0.0-rc10 |

---

## Подключение к прокси

### HTTP Proxy

**Браузер / Системные настройки:**
```
Хост: <IP сервера>
Порт: 8080
Логин: proxyuser
Пароль: <ваш пароль>
```

**curl:**
```bash
curl -x http://proxyuser:PASSWORD@SERVER_IP:8080 https://ifconfig.me
```

**Переменные окружения:**
```bash
export http_proxy="http://proxyuser:PASSWORD@SERVER_IP:8080"
export https_proxy="http://proxyuser:PASSWORD@SERVER_IP:8080"
```

### Shadowsocks

**Настройки клиента:**
```
Сервер: <IP сервера>
Порт: 9500
Пароль: <ваш пароль>
Шифрование: aes-128-gcm
```

**Клиенты:**
- **iOS:** Shadowrocket, Surge, Quantumult
- **Android:** Shadowsocks, v2rayNG
- **Windows:** Shadowsocks-windows, v2rayN
- **macOS:** ShadowsocksX-NG, Surge
- **Linux:** shadowsocks-libev, shadowsocks-rust

**Импорт по URI:**
```bash
# Получить URI для импорта
sudo wte credentials --uri
# Пример: ss://YWVzLTEyOC1nY206cGFzc3dvcmQ=@1.2.3.4:9500#WTE-Proxy
```

---

## Расположение файлов

| Файл | Описание |
|------|----------|
| `/usr/local/bin/gost` | Бинарник GOST |
| `/etc/wte/config.yaml` | Конфигурация WTE |
| `/etc/gost/config.yaml` | Конфигурация GOST |
| `/etc/systemd/system/gost.service` | Systemd сервис |
| `/root/proxy-credentials.txt` | Файл с учётными данными |

---

## Примеры использования

### Пример 1: Простой прокси для личного использования

```bash
sudo wte install
sudo wte credentials
```

### Пример 2: Публичный прокси без аутентификации

```bash
# Внимание: не рекомендуется для публичных серверов!
sudo wte install --http-no-auth --ss-enabled=false
```

### Пример 3: Только Shadowsocks на нестандартном порту

```bash
sudo wte install \
    --http-port 0 \
    --ss-enabled=true \
    --ss-port 443 \
    --ss-method chacha20-ietf-poly1305
```

### Пример 4: Корпоративный прокси с HTTPS

```bash
sudo wte install \
    --http-port 3128 \
    --http-user corporate \
    --http-pass SecurePass123 \
    --https-enabled \
    --https-port 3129 \
    --ss-enabled=false
```

---

## Устранение неполадок

### Сервис не запускается

```bash
# Проверить статус
sudo systemctl status gost

# Посмотреть логи
sudo journalctl -u gost -n 50

# Проверить конфигурацию
cat /etc/gost/config.yaml
```

### Порт уже занят

```bash
# Проверить что занимает порт
sudo ss -tlnp | grep 8080

# Использовать другой порт
sudo wte config set http.port 3128
sudo wte config apply
```

### Не удаётся подключиться снаружи

```bash
# Проверить файрвол
sudo ufw status
# или
sudo firewall-cmd --list-all

# Открыть порт вручную (UFW)
sudo ufw allow 8080/tcp
sudo ufw allow 9500/tcp
sudo ufw allow 9500/udp
```

### Сброс и переустановка

```bash
sudo wte uninstall --force
sudo wte install
```

---

## Глобальные флаги

| Флаг | Описание |
|------|----------|
| `-c, --config` | Путь к файлу конфигурации |
| `-v, --verbose` | Подробный вывод |
| `-q, --quiet` | Минимальный вывод (только ошибки) |
| `--no-color` | Отключить цветной вывод |
| `-h, --help` | Показать справку |

---

## Требования

- **ОС:** Ubuntu 18.04+, Debian 10+, CentOS 7+, Fedora 38+, Arch Linux
- **Архитектура:** x86_64 (amd64), ARM64, ARMv7
- **Права:** root (sudo)
- **Сеть:** Доступ к GitHub для скачивания GOST

---

## Участие в разработке

Мы рады вашим контрибуциям! Пожалуйста, создавайте Pull Request'ы.

1. Сделайте форк репозитория
2. Создайте ветку для функционала (`git checkout -b feature/amazing-feature`)
3. Закоммитьте изменения (`git commit -m 'Add some amazing feature'`)
4. Запушьте в ветку (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

---

## Лицензия

MIT License — подробности в файле [LICENSE](LICENSE).
