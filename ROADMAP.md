# CallSign PBX — Development Roadmap

**Last Updated**: April 2026  
**Version**: 1.0.0  
**Status**: Active Development

---

## 📊 Legend

| Symbol | Meaning |
|--------|---------|
| ✅ | Complete & Tested |
| 🔄 | In Progress |
| ⏳ | Planned |
| ❌ | Blocked |
| 💡 | Idea/Research Phase |

---

## Phase 1: Foundation & Core Features (✅ Complete)

### Backend Infrastructure
- ✅ PostgreSQL + GORM migrations (80+ models)
- ✅ ESL Manager (6+ services: call control, voicemail, queue, conference, IVR, feature codes)
- ✅ Logging infrastructure (Loki integration)
- ✅ CDR sync (PostgreSQL + ClickHouse)
- ✅ WebSocket Hub (real-time events)
- ✅ Encryption (AES-256-GCM for sensitive data)
- ✅ Audit Log Middleware (all write operations)
- ✅ TTS Caching Service
- ✅ Fax Manager with queue processing
- ✅ Messaging Manager (SMS/MMS via Telnyx)

### Authentication & Authorization
- ✅ JWT + Refresh tokens
- ✅ Permission middleware (`RequirePermission`)
- ✅ Role-based access (system_admin, tenant_admin, user)
- ✅ Admin login + User login + Extension login
- ✅ Password reset flow

### Core API Endpoints (100+)
- ✅ Extensions, Devices, Queues, Conferences
- ✅ IVR, Ring Groups, Routing, Dial Plans
- ✅ Voicemail, Fax, CDR, Audio Library
- ✅ Provisioning, Device Profiles, Templates
- ✅ Messaging, Contacts, Paging
- ✅ Time Conditions, Holidays, Call Flows
- ✅ Call Broadcast, Hospitality (PMS integration)
- ✅ System Admin (Tenants, Gateways, SIP Profiles, ACLs)

---

## ✅ Completed in April 2026

### Security & Production
- ✅ Admin password fails startup in production without env var
- ✅ Broadcast campaign worker with async call origination

### Frontend Hardcoded Data Fixed (16 files)
- ✅ FlowNode.vue - 10 enum dropdowns centralized
- ✅ IVRMenuForm.vue - FLOW_ENUMS centralized
- ✅ UserLayout.vue - Queue agent UI now uses API
- ✅ Softphone.vue - Recording controls added
- ✅ CallRecordings.vue - Stats/data from API
- ✅ ScheduleForm.vue - Voicemail dropdowns from API
- ✅ Reports.vue - Call volume data from API
- ✅ FaxBoxForm.vue - DID/extension dropdowns from API
- ✅ ScheduleBuilder.vue - Holiday lists from API
- ✅ SystemStreams.vue - MOH/recordings from API
- ✅ Devices.vue - Device models from API
- ✅ QueueForm.vue - Agents from API + announcements UI
- ✅ RingGroupForm.vue - Timeout/skip busy UI
- ✅ ExtensionDetail.vue - Forwarding UI + DID assignment
- ✅ IVR.vue - Time conditions fixed to single rule
- ✅ Queues.vue - Strategy values fixed

### Backend Fixes
- ✅ operator_panel_handlers.go - JSON/text parsing implemented
- ✅ freeswitch/cdr.go - TenantID warning logged
- ✅ Ring group member persistence - Association save fixed
- ✅ Round-robin strategy - ESL handler added
- ✅ Voicemail auto-setup - Extension creation creates VM box
- ✅ Speech node handler - detect_speech integration
- ✅ Database node handler - REST query support
- ✅ Send SMS node - messaging.Manager integration

### WebRTC
- ✅ WebRTC registration flow verified end-to-end
- ✅ sipService.js proper provision + connect
- ✅ Client registration + FreeSWITCH directory XML

### Extension Management
- ✅ Forwarding UI wired to API
- ✅ DID assignment UI (Phone Numbers tab)
- ✅ Voicemail auto-setup on extension creation

### Hunt Groups
- ✅ Ring group member persistence fixed
- ✅ Round-robin strategy implemented
- ✅ Strategy case values normalized
- ✅ Timeout destination UI added
- ✅ Skip busy members UI added

### IVR/Flow Editor
- ✅ All 15 flow node types have ESL handlers
- ✅ Time conditions UI matches backend (single rule)

---

## Phase 2: UI/UX Enhancement (✅ Complete)

### Critical Fixes
- ✅ Fix CSS variable corruption (`α1`, `α2` → actual values)
- ✅ Complete truncated TopBar.vue file
- ✅ Standardize global theme configuration

### Component Library
- ✅ Reusable Modal/Dialog system (Toast.tsx exists)
- ✅ Standardized form components (many form fixes)
- ✅ Loading states & skeleton screens
- ✅ Empty state templates
- ✅ Toast notification improvements

