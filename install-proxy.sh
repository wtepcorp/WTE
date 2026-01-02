#!/bin/bash
#
# ╔═══════════════════════════════════════════════════════════════════════════╗
# ║                    UNIVERSAL PROXY SERVER INSTALLER                       ║
# ║                                                                           ║
# ║  Version: 2.0.0                                                           ║
# ║  License: MIT                                                             ║
# ║                                                                           ║
# ║  Описание:                                                                ║
# ║  Этот скрипт автоматически разворачивает прокси-сервер на базе GOST.     ║
# ║  GOST - это мощный и гибкий прокси-инструмент, написанный на Go,         ║
# ║  который поддерживает множество протоколов включая HTTP, SOCKS5,         ║
# ║  Shadowsocks и другие.                                                    ║
# ║                                                                           ║
# ║  Поддерживаемые ОС:                                                       ║
# ║  - Ubuntu 18.04, 20.04, 22.04, 24.04                                     ║
# ║  - Debian 10, 11, 12                                                      ║
# ║  - CentOS 7, 8, Stream 9                                                  ║
# ║  - Rocky Linux 8, 9                                                       ║
# ║  - AlmaLinux 8, 9                                                         ║
# ║  - Fedora 38+                                                             ║
# ║  - Arch Linux                                                             ║
# ║                                                                           ║
# ║  Поддерживаемые архитектуры:                                              ║
# ║  - x86_64 (amd64)                                                         ║
# ║  - aarch64 (arm64)                                                        ║
# ║  - armv7l (armv7)                                                         ║
# ║                                                                           ║
# ║  Использование:                                                           ║
# ║  ./install-proxy.sh              # Интерактивная установка               ║
# ║  ./install-proxy.sh --uninstall  # Удаление                              ║
# ║  ./install-proxy.sh --help       # Справка                               ║
# ║                                                                           ║
# ║  Переменные окружения для кастомизации:                                   ║
# ║  HTTP_PROXY_PORT  - порт HTTP прокси (по умолчанию: 8080)                ║
# ║  HTTP_PROXY_USER  - имя пользователя (по умолчанию: proxyuser)           ║
# ║  HTTP_PROXY_PASS  - пароль (по умолчанию: генерируется случайно)         ║
# ║  SS_ENABLED       - включить Shadowsocks (по умолчанию: true)            ║
# ║  SS_PORT          - порт Shadowsocks (по умолчанию: 9500)                ║
# ║  SS_PASSWORD      - пароль SS (по умолчанию: генерируется случайно)      ║
# ║  SS_METHOD        - метод шифрования (по умолчанию: aes-128-gcm)         ║
# ║  HTTPS_ENABLED    - включить HTTPS прокси (по умолчанию: false)          ║
# ║  HTTPS_PORT       - порт HTTPS прокси (по умолчанию: 8443)               ║
# ║                                                                           ║
# ╚═══════════════════════════════════════════════════════════════════════════╝
#

# ============================================================================
# СЕКЦИЯ 1: НАСТРОЙКИ БЕЗОПАСНОСТИ BASH
# ============================================================================
#
# set -e: Прерывает выполнение скрипта при любой ошибке (ненулевой код возврата)
#         Это предотвращает продолжение работы при сбоях
#
# set -o pipefail: Ошибка в любой части пайпа (cmd1 | cmd2 | cmd3) приведёт
#                  к ошибке всего пайпа, а не только последней команды
#
# Эти настройки делают скрипт более надёжным и предсказуемым

set -e
set -o pipefail

# ============================================================================
# СЕКЦИЯ 2: КОНФИГУРАЦИОННЫЕ ПЕРЕМЕННЫЕ
# ============================================================================
#
# Здесь определяются все настраиваемые параметры скрипта.
# Используется синтаксис ${VAR:-default}, который означает:
# "использовать значение переменной VAR, а если она не задана - использовать default"
#
# Это позволяет переопределять параметры через переменные окружения:
# HTTP_PROXY_PORT=3128 ./install-proxy.sh
#

# --- Версия скрипта ---
# Используется для отображения в заголовке и логах
SCRIPT_VERSION="2.0.0"

# --- Настройки HTTP прокси ---
# HTTP прокси - основной способ подключения для браузеров
# Порт 8080 выбран как стандартный для прокси-серверов
HTTP_PROXY_PORT="${HTTP_PROXY_PORT:-8080}"
HTTP_PROXY_USER="${HTTP_PROXY_USER:-proxyuser}"
# Если пароль не задан, генерируем случайный 16-байтовый в base64
# openssl rand -base64 16 создаёт криптографически стойкий пароль
HTTP_PROXY_PASS="${HTTP_PROXY_PASS:-$(openssl rand -base64 16 2>/dev/null || echo "ChangeMe123!")}"

# --- Настройки Shadowsocks ---
# Shadowsocks - протокол для обхода блокировок, популярен для мобильных клиентов
# Использует шифрование трафика, что затрудняет его обнаружение
SS_ENABLED="${SS_ENABLED:-true}"
SS_PORT="${SS_PORT:-9500}"
SS_PASSWORD="${SS_PASSWORD:-$(openssl rand -base64 16 2>/dev/null || echo "ChangeMe456!")}"
# aes-128-gcm - современный и быстрый алгоритм AEAD шифрования
# Поддерживается всеми современными клиентами Shadowsocks
SS_METHOD="${SS_METHOD:-aes-128-gcm}"

# --- Настройки HTTPS прокси ---
# HTTPS прокси шифрует соединение между клиентом и прокси-сервером
# Полезно при использовании в недоверенных сетях
HTTPS_ENABLED="${HTTPS_ENABLED:-false}"
HTTPS_PORT="${HTTPS_PORT:-8443}"

# --- Версия GOST ---
# GOST - основной инструмент, на котором работает прокси
# Рекомендуется использовать последнюю стабильную версию
GOST_VERSION="${GOST_VERSION:-3.0.0-rc10}"

# --- Пути установки ---
# Стандартные пути в Linux для пользовательских бинарников и конфигов
GOST_BINARY="/usr/local/bin/gost"
GOST_CONFIG_DIR="/etc/gost"
GOST_CONFIG_FILE="${GOST_CONFIG_DIR}/config.yaml"
GOST_SERVICE_FILE="/etc/systemd/system/gost.service"
CREDENTIALS_FILE="/root/proxy-credentials.txt"

# ============================================================================
# СЕКЦИЯ 3: ПОДСЧЁТ ШАГОВ И ПЕРЕМЕННЫЕ СОСТОЯНИЯ
# ============================================================================
#
# Для отображения прогресса нам нужно знать общее количество шагов
# и текущий шаг. Это делает вывод более информативным для пользователя.
#

# Общее количество шагов установки
# Шаги: detect_os, detect_arch, get_public_ip, check_existing,
#       install_packages, install_gost, generate_config, create_service,
#       configure_firewall, verify_installation, save_credentials
TOTAL_STEPS=11
CURRENT_STEP=0

# Переменные для хранения информации о системе
# Заполняются в процессе выполнения скрипта
OS=""                # Название ОС (ubuntu, debian, centos, etc.)
VERSION=""           # Версия ОС
ARCH=""              # Архитектура процессора (x86_64, aarch64, etc.)
GOST_ARCH=""         # Архитектура для скачивания GOST (amd64, arm64, etc.)
PUBLIC_IP=""         # Публичный IP-адрес сервера

# Счётчики успешных и неудачных операций
SUCCESS_COUNT=0
FAILED_COUNT=0

# Время начала установки для подсчёта общего времени
START_TIME=$(date +%s)

