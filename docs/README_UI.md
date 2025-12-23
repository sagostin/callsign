# CallSign-UI

A modern, feature-rich PBX management interface built with Vue 3 and Vite. CallSign-UI provides a comprehensive admin panel for managing enterprise telephony systems powered by FreeSWITCH.

![Vue 3](https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat-square&logo=vue.js)
![Vite](https://img.shields.io/badge/Vite-5.x-646CFF?style=flat-square&logo=vite)
![License](https://img.shields.io/badge/License-Proprietary-blue?style=flat-square)

---

## 🚀 Features

### Multi-Tenant Architecture
- **System Admin Panel** - Manage tenants, global settings, and infrastructure
- **Tenant Admin Panel** - Per-tenant configuration and management
- **User Portal** - End-user softphone, voicemail, and settings

### Core PBX Features

| Feature | Description |
|---------|-------------|
| **Extensions** | Full extension management with profiles, permissions, and call handling rules |
| **IVR / Auto Attendant** | Visual menu builder, time conditions, and mode toggles |
| **Call Routing** | Inbound/outbound routing with FreeSWITCH-compatible dialplan logic |
| **Call Queues** | ACD queues with agent management, strategies, and real-time stats |
| **Ring Groups** | Simultaneous, sequential, and enterprise ring strategies |
| **Conferencing** | Audio conference rooms with live participant management |
| **Voicemail** | Voicemail boxes with email delivery, transcription, and MWI |
| **Call Recording** | Automatic and on-demand recording with playback and export |

### Communication

| Feature | Description |
|---------|-------------|
| **SMS/MMS Messaging** | Two-way messaging with provider integrations |
| **Fax Server** | Multi-server virtual fax with email-to-fax support |
| **WebRTC Softphone** | Browser-based dialer with device binding |

### Device Provisioning

| Feature | Description |
|---------|-------------|
| **Auto Provisioning** | Zero-touch provisioning for Yealink, Polycom, Grandstream |
| **Device Templates** | Customizable button layouts and settings |
| **Firmware Management** | Upload and deploy firmware to devices |

### Administration

| Feature | Description |
|---------|-------------|
| **Reports & Analytics** | Call volume, agent performance, KPIs |
| **Audit Logging** | Full activity tracking and compliance |
| **Feature Codes** | Customizable star codes for call features |
| **Music on Hold** | Stream management for hold music |
| **Config Inspector** | FreeSWITCH XML configuration viewer |

### Specialized Modules

| Feature | Description |
|---------|-------------|
| **Hospitality** | Hotel/PMS integration, wake-up calls, room management |
| **E911 Locations** | Location-based emergency routing |
| **Call Broadcast** | Mass notification campaigns |
| **Call Block** | Blacklist and pattern-based blocking |

---

## 📁 Project Structure

```
ui/
├── src/
│   ├── components/           # Reusable Vue components
│   │   ├── common/           # DataTable, StatusBadge, etc.
│   │   ├── flow/             # Visual call flow editor
│   │   ├── ivr/              # IVR-specific components
│   │   └── layout/           # LayoutShell, Sidebar, TopBar
│   │
│   ├── views/                # Page components (98 total)
│   │   ├── admin/            # 64 tenant admin views
│   │   ├── auth/             # 2 auth views
│   │   ├── system/           # 26 system admin views
│   │   └── user/             # 8 user portal views
│   │
│   ├── services/             # API service layer
│   ├── styles/               # Global CSS
│   ├── router.js             # Vue Router configuration
│   ├── main.js               # Application entry point
│   └── App.vue               # Root component
│
├── public/                   # Static assets
└── package.json
```

---

## 🛠️ Tech Stack

- **Framework**: Vue 3 with Composition API (`<script setup>`)
- **Build Tool**: Vite 5
- **Routing**: Vue Router 4
- **Icons**: Lucide Vue Next
- **Styling**: Vanilla CSS with CSS custom properties

---

## 🏃 Getting Started

### Prerequisites

- Node.js 18+ 
- npm or yarn

### Installation

```bash
# Clone the repository
git clone https://github.com/your-org/callsign.git
cd callsign/ui

# Install dependencies
npm install

# Start development server
npm run dev
```

The app will be available at `http://localhost:5173`

### Build for Production

```bash
npm run build
```

Output will be in the `dist/` directory.

---

## 🧭 Application Routes

### User Portal (`/`)
| Route | Description |
|-------|-------------|
| `/dialer` | WebRTC softphone |
| `/messages` | SMS/MMS conversations |
| `/voicemail` | Personal voicemail |
| `/conferences` | User conference rooms |
| `/fax` | Personal fax inbox |
| `/contacts` | Contact management |
| `/recordings` | Call recordings |
| `/history` | Call history |
| `/settings` | Personal settings |

### Tenant Admin (`/admin`)
| Route | Description |
|-------|-------------|
| `/admin` | Dashboard overview |
| `/admin/extensions` | Extension management |
| `/admin/ivr` | IVR/Auto Attendant |
| `/admin/routing` | Call routing & DIDs |
| `/admin/devices` | Device provisioning |
| `/admin/queues` | Queues & Ring Groups |
| `/admin/conferences` | Conference rooms |
| `/admin/voicemail-manager` | Voicemail boxes |
| `/admin/fax` | Fax servers |
| `/admin/messaging` | SMS/MMS |
| `/admin/call-recordings` | Recording manager |
| `/admin/reports` | Analytics & reports |
| `/admin/settings` | Tenant settings |
| `/admin/time-conditions` | Time conditions |
| `/admin/call-flows` | Call flow toggles |

### System Admin (`/system`)
| Route | Description |
|-------|-------------|
| `/system` | System dashboard |
| `/system/tenants` | Tenant management |
| `/system/profiles` | Tenant profiles/plans |
| `/system/gateways` | SIP gateways/trunks |
| `/system/sip-profiles` | SIP profile config |
| `/system/dial-plans` | Global dial plans |
| `/system/acls` | Access control lists |
| `/system/provisioning-templates` | Master device templates |
| `/system/firmware` | Firmware management |
| `/system/sounds` | System sounds |
| `/system/music` | Music on hold |
| `/system/messaging` | SMS provider config |
| `/system/logs` | System logs |
| `/system/settings` | Global settings |
| `/system/security` | Security & banned IPs |
| `/system/config-inspector` | FreeSWITCH config viewer |

---

## 🔌 Backend Integration

This is a **frontend-only** directory. The backend API is in the `../api` directory.

See [`BACKEND_TODO.md`](BACKEND_TODO.md) for a complete list of ~375 API endpoints, ~280 of which are implemented.

### Backend Stack
- **PBX Engine**: FreeSWITCH
- **API**: Go Iris + GORM + PostgreSQL
- **Auth**: JWT with role-based access control
- **Real-time**: WebSocket for notifications, FreeSWITCH console

---

## 🎨 Design System

The UI uses a consistent design system with CSS custom properties:

```css
/* Primary Colors */
--primary-color: #6366f1;
--primary-light: #eef2ff;

/* Status Colors */
--status-good: #22c55e;
--status-bad: #ef4444;

/* Typography */
--text-primary: #0f172a;
--text-main: #334155;
--text-muted: #64748b;

/* Spacing */
--spacing-sm: 8px;
--spacing-md: 16px;
--spacing-lg: 24px;
--spacing-xl: 32px;

/* Borders & Shadows */
--border-color: #e2e8f0;
--radius-sm: 6px;
--radius-md: 10px;
--shadow-sm: 0 1px 3px rgba(0,0,0,0.08);
```

---

## 📱 Responsive Design

The UI is optimized for:
- **Desktop**: Full-featured admin experience
- **Tablet**: Collapsible sidebar, touch-friendly controls
- **Mobile**: Essential functionality with adaptive layouts

---

## 🤝 Contributing

1. Create a feature branch from `main`
2. Follow Vue 3 Composition API patterns
3. Use existing component patterns (DataTable, StatusBadge, modal patterns)
4. Test on multiple viewport sizes
5. Submit a pull request

---

## 📄 License

Proprietary - All rights reserved.

---

## 📞 Support

For questions or issues, contact the development team.
