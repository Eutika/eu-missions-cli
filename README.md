<p align="center">
   <img width="auto" height="300" src="https://github.com/user-attachments/assets/26ff6ec2-1dc4-452b-9442-d66f20138073">
</p>

# Missions CLI

[![Release](https://img.shields.io/github/release/eutika/eu-missions-cli.svg)](https://github.com/eutika/eu-missions-cli/releases)

## Descripción General

Una potente interfaz de línea de comandos (CLI) para interactuar con las etapas de Missions, construida con Go y OAuth 2.0.

## Características Principales

- 🔐 Autenticación Segura mediante OAuth 2.0 Device Flow
- 🚀 Recuperación Dinámica de Comandos
- 🌐 Compatibilidad Multiplataforma
- 🔧 Gestión Flexible de Configuración

## Instalación

### Métodos de Instalación Rápida

#### Linux/macOS

```bash
curl -fsSL https://raw.githubusercontent.com/eutika/eu-missions-cli/main/scripts/install.sh | bash
```

#### Windows (PowerShell)

```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force;
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072;
iex ((New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/eutika/eu-missions-cli/main/scripts/install.ps1'))
```

### Instalación Manual

Descarga la última versión para tu plataforma desde la [página de Releases](https://github.com/eutika/eu-missions-cli/releases).

## Uso

```bash
missions [comando]
```

### Comandos Disponibles

- `login`: Autenticarse con tu cuenta

  - Inicia el Flujo de Dispositivo OAuth 2.0
  - Abre el navegador para verificación
  - Almacena de forma segura los tokens de autenticación

- `validate [id]`: Validar resultados de comandos desde el servicio remoto

  - Recupera y ejecuta comandos dinámicamente
  - Soporta ejecución flexible de comandos

- `submit [id]`: Enviar resultados de comandos al servicio remoto
  - Recupera y ejecuta comandos dinámicamente
  - Soporta ejecución flexible de comandos

## Configuración

La CLI soporta configuraciones específicas por entorno:

- Detección automática de archivos de configuración
- Soporte para `.env` y variables de entorno
- Gestión flexible de flags y configuración

## Compilación desde el Código Fuente

```bash
# Clonar el repositorio
git clone https://github.com/eutika/eu-missions-cli.git

# Compilar la CLI
go build -o missions .
```

## Desarrollo

### Requisitos

- Go 1.22 o más reciente
- Biblioteca Cobra CLI
- Servidor de autorización compatible con OAuth 2.0

### Contribuir

1. Haz un fork del repositorio
2. Crea tu rama de funcionalidad (`git checkout -b feature/FunciónIncreíble`)
3. Confirma tus cambios (`git commit -m 'Añadir alguna FunciónIncreíble'`)
4. Sube a la rama (`git push origin feature/FunciónIncreíble`)
5. Abre una Solicitud de Extracción (Pull Request)

## Licencia

[MIT](LICENSE)

## Soporte

Para problemas, solicitudes de funciones o discusiones, utiliza la [página de Issues de GitHub](https://github.com/eutika/eu-missions-cli/issues).