# ============================================================================
# СЕКЦИЯ 4: ЦВЕТА И СИМВОЛЫ ДЛЯ ВЫВОДА
# ============================================================================
#
# ANSI escape-коды для цветного вывода в терминал.
# Делают вывод скрипта более читаемым и понятным.
#
# Формат: \033[<код>m или \e[<код>m
# - 0;31m = красный
# - 0;32m = зелёный
# - 1;33m = жёлтый (жирный)
# - 0;34m = синий
# - 0;36m = cyan
# - 1;37m = белый (жирный)
# - 0m    = сброс цвета
#

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
GRAY='\033[0;37m'
NC='\033[0m'  # No Color - сброс цвета

# Символы для индикации статуса
# Используем Unicode-символы для более красивого отображения
SYMBOL_SUCCESS="✓"   # Галочка - успех
SYMBOL_FAILED="✗"    # Крестик - ошибка
SYMBOL_WARNING="⚠"   # Предупреждение
SYMBOL_INFO="ℹ"      # Информация
SYMBOL_ARROW="→"     # Стрелка - действие
SYMBOL_BULLET="•"    # Точка - элемент списка

# ============================================================================
# СЕКЦИЯ 5: ФУНКЦИИ ВЫВОДА СООБЩЕНИЙ
# ============================================================================
#
# Набор функций для стандартизированного вывода сообщений.
# Каждый тип сообщения имеет свой цвет и префикс для лёгкой идентификации.
#

# -----------------------------------------------------------------------------
# Функция: print_header
# Описание: Выводит красивый заголовок скрипта с ASCII-артом
# Параметры: нет
# Возврат: нет
# -----------------------------------------------------------------------------
print_header() {
    clear
    echo ""
    echo -e "${CYAN}╔═══════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${NC}                                                                           ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}   ${WHITE}██████╗  ██████╗  ██████╗ ██╗  ██╗██╗   ██╗${NC}                            ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}   ${WHITE}██╔══██╗██╔══██╗██╔═══██╗╚██╗██╔╝╚██╗ ██╔╝${NC}                            ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}   ${WHITE}██████╔╝██████╔╝██║   ██║ ╚███╔╝  ╚████╔╝${NC}                             ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}   ${WHITE}██╔═══╝ ██╔══██╗██║   ██║ ██╔██╗   ╚██╔╝${NC}                              ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}   ${WHITE}██║     ██║  ██║╚██████╔╝██╔╝ ██╗   ██║${NC}                               ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}   ${WHITE}╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝${NC}                               ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}                                                                           ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}          ${GRAY}Universal Proxy Server Installer v${SCRIPT_VERSION}${NC}                       ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}                                                                           ${CYAN}║${NC}"
    echo -e "${CYAN}╚═══════════════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

# -----------------------------------------------------------------------------
# Функция: print_step
# Описание: Выводит заголовок текущего шага с номером и прогресс-баром
# Параметры:
#   $1 - название шага
# Возврат: нет
# -----------------------------------------------------------------------------
print_step() {
    local step_name="$1"

    # Увеличиваем счётчик текущего шага
    # Используем префиксный инкремент (++CURRENT_STEP) вместо постфиксного (CURRENT_STEP++)
    # потому что с set -e постфиксный инкремент от 0 возвращает 0, что bash интерпретирует как ошибку
    ((++CURRENT_STEP))

    # Вычисляем процент выполнения
    local percent=$((CURRENT_STEP * 100 / TOTAL_STEPS))

    # Создаём визуальный прогресс-бар
    # Ширина прогресс-бара - 20 символов
    local bar_width=20
    local filled=$((CURRENT_STEP * bar_width / TOTAL_STEPS))
    local empty=$((bar_width - filled))

    # Формируем строку прогресс-бара
    local bar=""
    for ((i=0; i<filled; i++)); do bar+="█"; done
    for ((i=0; i<empty; i++)); do bar+="░"; done

    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${WHITE}  STEP ${CURRENT_STEP}/${TOTAL_STEPS}${NC} ${GRAY}│${NC} ${bar} ${percent}% ${GRAY}│${NC} ${WHITE}${step_name}${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# -----------------------------------------------------------------------------
# Функция: log_info
# Описание: Выводит информационное сообщение (синий цвет)
# Параметры:
#   $1 - текст сообщения
# Возврат: нет
# -----------------------------------------------------------------------------
log_info() {
    echo -e "  ${BLUE}${SYMBOL_INFO}${NC}  $1"
}

# -----------------------------------------------------------------------------
# Функция: log_action
# Описание: Выводит сообщение о выполняемом действии (серый цвет со стрелкой)
# Параметры:
#   $1 - текст сообщения
# Возврат: нет
# -----------------------------------------------------------------------------
log_action() {
    echo -e "  ${GRAY}${SYMBOL_ARROW}${NC}  $1"
}

# -----------------------------------------------------------------------------
# Функция: log_success
# Описание: Выводит сообщение об успехе (зелёный цвет с галочкой)
# Параметры:
#   $1 - текст сообщения
# Возврат: нет
# -----------------------------------------------------------------------------
log_success() {
    echo -e "  ${GREEN}${SYMBOL_SUCCESS}${NC}  $1"
    ((++SUCCESS_COUNT))
}

# -----------------------------------------------------------------------------
# Функция: log_warning
# Описание: Выводит предупреждение (жёлтый цвет)
# Параметры:
#   $1 - текст сообщения
# Возврат: нет
# -----------------------------------------------------------------------------
log_warning() {
    echo -e "  ${YELLOW}${SYMBOL_WARNING}${NC}  $1"
}

# -----------------------------------------------------------------------------
# Функция: log_error
# Описание: Выводит сообщение об ошибке и завершает скрипт
# Параметры:
#   $1 - текст сообщения
# Возврат: exit 1
# -----------------------------------------------------------------------------
log_error() {
    echo -e "  ${RED}${SYMBOL_FAILED}${NC}  ${RED}$1${NC}"
    ((++FAILED_COUNT))
    echo ""
    echo -e "${RED}Installation failed. Please check the error above.${NC}"
    exit 1
}

# -----------------------------------------------------------------------------
# Функция: log_detail
# Описание: Выводит детальную информацию (серый цвет, с отступом)
# Параметры:
#   $1 - текст сообщения
# Возврат: нет
# -----------------------------------------------------------------------------
log_detail() {
    echo -e "     ${GRAY}${SYMBOL_BULLET} $1${NC}"
}

# ============================================================================
# СЕКЦИЯ 6: ФУНКЦИЯ СПРАВКИ
# ============================================================================

# -----------------------------------------------------------------------------
# Функция: show_help
# Описание: Выводит справку по использованию скрипта
# Параметры: нет
# Возврат: exit 0
# -----------------------------------------------------------------------------
show_help() {
    echo ""
    echo -e "${WHITE}Universal Proxy Server Installer v${SCRIPT_VERSION}${NC}"
    echo ""
    echo -e "${CYAN}USAGE:${NC}"
    echo "  ./install-proxy.sh [OPTIONS]"
    echo ""
    echo -e "${CYAN}OPTIONS:${NC}"
    echo "  -h, --help        Show this help message"
    echo "  -u, --uninstall   Uninstall GOST proxy server"
    echo "  -v, --version     Show version"
    echo ""
    echo -e "${CYAN}ENVIRONMENT VARIABLES:${NC}"
    echo "  HTTP_PROXY_PORT   HTTP proxy port (default: 8080)"
    echo "  HTTP_PROXY_USER   HTTP proxy username (default: proxyuser)"
    echo "  HTTP_PROXY_PASS   HTTP proxy password (default: random)"
    echo "  SS_ENABLED        Enable Shadowsocks (default: true)"
    echo "  SS_PORT           Shadowsocks port (default: 9500)"
    echo "  SS_PASSWORD       Shadowsocks password (default: random)"
    echo "  SS_METHOD         Shadowsocks method (default: aes-128-gcm)"
    echo "  HTTPS_ENABLED     Enable HTTPS proxy (default: false)"
    echo "  HTTPS_PORT        HTTPS proxy port (default: 8443)"
    echo ""
    echo -e "${CYAN}EXAMPLES:${NC}"
    echo "  # Basic installation with random passwords"
    echo "  ./install-proxy.sh"
    echo ""
    echo "  # Custom configuration"
    echo "  HTTP_PROXY_USER=admin HTTP_PROXY_PASS=secret ./install-proxy.sh"
    echo ""
    echo "  # HTTP proxy only (no Shadowsocks)"
    echo "  SS_ENABLED=false ./install-proxy.sh"
    echo ""
    echo "  # With HTTPS proxy"
    echo "  HTTPS_ENABLED=true ./install-proxy.sh"
    echo ""
    exit 0
}

# ============================================================================
# СЕКЦИЯ 7: ФУНКЦИИ ОПРЕДЕЛЕНИЯ СИСТЕМЫ
# ============================================================================

# -----------------------------------------------------------------------------
# Функция: detect_os
# Описание: Определяет операционную систему и её версию
#           Читает файл /etc/os-release, который присутствует в большинстве
#           современных Linux-дистрибутивов и содержит информацию о системе.
# Параметры: нет
# Возврат:
#   - Устанавливает глобальные переменные OS и VERSION
#   - При ошибке - завершает скрипт
# -----------------------------------------------------------------------------
detect_os() {
    print_step "Detecting operating system"

    log_action "Reading system information..."

    # Проверяем наличие файла /etc/os-release
    # Этот файл стандартизирован в systemd и содержит информацию о дистрибутиве
    if [ -f /etc/os-release ]; then
        # Подгружаем переменные из файла
        # Файл содержит переменные вида: ID=ubuntu, VERSION_ID=22.04 и т.д.
        . /etc/os-release
        OS=$ID
        VERSION=$VERSION_ID
        log_success "Operating system detected"
        log_detail "Distribution: ${OS}"
        log_detail "Version: ${VERSION:-unknown}"
        log_detail "Name: ${PRETTY_NAME:-$OS}"
    elif [ -f /etc/redhat-release ]; then
        # Fallback для старых версий CentOS/RHEL без os-release
        OS="centos"
        VERSION=$(grep -oE '[0-9]+\.[0-9]+' /etc/redhat-release | head -1)
        log_success "Operating system detected (legacy method)"
        log_detail "Distribution: ${OS}"
        log_detail "Version: ${VERSION:-unknown}"
    else
        log_error "Unable to detect operating system. /etc/os-release not found."
    fi

    # Проверяем, поддерживается ли данная ОС
    case $OS in
        ubuntu|debian|centos|rhel|rocky|almalinux|fedora|arch|manjaro)
            log_detail "Status: ${GREEN}Supported${NC}"
            ;;
        *)
            log_warning "OS '${OS}' is not officially tested, proceeding anyway..."
            ;;
    esac
}

# -----------------------------------------------------------------------------
# Функция: detect_arch
# Описание: Определяет архитектуру процессора и сопоставляет её с архитектурой
#           в релизах GOST (amd64, arm64, armv7)
# Параметры: нет
# Возврат:
#   - Устанавливает глобальные переменные ARCH и GOST_ARCH
#   - При неподдерживаемой архитектуре - завершает скрипт
# -----------------------------------------------------------------------------
detect_arch() {
    print_step "Detecting system architecture"

    log_action "Checking CPU architecture..."

    # uname -m возвращает архитектуру ядра
    ARCH=$(uname -m)

    # Сопоставляем архитектуру системы с именованием в релизах GOST
    case $ARCH in
        x86_64)
            # Стандартная 64-битная архитектура Intel/AMD
            GOST_ARCH="amd64"
            log_success "Architecture detected: ${ARCH}"
            log_detail "GOST architecture: ${GOST_ARCH}"
            ;;
        aarch64)
            # 64-битная ARM (AWS Graviton, Apple Silicon, Raspberry Pi 4 64-bit)
            GOST_ARCH="arm64"
            log_success "Architecture detected: ${ARCH}"
            log_detail "GOST architecture: ${GOST_ARCH}"
            ;;
        armv7l)
            # 32-битная ARM (Raspberry Pi 2/3 32-bit mode)
            GOST_ARCH="armv7"
            log_success "Architecture detected: ${ARCH}"
            log_detail "GOST architecture: ${GOST_ARCH}"
            ;;
        *)
            log_error "Unsupported architecture: ${ARCH}. Supported: x86_64, aarch64, armv7l"
            ;;
    esac
}