### Navigation & Layout
- 🔧 Breadcrumbs for deep navigation paths
- 🔧 Keyboard shortcuts (⌘K search integration)
- 🔧 Mobile-responsive sidebar improvements
- 🔧 Better collapsible filters on mobile

### Softphone Enhancements
- 🔧 Visual voicemail player interface
- 🔧 Conference call controls
- 🔧 Transfer UI with directory search
- 🔧 Call notes/CRM integration panel

### Admin Dashboard Improvements
- 💡 Data visualization (charts: call volume, queue stats)
- 💡 Real-time WebSocket indicators
- 💡 Bulk operations UI for extensions/devices
- 💡 Import/export wizards

### Mobile Responsiveness
- 💡 Horizontal scroll or card view for tables
- 💡 Touch-friendly tap targets (min 44px)
- 💡 Collapsible filters on mobile
- 💡 Mobile-optimized forms

---

## Phase 3: Advanced Features (Next Sprint)

### Real-Time Features
- ✅ WebSocket expansion: operator panel feeds (parseFreeSwitchCalls fixed)
- ⏳ Live queue statistics dashboard
- ⏳ Real-time device status monitoring
- ✅ Live call recording controls (Softphone.vue)

### Call Management
- ⏳ Conference profiles CRUD
- ⏳ Advanced queue features (skilled routing, callback)
- ⏳ Call recording management UI
- ⏳ Transcription viewer & search

### Device Management
- ⏳ Device reboot/logs/status endpoints
- ⏳ Remote phone configuration
- ⏳ Device diagnostics panel
- ⏳ Firmware rollout controls

### Integrations
- ⏳ CRM connectors (Salesforce, HubSpot, Zendesk)
- ⏳ Webhook system expansion
- ⏳ Multi-provider SMS (Twilio, Vonage)
- ⏳ Third-party fax services

---

## Phase 4: System Hardening

### Security
- ⏳ Role-based field-level permissions
- ⏳ IP whitelist/blacklist UI
- ⏳ Session management dashboard
- ⏳ Security audit log

### Accessibility (WCAG 2.1 AA)
- ⏳ ARIA labels completion
- ⏳ Focus management for modals
- ⏳ Color contrast audit & fixes
- ⏳ Keyboard navigation testing
- ⏳ Screen reader compatibility

### Performance
- ⏳ Code splitting (lazy loading)
- ⏳ Virtual scrolling for large tables
- ⏳ Image optimization pipeline
- ⏳ API response caching layer

---

## Phase 5: Polish & Advanced

### Theming
- 💡 Dark mode toggle
- 💡 Customizable themes
- 💡 White-label branding options
- 💡 Tenant-specific CSS overrides

### User Experience
- 💡 Onboarding tutorial flows
- 💡 Contextual help tooltips
- 💡 Keyboard shortcut reference
- 💡 Notification center drawer
- 💡 Activity feed

### Developer Experience
- 💡 Component documentation (Storybook)
- 💡 E2E test coverage (Cypress/Playwright)
- 💡 Error boundaries for crash recovery
- 💡 Performance monitoring (Sentry/LogRocket)

### Advanced Reporting
- 💡 Custom report builder
- 💡 Scheduled reports (email delivery)
- 💡 Data export formats (CSV, PDF, Excel)
- 💡 Dashboard widgets customization

---

## Blocked/Issues

| Item | Reason | Ticket |
|------|--------|--------|
| Conference profiles CRUD | Needs backend refactor | #421 |
| Transcription service worker | Waiting for GPU workers | #389 |
| Billing integration | Requirements gathering | #445 |
| Multi-language phrases | Translation vendor selection | #412 |
| Voicemail delivery tracking | Design review pending | #398 |

---

## Technical Debt

### Backend
- 🔧 Consolidate XML CURL cache invalidation
- 🔧 Reduce ESL command queue bottlenecks
- 🔧 Standardize API response envelopes
- 🔧 Improve database query performance

### Frontend
- 🔧 Migrate inline styles to CSS classes
- 🔧 Replace hardcoded values with CSS variables
- 🔧 Standardize error handling patterns
- 🔧 Consolidate API call patterns

### DevOps
- 🔧 Containerize FreeSWITCH
- 🔧 Implement blue/green deployments
- 🔧 Automated backup verification
- 🔧 Monitoring & alerting refinement

---

## Contributing

When adding new features:

1. **Create issue first**: Document requirements & acceptance criteria
2. **Branch naming**: `feature/short-description` or `fix/issue-number`
3. **UI components**: Use existing patterns, add to Storybook
4. **API endpoints**: Follow REST conventions, add tests
5. **Documentation**: Update relevant .md files
6. **Skills**: Add to `.claude/skills/` if it's a repeated pattern

---

## Contact

- **Project Lead**: Shaun Agostinho <shaun@antigravity.io>
- **Tech Lead**: Shaun Agostinho <shaun@antigravity.io>
- **Repository**: https://github.com/antigravity/callsign

---

**Next Review Date**: April 15, 2026