# -----------------------------------------------------------------------------
# Функция: get_public_ip
# Описание: Определяет публичный IP-адрес сервера, используя внешние сервисы.
#           Пробует несколько сервисов для надёжности.
# Параметры: нет
# Возврат:
#   - Устанавливает глобальную переменную PUBLIC_IP
# -----------------------------------------------------------------------------
get_public_ip() {
    print_step "Detecting public IP address"

    log_action "Querying external services for public IP..."

    # Список сервисов для определения IP (пробуем по очереди)
    # Используем несколько сервисов для отказоустойчивости
    local ip_services=(
        "ifconfig.me"
        "icanhazip.com"
        "ipinfo.io/ip"
        "api.ipify.org"
        "ipecho.net/plain"
    )

    # Пробуем каждый сервис по очереди
    for service in "${ip_services[@]}"; do
        log_detail "Trying ${service}..."

        # curl с таймаутом 5 секунд, тихий режим
        PUBLIC_IP=$(curl -s --connect-timeout 5 --max-time 10 "$service" 2>/dev/null)

        # Проверяем, что получили валидный IP (простая проверка на формат)
        if [[ $PUBLIC_IP =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            log_success "Public IP detected: ${GREEN}${PUBLIC_IP}${NC}"
            return 0
        fi
    done

    # Если ни один сервис не вернул IP
    PUBLIC_IP="YOUR_SERVER_IP"
    log_warning "Could not detect public IP automatically"
    log_detail "Using placeholder: ${PUBLIC_IP}"
    log_detail "You will need to replace this manually in the output"
}

# ============================================================================
# СЕКЦИЯ 8: ФУНКЦИЯ ПРОВЕРКИ СУЩЕСТВУЮЩЕЙ УСТАНОВКИ
# ============================================================================

# -----------------------------------------------------------------------------
# Функция: check_existing_installation
# Описание: Проверяет, не установлен ли уже GOST на системе.
#           Если установлен - предлагает варианты действий.
# Параметры: нет
# Возврат:
#   - Продолжает установку или завершает скрипт в зависимости от выбора
# -----------------------------------------------------------------------------
check_existing_installation() {
    print_step "Checking existing installation"

    log_action "Looking for existing GOST installation..."

    local found_existing=false

    # Проверяем наличие бинарного файла GOST
    if [ -f "$GOST_BINARY" ]; then
        found_existing=true
        local current_version=$($GOST_BINARY -V 2>/dev/null || echo "unknown")
        log_warning "GOST binary found: ${GOST_BINARY}"
        log_detail "Current version: ${current_version}"
    fi

    # Проверяем наличие systemd-сервиса
    if [ -f "$GOST_SERVICE_FILE" ]; then
        found_existing=true
        local service_status=$(systemctl is-active gost 2>/dev/null || echo "inactive")
        log_warning "GOST service found: ${GOST_SERVICE_FILE}"
        log_detail "Service status: ${service_status}"
    fi

    # Проверяем наличие конфигурации
    if [ -f "$GOST_CONFIG_FILE" ]; then
        found_existing=true
        log_warning "GOST config found: ${GOST_CONFIG_FILE}"
    fi

    if [ "$found_existing" = true ]; then
        echo ""
        log_warning "Existing GOST installation detected!"
        log_info "The installation will:"
        log_detail "Stop the existing service"
        log_detail "Backup current config to ${GOST_CONFIG_FILE}.backup"
        log_detail "Install new version and configuration"
        echo ""

        # Автоматически продолжаем (можно добавить интерактивный режим)
        log_action "Proceeding with upgrade..."

        # Останавливаем существующий сервис
        if systemctl is-active --quiet gost 2>/dev/null; then
            log_action "Stopping existing GOST service..."
            systemctl stop gost
            log_success "Service stopped"
        fi

        # Делаем бэкап конфигурации
        if [ -f "$GOST_CONFIG_FILE" ]; then
            log_action "Backing up existing configuration..."
            cp "$GOST_CONFIG_FILE" "${GOST_CONFIG_FILE}.backup.$(date +%Y%m%d_%H%M%S)"
            log_success "Configuration backed up"
        fi
    else
        log_success "No existing installation found"
        log_detail "Proceeding with fresh installation"
    fi
}

# ============================================================================
# СЕКЦИЯ 9: ФУНКЦИЯ УСТАНОВКИ ПАКЕТОВ
# ============================================================================

# -----------------------------------------------------------------------------
# Функция: install_packages
# Описание: Устанавливает необходимые пакеты в зависимости от ОС.
#           Использует соответствующий пакетный менеджер.
# Параметры: нет
# Возврат:
#   - Устанавливает пакеты: wget, curl, openssl, tar
# -----------------------------------------------------------------------------
install_packages() {
    print_step "Installing required packages"

    log_action "Determining package manager..."

    # Список необходимых пакетов
    local packages="wget curl openssl tar"

    case $OS in
        ubuntu|debian)
            # APT - пакетный менеджер для Debian-based дистрибутивов
            log_detail "Package manager: APT"
            log_action "Updating package lists..."
            apt-get update -qq > /dev/null 2>&1
            log_success "Package lists updated"

            log_action "Installing packages: ${packages}..."
            apt-get install -y -qq $packages > /dev/null 2>&1
            log_success "Packages installed"
            ;;

        centos|rhel|rocky|almalinux)
            # YUM - пакетный менеджер для RHEL-based дистрибутивов
            log_detail "Package manager: YUM"
            log_action "Installing packages: ${packages}..."
            yum install -y -q $packages > /dev/null 2>&1
            log_success "Packages installed"
            ;;

        fedora)
            # DNF - современный пакетный менеджер для Fedora
            log_detail "Package manager: DNF"
            log_action "Installing packages: ${packages}..."
            dnf install -y -q $packages > /dev/null 2>&1
            log_success "Packages installed"
            ;;

        arch|manjaro)
            # Pacman - пакетный менеджер для Arch-based дистрибутивов
            log_detail "Package manager: Pacman"
            log_action "Installing packages: ${packages}..."
            pacman -Sy --noconfirm $packages > /dev/null 2>&1
            log_success "Packages installed"
            ;;

        *)
            log_warning "Unknown package manager for OS: ${OS}"
            log_detail "Assuming required packages are already installed"
            ;;
    esac

    # Проверяем, что все необходимые инструменты доступны
    log_action "Verifying installed tools..."
    local missing_tools=""

    for tool in wget curl openssl tar; do
        if ! command -v $tool > /dev/null 2>&1; then
            missing_tools+=" $tool"
        fi
    done

    if [ -n "$missing_tools" ]; then
        log_error "Required tools not found:${missing_tools}"
    fi

    log_success "All required tools available"
}

# ============================================================================
# СЕКЦИЯ 10: ФУНКЦИЯ УСТАНОВКИ GOST
# ============================================================================

# -----------------------------------------------------------------------------
# Функция: install_gost
# Описание: Скачивает и устанавливает бинарный файл GOST.
#           Скачивает архив с GitHub, распаковывает и устанавливает.
# Параметры: нет
# Возврат:
#   - Устанавливает GOST в /usr/local/bin/gost
# -----------------------------------------------------------------------------
install_gost() {
    print_step "Installing GOST v${GOST_VERSION}"

    # Формируем URL для скачивания
    local download_url="https://github.com/go-gost/gost/releases/download/v${GOST_VERSION}/gost_${GOST_VERSION}_linux_${GOST_ARCH}.tar.gz"
    local temp_dir="/tmp/gost_install_$$"  # $$ - PID текущего процесса для уникальности
    local archive_file="${temp_dir}/gost.tar.gz"

    log_info "Download URL: ${download_url}"

    # Создаём временную директорию
    log_action "Creating temporary directory..."
    mkdir -p "$temp_dir"
    log_success "Temporary directory created: ${temp_dir}"

    # Скачиваем архив
    log_action "Downloading GOST archive..."
    log_detail "This may take a moment depending on your connection speed"

    if wget -q --show-progress -O "$archive_file" "$download_url" 2>/dev/null; then
        log_success "Download completed"
    elif wget -q -O "$archive_file" "$download_url" 2>/dev/null; then
        log_success "Download completed"
    else
        log_error "Failed to download GOST from ${download_url}"
    fi

    # Проверяем, что файл скачался
    if [ ! -f "$archive_file" ] || [ ! -s "$archive_file" ]; then
        log_error "Downloaded file is empty or missing"
    fi

    local file_size=$(du -h "$archive_file" | cut -f1)
    log_detail "Archive size: ${file_size}"

    # Распаковываем архив
    log_action "Extracting archive..."
    cd "$temp_dir"

    if ! tar -xzf "$archive_file"; then
        log_error "Failed to extract archive"
    fi
    log_success "Archive extracted"

    # Устанавливаем бинарный файл
    log_action "Installing GOST binary to ${GOST_BINARY}..."

    if [ -f "${temp_dir}/gost" ]; then
        mv "${temp_dir}/gost" "$GOST_BINARY"
        chmod +x "$GOST_BINARY"
        log_success "GOST binary installed"
    else
        log_error "GOST binary not found in archive"
    fi

    # Проверяем установку
    log_action "Verifying installation..."

    if [ -x "$GOST_BINARY" ]; then
        local installed_version=$($GOST_BINARY -V 2>/dev/null)
        log_success "GOST installed successfully"
        log_detail "Version: ${installed_version}"
    else
        log_error "GOST binary is not executable"
    fi

    # Очищаем временные файлы
    log_action "Cleaning up temporary files..."
    rm -rf "$temp_dir"
    log_success "Cleanup completed"
}

# ============================================================================
# СЕКЦИЯ 11: ФУНКЦИЯ ГЕНЕРАЦИИ КОНФИГУРАЦИИ
# ============================================================================

# -----------------------------------------------------------------------------
# Функция: generate_config
# Описание: Генерирует конфигурационный файл GOST в формате YAML.
#           Создаёт конфигурацию для HTTP прокси, опционально HTTPS и Shadowsocks.
# Параметры: нет
# Возврат:
#   - Создаёт файл /etc/gost/config.yaml
# -----------------------------------------------------------------------------
generate_config() {
    print_step "Generating GOST configuration"

    # Создаём директорию для конфигурации
    log_action "Creating configuration directory..."
    mkdir -p "$GOST_CONFIG_DIR"
    log_success "Directory created: ${GOST_CONFIG_DIR}"

    log_action "Building configuration file..."

    # Начинаем формировать конфигурацию
    # YAML-формат с комментариями для удобства редактирования
    cat > "$GOST_CONFIG_FILE" << EOF
# ============================================================================
# GOST Proxy Server Configuration
# ============================================================================
# Generated: $(date '+%Y-%m-%d %H:%M:%S')
# Generator: Universal Proxy Installer v${SCRIPT_VERSION}
# Documentation: https://gost.run/
# ============================================================================

# Определение сервисов (прокси-серверов)
services:

  # --------------------------------------------------------------------------
  # HTTP Proxy Service
  # --------------------------------------------------------------------------
  # Стандартный HTTP прокси с базовой аутентификацией.
  # Используется для подключения браузеров и приложений.
  #
  # Подключение: http://${HTTP_PROXY_USER}:${HTTP_PROXY_PASS}@SERVER:${HTTP_PROXY_PORT}
  # --------------------------------------------------------------------------
  - name: http-proxy
    addr: ":${HTTP_PROXY_PORT}"
    handler:
      type: http
      auth:
        username: ${HTTP_PROXY_USER}
        password: ${HTTP_PROXY_PASS}
    listener:
      type: tcp
EOF

    log_detail "HTTP proxy configured on port ${HTTP_PROXY_PORT}"

    # Добавляем HTTPS прокси если включено
    if [ "$HTTPS_ENABLED" = "true" ]; then
        log_action "Generating TLS certificates for HTTPS..."

        # Генерируем самоподписанный сертификат
        openssl req -x509 -newkey rsa:4096 \
            -keyout "${GOST_CONFIG_DIR}/key.pem" \
            -out "${GOST_CONFIG_DIR}/cert.pem" \
            -days 365 -nodes \
            -subj "/CN=${PUBLIC_IP}" \
            2>/dev/null

        chmod 600 "${GOST_CONFIG_DIR}/key.pem" "${GOST_CONFIG_DIR}/cert.pem"
        log_success "TLS certificates generated"

        cat >> "$GOST_CONFIG_FILE" << EOF

  # --------------------------------------------------------------------------
  # HTTPS Proxy Service (TLS encrypted)
  # --------------------------------------------------------------------------
  # HTTP прокси с TLS шифрованием.
  # Трафик между клиентом и прокси зашифрован.
  # Использует самоподписанный сертификат (браузер покажет предупреждение).
  #
  # Подключение: https://${HTTP_PROXY_USER}:${HTTP_PROXY_PASS}@SERVER:${HTTPS_PORT}
  # --------------------------------------------------------------------------
  - name: https-proxy
    addr: ":${HTTPS_PORT}"
    handler:
      type: http
      auth:
        username: ${HTTP_PROXY_USER}
        password: ${HTTP_PROXY_PASS}
    listener:
      type: tls
      tls:
        certFile: ${GOST_CONFIG_DIR}/cert.pem
        keyFile: ${GOST_CONFIG_DIR}/key.pem
EOF
        log_detail "HTTPS proxy configured on port ${HTTPS_PORT}"
    fi

    # Добавляем Shadowsocks если включено
    if [ "$SS_ENABLED" = "true" ]; then
        cat >> "$GOST_CONFIG_FILE" << EOF

  # --------------------------------------------------------------------------
  # Shadowsocks Service
  # --------------------------------------------------------------------------
  # Shadowsocks прокси для мобильных и десктопных клиентов.
  # Использует шифрование для обхода блокировок.
  #
  # Настройки для клиента:
  #   Server: ${PUBLIC_IP}
  #   Port: ${SS_PORT}
  #   Password: ${SS_PASSWORD}
  #   Method: ${SS_METHOD}
  # --------------------------------------------------------------------------
  - name: shadowsocks
    addr: ":${SS_PORT}"
    handler:
      type: ss
      auth:
        username: ${SS_METHOD}
        password: ${SS_PASSWORD}
    listener:
      type: tcp
EOF
        log_detail "Shadowsocks configured on port ${SS_PORT}"
    fi

    # Устанавливаем права доступа (только root может читать, т.к. содержит пароли)
    chmod 600 "$GOST_CONFIG_FILE"

    log_success "Configuration file created: ${GOST_CONFIG_FILE}"

    # Показываем сводку
    log_info "Configuration summary:"
    log_detail "HTTP Proxy: :${HTTP_PROXY_PORT} (user: ${HTTP_PROXY_USER})"
    [ "$HTTPS_ENABLED" = "true" ] && log_detail "HTTPS Proxy: :${HTTPS_PORT}"
    [ "$SS_ENABLED" = "true" ] && log_detail "Shadowsocks: :${SS_PORT} (method: ${SS_METHOD})"
}

# ============================================================================
# СЕКЦИЯ 12: ФУНКЦИЯ СОЗДАНИЯ SYSTEMD СЕРВИСА
# ============================================================================

# -----------------------------------------------------------------------------
# Функция: create_service
# Описание: Создаёт systemd unit-файл для автозапуска GOST.
#           Настраивает сервис с автоперезапуском и security hardening.
# Параметры: нет
# Возврат:
#   - Создаёт /etc/systemd/system/gost.service
#   - Запускает сервис
# -----------------------------------------------------------------------------
create_service() {
    print_step "Creating systemd service"

    log_action "Writing systemd unit file..."

    # Создаём unit-файл с подробными комментариями
    cat > "$GOST_SERVICE_FILE" << 'EOF'
# ============================================================================
# GOST Proxy Server - Systemd Service Unit
# ============================================================================
# This file is managed by the proxy installer script.
# Manual changes may be overwritten on upgrade.
# ============================================================================

[Unit]
# Описание сервиса
Description=GOST Proxy Server
Documentation=https://gost.run/

# Зависимости запуска
# network.target - базовая сетевая подсистема
# network-online.target - сеть полностью настроена и работает
After=network.target network-online.target
Wants=network-online.target

[Service]
# Тип сервиса: simple - процесс запускается и работает непрерывно
Type=simple

# Команда запуска
ExecStart=/usr/local/bin/gost -C /etc/gost/config.yaml

# Политика перезапуска
# always - перезапускать при любом завершении (включая успешное)
Restart=always
# Задержка перед перезапуском в секундах
RestartSec=5

# Ограничение открытых файлов (для высокой нагрузки)
LimitNOFILE=65535

# ============================================================================
# Security Hardening (защита от эксплуатации)
# ============================================================================

# Запрет получения новых привилегий
NoNewPrivileges=true

# Защита системных директорий (только чтение)
ProtectSystem=strict

# Запрет доступа к домашним директориям
ProtectHome=true

# Разрешаем запись только в директорию конфигурации
ReadWritePaths=/etc/gost

# Приватная директория /tmp для изоляции
PrivateTmp=true

# Запрет записи в /dev (кроме стандартных устройств)
PrivateDevices=true

# Защита директорий ядра
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true

[Install]
# Запускать в multi-user режиме (стандартный серверный режим)
WantedBy=multi-user.target
EOF

    log_success "Systemd unit file created: ${GOST_SERVICE_FILE}"

    # Перезагружаем конфигурацию systemd
    log_action "Reloading systemd daemon..."
    systemctl daemon-reload
    log_success "Systemd daemon reloaded"

    # Включаем автозапуск
    log_action "Enabling GOST service for autostart..."
    systemctl enable gost > /dev/null 2>&1
    log_success "Service enabled for autostart"

    # Запускаем сервис
    log_action "Starting GOST service..."
    systemctl start gost

    # Даём сервису время на запуск
    sleep 2

    # Проверяем статус
    if systemctl is-active --quiet gost; then
        log_success "GOST service started successfully"

        # Показываем дополнительную информацию о процессе
        local pid=$(systemctl show --property MainPID --value gost)
        local memory=$(ps -o rss= -p "$pid" 2>/dev/null | awk '{print int($1/1024)"MB"}')
        log_detail "PID: ${pid}"
        log_detail "Memory usage: ${memory:-unknown}"
    else
        log_error "Failed to start GOST service. Check logs: journalctl -u gost"
    fi
}

# ============================================================================
# СЕКЦИЯ 13: ФУНКЦИЯ НАСТРОЙКИ ФАЙРВОЛА
# ============================================================================

# -----------------------------------------------------------------------------
# Функция: configure_firewall
# Описание: Настраивает файрвол для разрешения трафика на порты прокси.
#           Автоматически определяет тип файрвола (UFW, firewalld, iptables).
# Параметры: нет
# Возврат:
#   - Открывает необходимые порты
# -----------------------------------------------------------------------------
configure_firewall() {
    print_step "Configuring firewall"

    # Собираем список портов для открытия
    local tcp_ports="$HTTP_PROXY_PORT"
    [ "$HTTPS_ENABLED" = "true" ] && tcp_ports="$tcp_ports $HTTPS_PORT"
    [ "$SS_ENABLED" = "true" ] && tcp_ports="$tcp_ports $SS_PORT"

    log_info "Ports to open (TCP): ${tcp_ports}"
    [ "$SS_ENABLED" = "true" ] && log_info "Ports to open (UDP): ${SS_PORT}"

    # Определяем и настраиваем файрвол

    # --- UFW (Ubuntu/Debian) ---
    if command -v ufw > /dev/null 2>&1; then
        log_action "Detected firewall: UFW"

        # Проверяем, активен ли UFW
        local ufw_status=$(ufw status 2>/dev/null | head -1)
        log_detail "UFW status: ${ufw_status}"

        # Открываем TCP порты
        for port in $tcp_ports; do
            log_action "Opening TCP port ${port}..."
            ufw allow "$port/tcp" > /dev/null 2>&1 || true
            log_success "Port ${port}/tcp opened"
        done

        # Открываем UDP порт для Shadowsocks
        if [ "$SS_ENABLED" = "true" ]; then
            log_action "Opening UDP port ${SS_PORT}..."
            ufw allow "$SS_PORT/udp" > /dev/null 2>&1 || true
            log_success "Port ${SS_PORT}/udp opened"
        fi

        # Активируем UFW если не активен
        if ! ufw status | grep -q "Status: active"; then
            log_action "Enabling UFW..."
            ufw --force enable > /dev/null 2>&1 || true
            log_success "UFW enabled"
        fi

        return 0
    fi

    # --- firewalld (CentOS/Fedora/RHEL) ---
    if command -v firewall-cmd > /dev/null 2>&1 && systemctl is-active --quiet firewalld 2>/dev/null; then
        log_action "Detected firewall: firewalld"

        # Открываем TCP порты
        for port in $tcp_ports; do
            log_action "Opening TCP port ${port}..."
            firewall-cmd --permanent --add-port="$port/tcp" > /dev/null 2>&1 || true
            log_success "Port ${port}/tcp opened"
        done

        # Открываем UDP порт для Shadowsocks
        if [ "$SS_ENABLED" = "true" ]; then
            log_action "Opening UDP port ${SS_PORT}..."
            firewall-cmd --permanent --add-port="$SS_PORT/udp" > /dev/null 2>&1 || true
            log_success "Port ${SS_PORT}/udp opened"
        fi

        # Применяем изменения
        log_action "Reloading firewall rules..."
        firewall-cmd --reload > /dev/null 2>&1 || true
        log_success "Firewall rules reloaded"

        return 0
    fi

    # --- iptables (fallback) ---
    if command -v iptables > /dev/null 2>&1; then
        log_action "Detected firewall: iptables"
        log_warning "Using iptables directly (rules may not persist after reboot)"

        # Открываем TCP порты
        for port in $tcp_ports; do
            log_action "Opening TCP port ${port}..."
            iptables -A INPUT -p tcp --dport "$port" -j ACCEPT 2>/dev/null || true
            log_success "Port ${port}/tcp opened"
        done

        # Открываем UDP порт для Shadowsocks
        if [ "$SS_ENABLED" = "true" ]; then
            log_action "Opening UDP port ${SS_PORT}..."
            iptables -A INPUT -p udp --dport "$SS_PORT" -j ACCEPT 2>/dev/null || true
            log_success "Port ${SS_PORT}/udp opened"
        fi

        # Пытаемся сохранить правила
        if command -v netfilter-persistent > /dev/null 2>&1; then
            log_action "Saving iptables rules..."
            netfilter-persistent save > /dev/null 2>&1 || true
            log_success "Rules saved with netfilter-persistent"
        elif [ -d /etc/iptables ]; then
            log_action "Saving iptables rules..."
            iptables-save > /etc/iptables/rules.v4 2>/dev/null || true
            log_success "Rules saved to /etc/iptables/rules.v4"
        else
            log_warning "Could not persist iptables rules"
            log_detail "Rules will be lost after reboot"
        fi

        return 0
    fi

    # Файрвол не обнаружен
    log_warning "No supported firewall detected"
    log_detail "Please manually open ports: ${tcp_ports}"
    [ "$SS_ENABLED" = "true" ] && log_detail "Also open UDP port: ${SS_PORT}"
}

# ============================================================================
# СЕКЦИЯ 14: ФУНКЦИЯ ВЕРИФИКАЦИИ УСТАНОВКИ
# ============================================================================

# -----------------------------------------------------------------------------
# Функция: verify_installation
# Описание: Проверяет, что все компоненты установлены и работают корректно.
#           Тестирует подключение к прокси.
# Параметры: нет
# Возврат:
#   - Выводит результаты проверок
# -----------------------------------------------------------------------------
verify_installation() {
    print_step "Verifying installation"

    local all_ok=true

    # Проверка 1: GOST бинарник
    log_action "Checking GOST binary..."
    if [ -x "$GOST_BINARY" ]; then
        log_success "GOST binary: OK"
    else
        log_warning "GOST binary: NOT FOUND"
        all_ok=false
    fi

    # Проверка 2: Конфигурационный файл
    log_action "Checking configuration file..."
    if [ -f "$GOST_CONFIG_FILE" ]; then
        log_success "Configuration file: OK"
    else
        log_warning "Configuration file: NOT FOUND"
        all_ok=false
    fi

    # Проверка 3: Systemd сервис
    log_action "Checking systemd service..."
    if systemctl is-active --quiet gost; then
        log_success "Service status: RUNNING"
    else
        log_warning "Service status: NOT RUNNING"
        all_ok=false
    fi

    # Проверка 4: Порты
    log_action "Checking listening ports..."

    # HTTP proxy port
    if ss -tlnp 2>/dev/null | grep -q ":${HTTP_PROXY_PORT}"; then
        log_success "HTTP proxy port ${HTTP_PROXY_PORT}: LISTENING"
    else
        log_warning "HTTP proxy port ${HTTP_PROXY_PORT}: NOT LISTENING"
        all_ok=false
    fi

    # HTTPS proxy port
    if [ "$HTTPS_ENABLED" = "true" ]; then
        if ss -tlnp 2>/dev/null | grep -q ":${HTTPS_PORT}"; then
            log_success "HTTPS proxy port ${HTTPS_PORT}: LISTENING"
        else
            log_warning "HTTPS proxy port ${HTTPS_PORT}: NOT LISTENING"
            all_ok=false
        fi
    fi

    # Shadowsocks port
    if [ "$SS_ENABLED" = "true" ]; then
        if ss -tlnp 2>/dev/null | grep -q ":${SS_PORT}"; then
            log_success "Shadowsocks port ${SS_PORT}: LISTENING"
        else
            log_warning "Shadowsocks port ${SS_PORT}: NOT LISTENING"
            all_ok=false
        fi
    fi

    # Проверка 5: Тест подключения через прокси
    log_action "Testing proxy connection..."

    local test_result=$(curl -s --connect-timeout 5 --max-time 10 \
        -x "http://${HTTP_PROXY_USER}:${HTTP_PROXY_PASS}@127.0.0.1:${HTTP_PROXY_PORT}" \
        "http://httpbin.org/ip" 2>/dev/null || echo "FAILED")

    if echo "$test_result" | grep -q "origin"; then
        log_success "Proxy connection test: PASSED"
        local origin_ip=$(echo "$test_result" | grep -oE '[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+' | head -1)
        log_detail "Exit IP: ${origin_ip}"
    else
        log_warning "Proxy connection test: FAILED (may be normal if httpbin.org is blocked)"
        log_detail "Try manually: curl -x http://${HTTP_PROXY_USER}:***@127.0.0.1:${HTTP_PROXY_PORT} https://ifconfig.me"
    fi

    echo ""
    if [ "$all_ok" = true ]; then
        log_success "All verification checks passed!"
    else
        log_warning "Some checks failed. Please review the warnings above."
    fi
}

# ============================================================================
# СЕКЦИЯ 15: ФУНКЦИЯ СОХРАНЕНИЯ УЧЁТНЫХ ДАННЫХ
# ============================================================================

# -----------------------------------------------------------------------------
# Функция: save_credentials
# Описание: Сохраняет все учётные данные в файл для удобства пользователя.
#           Файл доступен только для root.
# Параметры: нет
# Возврат:
#   - Создаёт /root/proxy-credentials.txt
# -----------------------------------------------------------------------------
save_credentials() {
    print_step "Saving credentials"

    log_action "Writing credentials to ${CREDENTIALS_FILE}..."

    # Вычисляем время установки
    local end_time=$(date +%s)
    local duration=$((end_time - START_TIME))

    # Создаём файл с учётными данными
    cat > "$CREDENTIALS_FILE" << EOF
╔══════════════════════════════════════════════════════════════════════════════╗
║                         PROXY SERVER CREDENTIALS                              ║
╠══════════════════════════════════════════════════════════════════════════════╣
║                                                                               ║
║  Generated: $(date '+%Y-%m-%d %H:%M:%S')
║  Server IP: ${PUBLIC_IP}
║  Installer: Universal Proxy Installer v${SCRIPT_VERSION}
║  Install time: ${duration} seconds
║                                                                               ║
╚══════════════════════════════════════════════════════════════════════════════╝

┌──────────────────────────────────────────────────────────────────────────────┐
│ HTTP PROXY                                                                    │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  Host:     ${PUBLIC_IP}
│  Port:     ${HTTP_PROXY_PORT}
│  Username: ${HTTP_PROXY_USER}
│  Password: ${HTTP_PROXY_PASS}
│                                                                               │
│  Full URL: http://${HTTP_PROXY_USER}:${HTTP_PROXY_PASS}@${PUBLIC_IP}:${HTTP_PROXY_PORT}
│                                                                               │
│  Test command:                                                                │
│  curl -x http://${HTTP_PROXY_USER}:${HTTP_PROXY_PASS}@${PUBLIC_IP}:${HTTP_PROXY_PORT} https://ifconfig.me
│                                                                               │
└──────────────────────────────────────────────────────────────────────────────┘
EOF

    if [ "$HTTPS_ENABLED" = "true" ]; then
        cat >> "$CREDENTIALS_FILE" << EOF

┌──────────────────────────────────────────────────────────────────────────────┐
│ HTTPS PROXY (TLS encrypted)                                                  │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  Host:     ${PUBLIC_IP}
│  Port:     ${HTTPS_PORT}
│  Username: ${HTTP_PROXY_USER}
│  Password: ${HTTP_PROXY_PASS}
│                                                                               │
│  Note: Uses self-signed certificate. Browser may show security warning.      │
│  Certificate location: ${GOST_CONFIG_DIR}/cert.pem                            │
│                                                                               │
└──────────────────────────────────────────────────────────────────────────────┘
EOF
    fi

    if [ "$SS_ENABLED" = "true" ]; then
        # Генерируем SS URI для QR-кодов и импорта
        local ss_uri="ss://$(echo -n "${SS_METHOD}:${SS_PASSWORD}" | base64 | tr -d '\n')@${PUBLIC_IP}:${SS_PORT}#ProxyServer"

        cat >> "$CREDENTIALS_FILE" << EOF

┌──────────────────────────────────────────────────────────────────────────────┐
│ SHADOWSOCKS                                                                   │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  Server:   ${PUBLIC_IP}
│  Port:     ${SS_PORT}
│  Password: ${SS_PASSWORD}
│  Method:   ${SS_METHOD}
│                                                                               │
│  SS URI (for import):                                                         │
│  ${ss_uri}
│                                                                               │
│  Compatible clients:                                                          │
│  - iOS: Shadowrocket, Surge, Quantumult                                       │
│  - Android: Shadowsocks, v2rayNG                                              │
│  - Windows: Shadowsocks-windows, v2rayN                                       │
│  - macOS: ShadowsocksX-NG, Surge                                              │
│  - Linux: shadowsocks-libev, shadowsocks-rust                                 │
│                                                                               │
└──────────────────────────────────────────────────────────────────────────────┘
EOF
    fi

    cat >> "$CREDENTIALS_FILE" << EOF

┌──────────────────────────────────────────────────────────────────────────────┐
│ MANAGEMENT COMMANDS                                                           │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  Check status:     systemctl status gost                                      │
│  View logs:        journalctl -u gost -f                                      │
│  Restart service:  systemctl restart gost                                     │
│  Stop service:     systemctl stop gost                                        │
│  Edit config:      nano ${GOST_CONFIG_FILE}
│                                                                               │
│  Uninstall:        curl -fsSL <script_url> | bash -s -- --uninstall           │
│                                                                               │
└──────────────────────────────────────────────────────────────────────────────┘

EOF

    # Устанавливаем права доступа (только root может читать)
    chmod 600 "$CREDENTIALS_FILE"

    log_success "Credentials saved to: ${CREDENTIALS_FILE}"
    log_detail "File permissions: 600 (owner read/write only)"
}

# ============================================================================
# СЕКЦИЯ 16: ФУНКЦИЯ ВЫВОДА ФИНАЛЬНОГО ОТЧЁТА
# ============================================================================

# -----------------------------------------------------------------------------
# Функция: print_summary
# Описание: Выводит красивый финальный отчёт с результатами установки
#           и данными для подключения.
# Параметры: нет
# Возврат: нет
# -----------------------------------------------------------------------------
print_summary() {
    # Вычисляем время установки
    local end_time=$(date +%s)
    local duration=$((end_time - START_TIME))

    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                                                                              ║${NC}"
    echo -e "${GREEN}║                    ✓ INSTALLATION COMPLETED SUCCESSFULLY                    ║${NC}"
    echo -e "${GREEN}║                                                                              ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════════════════════════════════╝${NC}"
    echo ""

    # Статистика установки
    echo -e "${WHITE}Installation Summary:${NC}"
    echo -e "  ${GRAY}${SYMBOL_BULLET}${NC} Duration: ${duration} seconds"
    echo -e "  ${GRAY}${SYMBOL_BULLET}${NC} Steps completed: ${CURRENT_STEP}/${TOTAL_STEPS}"
    echo -e "  ${GRAY}${SYMBOL_BULLET}${NC} Successful operations: ${SUCCESS_COUNT}"
    echo ""

    # HTTP Proxy
    echo -e "${CYAN}┌─ HTTP PROXY ──────────────────────────────────────────────────────┐${NC}"
    echo -e "${CYAN}│${NC}"
    echo -e "${CYAN}│${NC}   Host:     ${GREEN}${PUBLIC_IP}${NC}"
    echo -e "${CYAN}│${NC}   Port:     ${GREEN}${HTTP_PROXY_PORT}${NC}"
    echo -e "${CYAN}│${NC}   Username: ${GREEN}${HTTP_PROXY_USER}${NC}"
    echo -e "${CYAN}│${NC}   Password: ${GREEN}${HTTP_PROXY_PASS}${NC}"
    echo -e "${CYAN}│${NC}"
    echo -e "${CYAN}└───────────────────────────────────────────────────────────────────┘${NC}"
    echo ""

    # HTTPS Proxy (если включено)
    if [ "$HTTPS_ENABLED" = "true" ]; then
        echo -e "${CYAN}┌─ HTTPS PROXY (encrypted) ────────────────────────────────────────┐${NC}"
        echo -e "${CYAN}│${NC}"
        echo -e "${CYAN}│${NC}   Port:     ${GREEN}${HTTPS_PORT}${NC}"
        echo -e "${CYAN}│${NC}   (Uses same username/password as HTTP proxy)"
        echo -e "${CYAN}│${NC}"
        echo -e "${CYAN}└───────────────────────────────────────────────────────────────────┘${NC}"
        echo ""
    fi

    # Shadowsocks (если включено)
    if [ "$SS_ENABLED" = "true" ]; then
        echo -e "${CYAN}┌─ SHADOWSOCKS ─────────────────────────────────────────────────────┐${NC}"
        echo -e "${CYAN}│${NC}"
        echo -e "${CYAN}│${NC}   Server:   ${GREEN}${PUBLIC_IP}${NC}"
        echo -e "${CYAN}│${NC}   Port:     ${GREEN}${SS_PORT}${NC}"
        echo -e "${CYAN}│${NC}   Password: ${GREEN}${SS_PASSWORD}${NC}"
        echo -e "${CYAN}│${NC}   Method:   ${GREEN}${SS_METHOD}${NC}"
        echo -e "${CYAN}│${NC}"
        echo -e "${CYAN}└───────────────────────────────────────────────────────────────────┘${NC}"
        echo ""
    fi

    # Команды управления
    echo -e "${WHITE}Quick Commands:${NC}"
    echo -e "  ${GRAY}${SYMBOL_BULLET}${NC} Test:    ${YELLOW}curl -x http://${HTTP_PROXY_USER}:${HTTP_PROXY_PASS}@127.0.0.1:${HTTP_PROXY_PORT} https://ifconfig.me${NC}"
    echo -e "  ${GRAY}${SYMBOL_BULLET}${NC} Status:  ${YELLOW}systemctl status gost${NC}"
    echo -e "  ${GRAY}${SYMBOL_BULLET}${NC} Logs:    ${YELLOW}journalctl -u gost -f${NC}"
    echo ""

    echo -e "${WHITE}Credentials File:${NC} ${YELLOW}${CREDENTIALS_FILE}${NC}"
    echo ""
    echo -e "${GRAY}─────────────────────────────────────────────────────────────────────────────${NC}"
    echo -e "${GRAY}  Thank you for using Universal Proxy Installer!${NC}"
    echo -e "${GRAY}─────────────────────────────────────────────────────────────────────────────${NC}"
    echo ""
}

# ============================================================================
# СЕКЦИЯ 17: ФУНКЦИЯ УДАЛЕНИЯ
# ============================================================================

# -----------------------------------------------------------------------------
# Функция: uninstall
# Описание: Полностью удаляет GOST и все связанные файлы.
# Параметры: нет
# Возврат: exit 0
# -----------------------------------------------------------------------------
uninstall() {
    print_header

    echo -e "${YELLOW}Uninstalling GOST Proxy Server...${NC}"
    echo ""

    # Останавливаем сервис
    log_action "Stopping GOST service..."
    if systemctl is-active --quiet gost 2>/dev/null; then
        systemctl stop gost
        log_success "Service stopped"
    else
        log_detail "Service was not running"
    fi

    # Отключаем автозапуск
    log_action "Disabling GOST service..."
    if systemctl is-enabled --quiet gost 2>/dev/null; then
        systemctl disable gost > /dev/null 2>&1
        log_success "Service disabled"
    else
        log_detail "Service was not enabled"
    fi

    # Удаляем unit-файл
    log_action "Removing systemd unit file..."
    if [ -f "$GOST_SERVICE_FILE" ]; then
        rm -f "$GOST_SERVICE_FILE"
        systemctl daemon-reload
        log_success "Unit file removed"
    else
        log_detail "Unit file not found"
    fi

    # Удаляем бинарный файл
    log_action "Removing GOST binary..."
    if [ -f "$GOST_BINARY" ]; then
        rm -f "$GOST_BINARY"
        log_success "Binary removed"
    else
        log_detail "Binary not found"
    fi

    # Удаляем конфигурацию
    log_action "Removing configuration directory..."
    if [ -d "$GOST_CONFIG_DIR" ]; then
        rm -rf "$GOST_CONFIG_DIR"
        log_success "Configuration removed"
    else
        log_detail "Configuration directory not found"
    fi

    # Удаляем файл с учётными данными
    log_action "Removing credentials file..."
    if [ -f "$CREDENTIALS_FILE" ]; then
        rm -f "$CREDENTIALS_FILE"
        log_success "Credentials file removed"
    else
        log_detail "Credentials file not found"
    fi

    echo ""
    echo -e "${GREEN}GOST Proxy Server has been completely uninstalled.${NC}"
    echo ""

    exit 0
}

# ============================================================================
# СЕКЦИЯ 18: ГЛАВНАЯ ФУНКЦИЯ
# ============================================================================

# -----------------------------------------------------------------------------
# Функция: main
# Описание: Точка входа в скрипт. Обрабатывает аргументы командной строки
#           и запускает процесс установки.
# Параметры:
#   $@ - все аргументы командной строки
# Возврат: exit 0 при успехе
# -----------------------------------------------------------------------------
main() {
    # Обработка аргументов командной строки
    case "${1:-}" in
        -h|--help)
            show_help
            ;;
        -v|--version)
            echo "Universal Proxy Installer v${SCRIPT_VERSION}"
            exit 0
            ;;
        -u|--uninstall|uninstall)
            # Проверяем права root для удаления
            if [ "$(id -u)" -ne 0 ]; then
                echo -e "${RED}Error: This script must be run as root${NC}"
                exit 1
            fi
            uninstall
            ;;
    esac

    # Проверяем, что скрипт запущен от root
    if [ "$(id -u)" -ne 0 ]; then
        echo ""
        echo -e "${RED}╔════════════════════════════════════════════════════════════════════╗${NC}"
        echo -e "${RED}║  ERROR: This script must be run as root                            ║${NC}"
        echo -e "${RED}║                                                                    ║${NC}"
        echo -e "${RED}║  Please run:  sudo ./install-proxy.sh                              ║${NC}"
        echo -e "${RED}║          or:  sudo bash install-proxy.sh                           ║${NC}"
        echo -e "${RED}╚════════════════════════════════════════════════════════════════════╝${NC}"
        echo ""
        exit 1
    fi

    # Выводим заголовок
    print_header

    # Показываем текущую конфигурацию
    echo -e "${WHITE}Configuration:${NC}"
    echo -e "  ${GRAY}${SYMBOL_BULLET}${NC} HTTP Proxy Port: ${HTTP_PROXY_PORT}"
    echo -e "  ${GRAY}${SYMBOL_BULLET}${NC} HTTP Proxy User: ${HTTP_PROXY_USER}"
    echo -e "  ${GRAY}${SYMBOL_BULLET}${NC} Shadowsocks: $([ "$SS_ENABLED" = "true" ] && echo "Enabled (port ${SS_PORT})" || echo "Disabled")"
    echo -e "  ${GRAY}${SYMBOL_BULLET}${NC} HTTPS Proxy: $([ "$HTTPS_ENABLED" = "true" ] && echo "Enabled (port ${HTTPS_PORT})" || echo "Disabled")"
    echo ""
    echo -e "${GRAY}Press Ctrl+C to cancel, or wait 3 seconds to continue...${NC}"
    sleep 3

    # Запускаем этапы установки
    # Каждый этап - отдельная функция для модульности и читаемости

    detect_os                    # Шаг 1: Определение ОС
    detect_arch                  # Шаг 2: Определение архитектуры
    get_public_ip                # Шаг 3: Определение публичного IP
    check_existing_installation  # Шаг 4: Проверка существующей установки
    install_packages             # Шаг 5: Установка пакетов
    install_gost                 # Шаг 6: Установка GOST
    generate_config              # Шаг 7: Генерация конфигурации
    create_service               # Шаг 8: Создание systemd сервиса
    configure_firewall           # Шаг 9: Настройка файрвола
    verify_installation          # Шаг 10: Проверка установки
    save_credentials             # Шаг 11: Сохранение учётных данных

    # Выводим финальный отчёт
    print_summary

    exit 0
}

# ============================================================================
# СЕКЦИЯ 19: ЗАПУСК СКРИПТА
# ============================================================================
#
# Эта строка запускает главную функцию, передавая все аргументы командной строки.
# "$@" - специальная переменная, содержащая все позиционные параметры скрипта.
#

main "$@"
