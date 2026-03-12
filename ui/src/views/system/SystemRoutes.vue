<template>
  <div class="view-header">
    <div class="header-content">
      <h2>System Routing</h2>
      <p class="text-muted text-sm">Manage global inbound/outbound routing and dialplan logic.</p>
    </div>
    <div class="header-actions">
      <button class="btn-secondary" @click="recalculateOrder" v-if="activeTab === 'inbound' || activeTab === 'outbound'">
        <RefreshCwIcon class="btn-icon" /> Recalculate Order
      </button>
      <button class="btn-primary" v-if="activeTab === 'numbers' && numbersSubTab === 'numbers'" @click="openAddNumber">+ Add Number</button>
      <button class="btn-primary" v-if="activeTab === 'numbers' && numbersSubTab === 'groups'" @click="openAddGroup">+ New Group</button>
      <button class="btn-primary" v-if="activeTab === 'inbound'" @click="showInboundModal = true">+ Global Inbound</button>
      <button class="btn-primary" v-if="activeTab === 'outbound'" @click="showOutboundModal = true">+ Global Outbound</button>
    </div>
  </div>

  <!-- Stats Row -->
  <div class="stats-row">
    <div class="stat-card">
      <div class="stat-value">{{ allNumbers.length }}</div>
      <div class="stat-label">Total Numbers</div>
    </div>
    <div class="stat-card">
      <div class="stat-value">{{ inboundRoutes.length }}</div>
      <div class="stat-label">Inbound Routes</div>
    </div>
    <div class="stat-card">
      <div class="stat-value">{{ outboundRoutes.length }}</div>
      <div class="stat-label">Outbound Routes</div>
    </div>
    <div class="stat-card accent">
      <div class="stat-value">{{ enabledRoutesCount }}</div>
      <div class="stat-label">Active Routes</div>
    </div>
  </div>

  <div class="tabs">
    <button class="tab" :class="{ active: activeTab === 'numbers' }" @click="activeTab = 'numbers'">All Numbers</button>
    <button class="tab" :class="{ active: activeTab === 'inbound' }" @click="activeTab = 'inbound'">Global Inbound</button>
    <button class="tab" :class="{ active: activeTab === 'outbound' }" @click="activeTab = 'outbound'">Global Outbound</button>
    <button class="tab" :class="{ active: activeTab === 'settings' }" @click="activeTab = 'settings'">Global Settings</button>
  </div>

  <!-- NUMBERS TAB (System Number Pool) -->
  <div class="tab-content" v-if="activeTab === 'numbers'">
    <div class="sub-tabs">
      <button class="sub-tab" :class="{ active: numbersSubTab === 'numbers' }" @click="numbersSubTab = 'numbers'">All Numbers</button>
      <button class="sub-tab" :class="{ active: numbersSubTab === 'groups' }" @click="numbersSubTab = 'groups'">Number Groups</button>
    </div>

    <!-- Numbers Sub-Tab -->
    <div v-if="numbersSubTab === 'numbers'">
      <div class="route-help">
        <InfoIcon class="help-icon" />
        <span>System-managed phone numbers. Add numbers here and assign them to tenants. Tenants cannot add their own.</span>
      </div>

      <div class="filter-bar">
        <div class="search-box">
          <SearchIcon class="search-icon" />
          <input type="text" v-model="numberSearch" placeholder="Search numbers..." class="search-input">
        </div>
        <select v-model="statusFilter" class="filter-dropdown">
          <option value="">All Statuses</option>
          <option value="available">Available</option>
          <option value="assigned">Assigned</option>
          <option value="reserved">Reserved</option>
          <option value="porting">Porting</option>
        </select>
        <select v-model="tenantFilter" class="filter-dropdown">
          <option value="">All Tenants</option>
          <option v-for="t in tenants" :key="t.id" :value="t.id">{{ t.name }}</option>
        </select>
        <select v-model="groupFilter" class="filter-dropdown">
          <option value="">All Groups</option>
          <option v-for="g in numberGroups" :key="g.id" :value="g.id">{{ g.name }}</option>
        </select>
      </div>

      <DataTable :columns="numberColumns" :data="filteredNumbers" :actions="numberActions">
        <template #phone_number="{ value }">
          <span class="font-mono font-semibold">{{ formatPhoneNumber(value) }}</span>
        </template>
        <template #status="{ value }">
          <StatusBadge :status="value === 'assigned' ? 'Active' : value === 'available' ? 'Available' : value" />
        </template>
        <template #tenant="{ row }">
          <span class="tenant-badge" v-if="row.tenant">{{ row.tenant.name }}</span>
          <span class="text-muted" v-else>—</span>
        </template>
        <template #number_group="{ row }">
          <span class="badge gateway" v-if="row.number_group">{{ row.number_group.name }}</span>
          <span class="text-muted" v-else>—</span>
        </template>
        <template #capabilities="{ row }">
          <div class="cap-badges">
            <span class="cap-badge voice">Voice</span>
            <span class="cap-badge sms" v-if="row.sms_enabled">SMS</span>
            <span class="cap-badge mms" v-if="row.mms_enabled">MMS</span>
            <span class="cap-badge fax" v-if="row.fax_enabled">Fax</span>
            <span class="cap-badge e911" v-if="row.e911_eligible">E911</span>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- Number Groups Sub-Tab -->
    <div v-if="numbersSubTab === 'groups'">
      <div class="route-help">
        <InfoIcon class="help-icon" />
        <span>Number groups define outbound carrier routing and SMS provider. Each group has routing rules, gateway priorities, and messaging config.</span>
      </div>

      <div class="routes-list" v-if="numberGroups.length">
        <div class="route-card" v-for="group in numberGroups" :key="group.id">
          <div class="route-main">
            <div class="route-name-row">
              <h4>{{ group.name }}</h4>
              <div class="route-badges">
                <span class="badge context">{{ group.number_count || 0 }} numbers</span>
                <span class="badge gateway" v-if="group.default_gateway">{{ group.default_gateway.name || group.default_gateway.gateway_name }}</span>
              </div>
            </div>
            <p class="text-muted text-sm" v-if="group.description">{{ group.description }}</p>
            <div class="gw-priority-list" v-if="group.gateway_priorities && group.gateway_priorities.length">
              <span class="gw-priority-item" v-for="(gp, i) in group.gateway_priorities" :key="i">
                {{ i + 1 }}. {{ gp.gateway_name }} (P{{ gp.priority }}, W{{ gp.weight }})
              </span>
            </div>
            <div class="route-badges" style="margin-top: 4px">
              <span class="badge context" v-if="group.routing_rules && group.routing_rules.length">{{ group.routing_rules.length }} routing rules</span>
              <span class="badge gateway" v-if="group.messaging_provider">SMS: {{ group.messaging_provider.name }}</span>
            </div>
          </div>
          <div class="route-controls">
            <label class="switch small">
              <input type="checkbox" v-model="group.enabled">
              <span class="slider round"></span>
            </label>
            <button class="btn-icon" @click="editGroup(group)"><EditIcon class="icon-sm" /></button>
            <button class="btn-icon" @click="deleteGroup(group)"><TrashIcon class="icon-sm text-bad" /></button>
          </div>
        </div>
      </div>
      <div class="empty-state" v-else>
        <p>No number groups configured. Create one to organize outbound routing.</p>
      </div>
    </div>
  </div>

  <!-- INBOUND ROUTES TAB -->
  <div class="tab-content" v-if="activeTab === 'inbound'">
    <div class="route-help">
      <InfoIcon class="help-icon" />
      <span>Global inbound routes process calls that don't match specific tenant routes. First matching route wins.</span>
    </div>

    <div class="routes-list">
      <div class="route-card" v-for="(route, idx) in inboundRoutes" :key="route.id" 
        :class="{ disabled: !route.enabled, dragging: dragIndex === idx && dragType === 'inbound', dragover: dragOverIndex === idx && dragType === 'inbound' }"
        draggable="true"
        @dragstart="onDragStart($event, idx, 'inbound')"
        @dragover.prevent="onDragOver($event, idx, 'inbound')"
        @dragleave="onDragLeave()"
        @drop="onDrop($event, idx, 'inbound')"
        @dragend="onDragEnd()">
        <div class="route-handle">
          <GripVerticalIcon class="grip-icon" />
          <span class="route-order">{{ idx + 1 }}</span>
        </div>
        
        <div class="route-main">
          <div class="route-name-row">
            <h4>{{ route.name }}</h4>
            <div class="route-badges">
              <span class="badge context">{{ route.context }}</span>
            </div>
          </div>
          
          <div class="route-conditions">
            <div class="condition" v-for="(cond, i) in route.conditions" :key="i">
              <span class="cond-var">{{ cond.variable }}</span>
              <span class="cond-op">{{ cond.operator }}</span>
              <span class="cond-val font-mono">{{ cond.value }}</span>
            </div>
          </div>
          
          <div class="route-actions-display">
            <ArrowRightIcon class="arrow-icon" />
            <div class="action-chain">
              <span class="action-item" v-for="(act, i) in route.actions" :key="i">
                <span class="action-app">{{ act.app }}</span>
                <span class="action-data" v-if="act.data">{{ act.data }}</span>
              </span>
            </div>
          </div>
        </div>

        <div class="route-controls">
          <label class="switch small">
            <input type="checkbox" v-model="route.enabled">
            <span class="slider round"></span>
          </label>
          <button class="btn-icon" @click="editInboundRoute(route)"><EditIcon class="icon-sm" /></button>
          <button class="btn-icon" @click="deleteInboundRoute(route)"><TrashIcon class="icon-sm text-bad" /></button>
        </div>
      </div>
    </div>
  </div>

  <!-- OUTBOUND ROUTES TAB -->
  <div class="tab-content" v-else-if="activeTab === 'outbound'">
    <div class="route-help">
      <InfoIcon class="help-icon" />
      <span>Global outbound routes apply to all calls unless overridden by tenant specific routes.</span>
    </div>

    <div class="routes-list">
      <div class="route-card" v-for="(route, idx) in outboundRoutes" :key="route.id"
        :class="{ disabled: !route.enabled, dragging: dragIndex === idx && dragType === 'outbound', dragover: dragOverIndex === idx && dragType === 'outbound' }"
        draggable="true"
        @dragstart="onDragStart($event, idx, 'outbound')"
        @dragover.prevent="onDragOver($event, idx, 'outbound')"
        @dragleave="onDragLeave()"
        @drop="onDrop($event, idx, 'outbound')"
        @dragend="onDragEnd()">
        <div class="route-handle">
          <GripVerticalIcon class="grip-icon" />
          <span class="route-order">{{ idx + 1 }}</span>
        </div>
        
        <div class="route-main">
          <div class="route-name-row">
            <h4>{{ route.name }}</h4>
            <div class="route-badges">
              <span class="badge gateway">{{ route.gateway }}</span>
              <span class="badge intl" v-if="route.international">International</span>
            </div>
          </div>
          
          <div class="route-pattern">
            <span class="pattern-label">Pattern:</span>
            <code class="pattern-regex">{{ route.pattern }}</code>
            <span class="pattern-desc" v-if="route.description">{{ route.description }}</span>
          </div>
          
          <div class="route-transforms" v-if="route.prepend || route.strip">
            <span class="transform" v-if="route.strip">Strip: {{ route.strip }} digits</span>
            <span class="transform" v-if="route.prepend">Prepend: {{ route.prepend }}</span>
          </div>
        </div>

        <div class="route-controls">
          <label class="switch small">
            <input type="checkbox" v-model="route.enabled">
            <span class="slider round"></span>
          </label>
          <button class="btn-icon" @click="editOutboundRoute(route)"><EditIcon class="icon-sm" /></button>
          <button class="btn-icon" @click="deleteOutboundRoute(route)"><TrashIcon class="icon-sm text-bad" /></button>
        </div>
      </div>
    </div>
  </div>

  <!-- SETTINGS TAB -->
  <div class="tab-content settings-panel" v-else-if="activeTab === 'settings'">
    <div class="settings-section">
      <h3>System Dialing Defaults</h3>
      <div class="settings-grid">
        <div class="form-group">
          <label>Default Region</label>
          <select class="input-field" v-model="settings.region">
            <option value="nanp">North America (NANP)</option>
            <option value="uk">United Kingdom</option>
            <option value="eu">Europe (General)</option>
            <option value="au">Australia</option>
          </select>
          <span class="help-text">System-wide default for interpretation.</span>
        </div>
        <div class="form-group">
          <label>Default Outbound Format</label>
          <select class="input-field" v-model="settings.format">
            <option value="e164">E.164 (Global Standard, +1...)</option>
            <option value="national">National (10-digit)</option>
            <option value="passthrough">Passthrough (As Dialed)</option>
          </select>
        </div>
      </div>
    </div>

    <div class="form-actions">
      <button class="btn-primary" @click="saveSettings">Save Settings</button>
    </div>
  </div>

  <!-- INBOUND ROUTE MODAL -->
  <div v-if="showInboundModal" class="modal-overlay" @click.self="showInboundModal = false">
    <div class="modal-card large">
      <div class="modal-header">
        <h3>{{ editingInbound ? 'Edit Global Inbound Route' : 'New Global Inbound Route' }}</h3>
        <button class="btn-icon" @click="showInboundModal = false"><XIcon class="icon-sm" /></button>
      </div>
      
      <div class="modal-body">
        <div class="form-row">
          <div class="form-group flex-2">
            <label>Route Name</label>
            <input v-model="inboundForm.name" class="input-field" placeholder="e.g. Carrier Handover">
          </div>
          <div class="form-group">
            <label>Context</label>
            <input v-model="inboundForm.context" class="input-field code" placeholder="public">
          </div>
        </div>

        <div class="divider"></div>

        <div class="form-section">
          <div class="section-header">
            <h4>Conditions</h4>
            <button class="btn-small" @click="addInboundCondition">+ Add Condition</button>
          </div>
          
          <div class="conditions-editor">
            <div class="condition-row" v-for="(cond, i) in inboundForm.conditions" :key="i">
              <select v-model="cond.variable" class="input-field">
                <option value="destination_number">destination_number</option>
                <option value="caller_id_number">caller_id_number</option>
                <option value="network_addr">network_addr</option>
                <option value="sip_user_agent">sip_user_agent</option>
              </select>
              <select v-model="cond.operator" class="input-field small">
                <option value="=~">matches regex</option>
                <option value="==">equals</option>
                <option value="!=">not equals</option>
              </select>
              <input v-model="cond.value" class="input-field code" placeholder="^\\+?1?(\\d{10})$">
              <button class="btn-icon" @click="removeInboundCondition(i)"><XIcon class="icon-sm" /></button>
            </div>
          </div>
        </div>

        <div class="divider"></div>

        <div class="form-section">
          <div class="section-header">
            <h4>Actions</h4>
            <button class="btn-small" @click="addInboundAction">+ Add Action</button>
          </div>
          
          <div class="actions-editor">
            <div class="action-row" v-for="(act, i) in inboundForm.actions" :key="i">
              <select v-model="act.app" class="input-field">
                <optgroup label="Routing">
                  <option value="transfer">transfer</option>
                  <option value="bridge">bridge</option>
                </optgroup>
                <optgroup label="System">
                  <option value="log">log</option>
                  <option value="info">info</option>
                  <option value="system">system (exec)</option>
                </optgroup>
              </select>
              <input v-model="act.data" class="input-field flex-2" placeholder="application data">
              <button class="btn-icon" @click="removeInboundAction(i)"><XIcon class="icon-sm" /></button>
            </div>
          </div>
        </div>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" @click="showInboundModal = false">Cancel</button>
        <button class="btn-primary" @click="saveInboundRoute" :disabled="!inboundForm.name">Save Route</button>
      </div>
    </div>
  </div>

  <!-- OUTBOUND ROUTE MODAL -->
  <div v-if="showOutboundModal" class="modal-overlay" @click.self="showOutboundModal = false">
    <div class="modal-card">
      <div class="modal-header">
        <h3>{{ editingOutbound ? 'Edit Global Outbound Route' : 'New Global Outbound Route' }}</h3>
        <button class="btn-icon" @click="showOutboundModal = false"><XIcon class="icon-sm" /></button>
      </div>
      
      <div class="modal-body">
        <div class="form-group">
          <label>Route Name</label>
          <input v-model="outboundForm.name" class="input-field" placeholder="e.g. Emergency Override">
        </div>

        <div class="form-group">
          <label>Pattern (Regex)</label>
          <input v-model="outboundForm.pattern" class="input-field code" placeholder="^911$">
          <span class="help-text">Global regex match on dialed digits.</span>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>Strip Digits</label>
            <input type="number" v-model="outboundForm.strip" class="input-field" placeholder="0">
          </div>
          <div class="form-group">
            <label>Prepend</label>
            <input v-model="outboundForm.prepend" class="input-field code" placeholder="">
          </div>
        </div>

        <div class="form-group">
          <label>Gateway / Trunk</label>
          <select v-model="outboundForm.gateway" class="input-field">
            <option value="">Select System Gateway...</option>
            <option v-for="gw in gateways" :key="gw.id" :value="gw.gateway_name">{{ gw.name || gw.gateway_name }}</option>
          </select>
        </div>

        <div class="checkbox-group">
          <label class="checkbox-row">
            <input type="checkbox" v-model="outboundForm.continue">
            <span>Continue to tenant routes on failure</span>
          </label>
        </div>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" @click="showOutboundModal = false">Cancel</button>
        <button class="btn-primary" @click="saveOutboundRoute" :disabled="!outboundForm.name || !outboundForm.pattern">Save Route</button>
      </div>
    </div>
  </div>

  <!-- ADD/EDIT NUMBER MODAL -->
  <div v-if="showNumberModal" class="modal-overlay" @click.self="showNumberModal = false">
    <div class="modal-card">
      <div class="modal-header">
        <h3>{{ editingNumber ? 'Edit System Number' : 'Add System Number' }}</h3>
        <button class="btn-icon" @click="showNumberModal = false"><XIcon class="icon-sm" /></button>
      </div>

      <div class="modal-body">
        <div class="form-group">
          <label>Phone Number (E.164)</label>
          <input v-model="numberForm.phone_number" class="input-field code" placeholder="+14155551234" :disabled="editingNumber">
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>Caller ID Name</label>
            <input v-model="numberForm.caller_id_name" class="input-field" placeholder="Company Name">
          </div>
          <div class="form-group">
            <label>Status</label>
            <select v-model="numberForm.status" class="input-field">
              <option value="available">Available</option>
              <option value="assigned">Assigned</option>
              <option value="reserved">Reserved</option>
              <option value="porting">Porting</option>
            </select>
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>Number Group</label>
            <select v-model="numberForm.number_group_id" class="input-field">
              <option :value="null">None</option>
              <option v-for="g in numberGroups" :key="g.id" :value="g.id">{{ g.name }}</option>
            </select>
          </div>
          <div class="form-group">
            <label>Assign to Tenant</label>
            <select v-model="numberForm.tenant_id" class="input-field">
              <option :value="null">Unassigned</option>
              <option v-for="t in tenants" :key="t.id" :value="t.id">{{ t.name }}</option>
            </select>
          </div>
        </div>

        <div class="form-group">
          <label>Description</label>
          <input v-model="numberForm.description" class="input-field" placeholder="Main office line">
        </div>

        <div class="checkbox-group">
          <label class="checkbox-row"><input type="checkbox" v-model="numberForm.sms_enabled"><span>SMS Capable</span></label>
          <label class="checkbox-row"><input type="checkbox" v-model="numberForm.mms_enabled"><span>MMS Capable</span></label>
          <label class="checkbox-row"><input type="checkbox" v-model="numberForm.fax_enabled"><span>Fax Capable</span></label>
          <label class="checkbox-row"><input type="checkbox" v-model="numberForm.e911_eligible"><span>E911 Eligible</span></label>
          <label class="checkbox-row"><input type="checkbox" v-model="numberForm.enabled"><span>Enabled</span></label>
        </div>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" @click="showNumberModal = false">Cancel</button>
        <button class="btn-primary" @click="saveNumber" :disabled="!numberForm.phone_number">{{ editingNumber ? 'Update' : 'Add Number' }}</button>
      </div>
    </div>
  </div>

  <!-- ASSIGN TENANT MODAL -->
  <div v-if="showAssignModal" class="modal-overlay" @click.self="showAssignModal = false">
    <div class="modal-card small">
      <div class="modal-header">
        <h3>Assign to Tenant</h3>
        <button class="btn-icon" @click="showAssignModal = false"><XIcon class="icon-sm" /></button>
      </div>
      <div class="modal-body">
        <p class="text-sm">Assign <strong>{{ formatPhoneNumber(assigningNumber?.phone_number) }}</strong> to:</p>
        <div class="form-group">
          <label>Tenant</label>
          <select v-model="assignTenantId" class="input-field">
            <option v-for="t in tenants" :key="t.id" :value="t.id">{{ t.name }}</option>
          </select>
        </div>
      </div>
      <div class="modal-actions">
        <button class="btn-secondary" @click="showAssignModal = false">Cancel</button>
        <button class="btn-primary" @click="confirmAssign" :disabled="!assignTenantId">Assign</button>
      </div>
    </div>
  </div>

  <!-- NUMBER GROUP MODAL (expanded with tabs) -->
  <div v-if="showGroupModal" class="modal-overlay" @click.self="showGroupModal = false">
    <div class="modal-card large">
      <div class="modal-header">
        <h3>{{ editingGroup ? 'Edit Number Group' : 'New Number Group' }}</h3>
        <button class="btn-icon" @click="showGroupModal = false"><XIcon class="icon-sm" /></button>
      </div>

      <!-- Group Modal Tabs -->
      <div class="sub-tabs" style="padding: 0 1.5rem">
        <button class="sub-tab" :class="{ active: groupModalTab === 'general' }" @click="groupModalTab = 'general'">General</button>
        <button class="sub-tab" :class="{ active: groupModalTab === 'gateways' }" @click="groupModalTab = 'gateways'">Gateway Priorities</button>
        <button class="sub-tab" :class="{ active: groupModalTab === 'rules' }" @click="groupModalTab = 'rules'">Routing Rules</button>
      </div>

      <div class="modal-body">
        <!-- GENERAL TAB -->
        <div v-if="groupModalTab === 'general'">
          <div class="form-row">
            <div class="form-group flex-2">
              <label>Group Name</label>
              <input v-model="groupForm.name" class="input-field" placeholder="e.g. US Domestic">
            </div>
            <div class="form-group">
              <label>Default Gateway</label>
              <select v-model="groupForm.default_gateway_id" class="input-field">
                <option :value="null">None</option>
                <option v-for="gw in gateways" :key="gw.id" :value="gw.id">{{ gw.name || gw.gateway_name }}</option>
              </select>
            </div>
          </div>

          <div class="form-group">
            <label>Description</label>
            <input v-model="groupForm.description" class="input-field" placeholder="Routes for domestic calls">
          </div>

          <div class="form-group">
            <label>SMS / Messaging Provider</label>
            <select v-model="groupForm.messaging_provider_id" class="input-field">
              <option :value="null">None</option>
              <option v-for="mp in messagingProviders" :key="mp.id" :value="mp.id">{{ mp.name }}</option>
            </select>
            <span class="help-text">Default messaging provider for all numbers in this group. Individual numbers can override.</span>
          </div>

          <div class="checkbox-group">
            <label class="checkbox-row"><input type="checkbox" v-model="groupForm.enabled"><span>Enabled</span></label>
          </div>
        </div>

        <!-- GATEWAY PRIORITIES TAB -->
        <div v-if="groupModalTab === 'gateways'">
          <div class="form-section">
            <div class="section-header">
              <h4>Gateway Priorities (failover order)</h4>
              <button class="btn-small" @click="addGatewayPriority">+ Add Gateway</button>
            </div>
            <div class="gw-editor">
              <div class="gw-row" v-for="(gp, i) in groupForm.gateway_priorities" :key="i">
                <span class="gw-order">{{ i + 1 }}</span>
                <select v-model="gp.gateway_id" class="input-field" @change="updateGatewayName(gp)">
                  <option v-for="gw in gateways" :key="gw.id" :value="gw.id">{{ gw.name || gw.gateway_name }}</option>
                </select>
                <input type="number" v-model.number="gp.priority" class="input-field small" placeholder="Priority">
                <input type="number" v-model.number="gp.weight" class="input-field small" placeholder="Weight">
                <button class="btn-icon" @click="groupForm.gateway_priorities.splice(i, 1)"><XIcon class="icon-sm" /></button>
              </div>
            </div>
          </div>
        </div>

        <!-- ROUTING RULES TAB -->
        <div v-if="groupModalTab === 'rules'">
          <div class="route-help" style="margin-bottom: 12px">
            <InfoIcon class="help-icon" />
            <span>Regex-based routing rules control how outbound calls are processed. Lower priority = evaluated first.</span>
          </div>

          <!-- Preset Buttons -->
          <div class="preset-bar" v-if="!editingRule">
            <button class="btn-small" @click="addRulePreset('us_domestic')">+ US Domestic</button>
            <button class="btn-small" @click="addRulePreset('international')">+ International</button>
            <button class="btn-small" @click="addRulePreset('emergency')">+ Emergency</button>
            <button class="btn-small" @click="addRulePreset('toll_free')">+ Toll Free</button>
            <button class="btn-small" @click="addRulePreset('custom')">+ Custom Rule</button>
          </div>

          <!-- Rules List -->
          <div class="routes-list" v-if="groupRoutingRules.length && !editingRule" style="margin-top: 12px">
            <div class="route-card compact" v-for="rule in groupRoutingRules" :key="rule.id" :class="{ disabled: !rule.enabled }">
              <div class="route-main">
                <div class="route-name-row">
                  <h4>{{ rule.name }}</h4>
                  <div class="route-badges">
                    <span class="badge context">P{{ rule.priority }}</span>
                    <span class="badge gateway" v-if="rule.gateway_name">{{ rule.gateway_name }}</span>
                    <span class="badge intl" v-if="rule.dial_format !== 'e164'">{{ rule.dial_format }}</span>
                  </div>
                </div>
                <div class="route-pattern">
                  <span class="pattern-label">Pattern:</span>
                  <code class="pattern-regex">{{ rule.pattern }}</code>
                </div>
                <div class="route-transforms" v-if="rule.prepend || rule.strip_digits">
                  <span class="transform" v-if="rule.strip_digits">Strip: {{ rule.strip_digits }}</span>
                  <span class="transform" v-if="rule.prepend">Prepend: {{ rule.prepend }}</span>
                  <span class="transform" v-if="rule.prefix">Prefix: {{ rule.prefix }}</span>
                </div>
              </div>
              <div class="route-controls">
                <label class="switch small">
                  <input type="checkbox" v-model="rule.enabled">
                  <span class="slider round"></span>
                </label>
                <button class="btn-icon" @click="editRoutingRule(rule)"><EditIcon class="icon-sm" /></button>
                <button class="btn-icon" @click="deleteRoutingRule(rule)"><TrashIcon class="icon-sm text-bad" /></button>
              </div>
            </div>
          </div>

          <div class="empty-state" v-if="!groupRoutingRules.length && !editingRule">
            <p>No routing rules. Add one using the presets above or create a custom rule.</p>
          </div>

          <!-- Rule Editor (inline) -->
          <div class="rule-editor" v-if="editingRule">
            <h4 style="margin-bottom: 12px">{{ ruleForm.id ? 'Edit Rule' : 'New Rule' }}</h4>
            <div class="form-row">
              <div class="form-group flex-2">
                <label>Rule Name</label>
                <input v-model="ruleForm.name" class="input-field" placeholder="e.g. US Domestic 10-digit">
              </div>
              <div class="form-group">
                <label>Priority</label>
                <input type="number" v-model.number="ruleForm.priority" class="input-field" placeholder="100">
              </div>
            </div>

            <div class="form-group">
              <label>Pattern (Regex)</label>
              <input v-model="ruleForm.pattern" class="input-field code" placeholder="^\\+?1?(\\d{10})$">
              <span class="help-text">Regex matched against the dialed number. Use capture groups for transforms.</span>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label>Match Field</label>
                <select v-model="ruleForm.match_field" class="input-field">
                  <option value="destination_number">Destination Number</option>
                  <option value="caller_id_number">Caller ID Number</option>
                </select>
              </div>
              <div class="form-group">
                <label>Dial Format</label>
                <select v-model="ruleForm.dial_format" class="input-field">
                  <option value="e164">E.164 (+1NXXXXXXXXX)</option>
                  <option value="11d">11 Digit (1NXXXXXXXXX)</option>
                  <option value="10d">10 Digit (NXXXXXXXXX)</option>
                  <option value="custom">Custom</option>
                </select>
              </div>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label>Strip Leading Digits</label>
                <input type="number" v-model.number="ruleForm.strip_digits" class="input-field" min="0">
              </div>
              <div class="form-group">
                <label>Prepend</label>
                <input v-model="ruleForm.prepend" class="input-field code" placeholder="">
              </div>
              <div class="form-group">
                <label>Prefix</label>
                <input v-model="ruleForm.prefix" class="input-field code" placeholder="e.g. 011">
              </div>
            </div>

            <div class="form-group">
              <label>Gateway / Trunk</label>
              <select v-model="ruleForm.gateway_id" class="input-field" @change="updateRuleGatewayName">
                <option :value="null">Use Group Default</option>
                <option v-for="gw in gateways" :key="gw.id" :value="gw.id">{{ gw.name || gw.gateway_name }}</option>
              </select>
            </div>

            <div class="checkbox-group">
              <label class="checkbox-row"><input type="checkbox" v-model="ruleForm.continue_on_fail"><span>Continue to next rule on failure</span></label>
              <label class="checkbox-row"><input type="checkbox" v-model="ruleForm.enabled"><span>Enabled</span></label>
            </div>

            <div class="form-actions" style="margin-top: 12px">
              <button class="btn-secondary" @click="editingRule = false">Cancel</button>
              <button class="btn-primary" @click="saveRoutingRule" :disabled="!ruleForm.name || !ruleForm.pattern">{{ ruleForm.id ? 'Update Rule' : 'Add Rule' }}</button>
            </div>
          </div>
        </div>
      </div>

      <div class="modal-actions" v-if="!editingRule">
        <button class="btn-secondary" @click="showGroupModal = false">Cancel</button>
        <button class="btn-primary" @click="saveGroup" :disabled="!groupForm.name">{{ editingGroup ? 'Update' : 'Create Group' }}</button>
      </div>
    </div>
  </div>

</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { 
  Search as SearchIcon, Info as InfoIcon, GripVertical as GripVerticalIcon,
  ArrowRight as ArrowRightIcon, Edit as EditIcon,
  Trash2 as TrashIcon, X as XIcon, RefreshCw as RefreshCwIcon,
  UserPlus as UserPlusIcon, UserMinus as UserMinusIcon
} from 'lucide-vue-next'
import DataTable from '../../components/common/DataTable.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'
import { formatPhoneNumber } from '../../utils/formatters'
import { systemAPI } from '../../services/api'

const activeTab = ref('numbers')

// Drag and Drop State
const dragIndex = ref(null)
const dragOverIndex = ref(null)
const dragType = ref(null)

const onDragStart = (e, idx, type) => {
  dragIndex.value = idx
  dragType.value = type
  e.dataTransfer.effectAllowed = 'move'
}

const onDragOver = (e, idx, type) => {
  if (dragType.value !== type) return
  dragOverIndex.value = idx
  e.dataTransfer.dropEffect = 'move'
}

const onDragLeave = () => {
  dragOverIndex.value = null
}

const onDrop = (e, idx, type) => {
  if (dragType.value !== type || dragIndex.value === null) return
  
  const routes = type === 'inbound' ? inboundRoutes : outboundRoutes
  const fromIdx = dragIndex.value
  const toIdx = idx
  
  if (fromIdx !== toIdx) {
    const item = routes.value.splice(fromIdx, 1)[0]
    routes.value.splice(toIdx, 0, item)
  }
  
  dragIndex.value = null
  dragOverIndex.value = null
  dragType.value = null
}

const onDragEnd = () => {
  dragIndex.value = null
  dragOverIndex.value = null
  dragType.value = null
}

// =====================
// System Numbers
// =====================
const allNumbers = ref([])
const numberSearch = ref('')
const statusFilter = ref('')
const tenantFilter = ref('')
const groupFilter = ref('')
const numbersSubTab = ref('numbers')
const tenants = ref([])
const numberGroups = ref([])

const numberColumns = [
  { key: 'phone_number', label: 'Number', width: '160px' },
  { key: 'tenant', label: 'Tenant', width: '140px' },
  { key: 'number_group', label: 'Group', width: '120px' },
  { key: 'capabilities', label: 'Capabilities', width: '180px' },
  { key: 'status', label: 'Status', width: '100px' },
  { key: 'caller_id_name', label: 'Caller ID', width: '140px' },
]

const numberActions = [
  { label: 'Edit', icon: EditIcon, handler: (row) => editNumber(row) },
  { label: 'Assign', icon: UserPlusIcon, handler: (row) => openAssign(row), show: (row) => !row.tenant_id },
  { label: 'Unassign', icon: UserMinusIcon, handler: (row) => unassignNum(row), show: (row) => !!row.tenant_id },
  { label: 'Delete', icon: TrashIcon, handler: (row) => deleteNumber(row), className: 'text-bad' },
]

const filteredNumbers = computed(() => {
  return allNumbers.value.filter(n => {
    const matchesSearch = !numberSearch.value || (n.phone_number || '').includes(numberSearch.value)
    const matchesTenant = !tenantFilter.value || n.tenant_id === parseInt(tenantFilter.value)
    const matchesStatus = !statusFilter.value || n.status === statusFilter.value
    const matchesGroup = !groupFilter.value || n.number_group_id === parseInt(groupFilter.value)
    return matchesSearch && matchesTenant && matchesStatus && matchesGroup
  })
})

const loadNumbers = async () => {
  try {
    const response = await systemAPI.listSystemNumbers()
    allNumbers.value = response.data?.data || response.data || []
  } catch (e) {
    console.error('Failed to load system numbers', e)
  }
}

const loadTenants = async () => {
  try {
    const response = await systemAPI.listTenants()
    tenants.value = response.data?.data || response.data || []
  } catch (e) {
    console.error('Failed to load tenants', e)
  }
}

const loadNumberGroups = async () => {
  try {
    const response = await systemAPI.listNumberGroups()
    numberGroups.value = response.data?.data || response.data || []
  } catch (e) {
    console.error('Failed to load number groups', e)
  }
}

// Number CRUD
const showNumberModal = ref(false)
const editingNumber = ref(false)
const numberForm = ref({
  phone_number: '', caller_id_name: '', description: '',
  number_group_id: null, tenant_id: null,
  sms_enabled: false, mms_enabled: false, fax_enabled: false,
  e911_eligible: true, enabled: true, status: 'available'
})

const openAddNumber = () => {
  numberForm.value = {
    phone_number: '', caller_id_name: '', description: '',
    number_group_id: null, tenant_id: null,
    sms_enabled: false, mms_enabled: false, fax_enabled: false,
    e911_eligible: true, enabled: true, status: 'available'
  }
  editingNumber.value = false
  showNumberModal.value = true
}

const editNumber = (row) => {
  numberForm.value = { ...row }
  editingNumber.value = true
  showNumberModal.value = true
}

const saveNumber = async () => {
  try {
    if (editingNumber.value) {
      await systemAPI.updateSystemNumber(numberForm.value.id, numberForm.value)
    } else {
      await systemAPI.createSystemNumber(numberForm.value)
    }
    showNumberModal.value = false
    await loadNumbers()
  } catch (e) {
    alert('Failed to save number: ' + (e.message || 'Unknown error'))
  }
}

const deleteNumber = async (row) => {
  if (confirm(`Delete number ${row.phone_number}? This cannot be undone.`)) {
    try {
      await systemAPI.deleteSystemNumber(row.id)
      await loadNumbers()
    } catch (e) {
      alert('Failed to delete number')
    }
  }
}

// Assignment
const showAssignModal = ref(false)
const assigningNumber = ref(null)
const assignTenantId = ref(null)

const openAssign = (row) => {
  assigningNumber.value = row
  assignTenantId.value = null
  showAssignModal.value = true
}

const confirmAssign = async () => {
  try {
    await systemAPI.assignNumber(assigningNumber.value.id, { tenant_id: assignTenantId.value })
    showAssignModal.value = false
    await loadNumbers()
  } catch (e) {
    alert('Failed to assign number: ' + (e.message || 'Unknown error'))
  }
}

const unassignNum = async (row) => {
  if (confirm(`Unassign ${row.phone_number} from ${row.tenant?.name || 'tenant'}?`)) {
    try {
      await systemAPI.unassignNumber(row.id)
      await loadNumbers()
    } catch (e) {
      alert('Failed to unassign number')
    }
  }
}

// Number Groups
const showGroupModal = ref(false)
const editingGroup = ref(false)
const groupForm = ref({
  name: '', description: '', enabled: true,
  default_gateway_id: null, gateway_priorities: [],
  messaging_provider_id: null
})
const groupModalTab = ref('general')
const messagingProviders = ref([])
const groupRoutingRules = ref([])
const editingRule = ref(false)
const ruleForm = ref({
  name: '', pattern: '', match_field: 'destination_number',
  priority: 100, weight: 1, strip_digits: 0, prepend: '', prefix: '',
  dial_format: 'e164', gateway_id: null, gateway_name: '',
  continue_on_fail: true, enabled: true
})

const openAddGroup = () => {
  groupForm.value = { name: '', description: '', enabled: true, default_gateway_id: null, gateway_priorities: [], messaging_provider_id: null }
  groupModalTab.value = 'general'
  groupRoutingRules.value = []
  editingRule.value = false
  editingGroup.value = false
  showGroupModal.value = true
}

const editGroup = async (group) => {
  groupForm.value = { ...group, gateway_priorities: [...(group.gateway_priorities || [])] }
  groupModalTab.value = 'general'
  editingRule.value = false
  editingGroup.value = true
  showGroupModal.value = true
  // Load routing rules for this group
  await loadGroupRoutingRules(group.id)
}

const saveGroup = async () => {
  try {
    if (editingGroup.value) {
      await systemAPI.updateNumberGroup(groupForm.value.id, groupForm.value)
    } else {
      await systemAPI.createNumberGroup(groupForm.value)
    }
    showGroupModal.value = false
    await loadNumberGroups()
  } catch (e) {
    alert('Failed to save group: ' + (e.message || 'Unknown error'))
  }
}

const deleteGroup = async (group) => {
  if (confirm(`Delete group "${group.name}"?`)) {
    try {
      await systemAPI.deleteNumberGroup(group.id)
      await loadNumberGroups()
    } catch (e) {
      alert('Failed to delete group')
    }
  }
}

const addGatewayPriority = () => {
  groupForm.value.gateway_priorities.push({ gateway_id: null, gateway_name: '', priority: (groupForm.value.gateway_priorities.length + 1) * 10, weight: 1 })
}

const updateGatewayName = (gp) => {
  const gw = gateways.value.find(g => g.id === gp.gateway_id)
  if (gw) gp.gateway_name = gw.name || gw.gateway_name
}

// Messaging Providers
const loadMessagingProviders = async () => {
  try {
    const response = await systemAPI.listMessagingProviders()
    messagingProviders.value = response.data?.data || response.data || []
  } catch (e) {
    console.error('Failed to load messaging providers', e)
  }
}

// Routing Rules
const loadGroupRoutingRules = async (groupId) => {
  try {
    const response = await systemAPI.listRoutingRules(groupId)
    groupRoutingRules.value = response.data?.data || response.data || []
  } catch (e) {
    console.error('Failed to load routing rules', e)
    groupRoutingRules.value = []
  }
}

const rulePresets = {
  us_domestic: { name: 'US Domestic', pattern: '^\\+?1?(\\d{10})$', dial_format: 'e164', priority: 100, strip_digits: 0, prepend: '+1' },
  international: { name: 'International', pattern: '^\\+(?!1)(\\d+)$', dial_format: 'e164', priority: 200, strip_digits: 0, prepend: '' },
  emergency: { name: 'Emergency', pattern: '^(911|933)$', dial_format: 'e164', priority: 10, strip_digits: 0, prepend: '' },
  toll_free: { name: 'Toll Free', pattern: '^\\+?1?(8[0-9]{2}\\d{7})$', dial_format: 'e164', priority: 90, strip_digits: 0, prepend: '+1' },
  custom: { name: '', pattern: '', dial_format: 'e164', priority: 100, strip_digits: 0, prepend: '' }
}

const addRulePreset = (presetKey) => {
  const preset = rulePresets[presetKey]
  ruleForm.value = {
    name: preset.name, pattern: preset.pattern, match_field: 'destination_number',
    priority: preset.priority, weight: 1, strip_digits: preset.strip_digits,
    prepend: preset.prepend, prefix: '', dial_format: preset.dial_format,
    gateway_id: null, gateway_name: '', continue_on_fail: true, enabled: true
  }
  editingRule.value = true
}

const editRoutingRule = (rule) => {
  ruleForm.value = { ...rule }
  editingRule.value = true
}

const updateRuleGatewayName = () => {
  const gw = gateways.value.find(g => g.id === ruleForm.value.gateway_id)
  ruleForm.value.gateway_name = gw ? (gw.name || gw.gateway_name) : ''
}

const saveRoutingRule = async () => {
  try {
    const groupId = groupForm.value.id
    if (ruleForm.value.id) {
      await systemAPI.updateRoutingRule(groupId, ruleForm.value.id, ruleForm.value)
    } else {
      await systemAPI.createRoutingRule(groupId, ruleForm.value)
    }
    editingRule.value = false
    await loadGroupRoutingRules(groupId)
    await loadNumberGroups()
  } catch (e) {
    alert('Failed to save routing rule: ' + (e.message || 'Unknown error'))
  }
}

const deleteRoutingRule = async (rule) => {
  if (confirm(`Delete rule "${rule.name}"?`)) {
    try {
      await systemAPI.deleteRoutingRule(groupForm.value.id, rule.id)
      await loadGroupRoutingRules(groupForm.value.id)
      await loadNumberGroups()
    } catch (e) {
      alert('Failed to delete routing rule')
    }
  }
}

// Computed: Total enabled routes
const enabledRoutesCount = computed(() => {
  const inEnabled = inboundRoutes.value.filter(r => r.enabled).length
  const outEnabled = outboundRoutes.value.filter(r => r.enabled).length
  return inEnabled + outEnabled
})

// Auto-recalculate dialplan order based on specificity
const recalculateOrder = async () => {
  // Sort by specificity: exact matches first, then prefixes, then wildcards
  const sortBySpecificity = (routes) => {
    return [...routes].sort((a, b) => {
      const aPattern = a.conditions?.[0]?.value || ''
      const bPattern = b.conditions?.[0]?.value || ''
      const aIsExact = !aPattern.includes('*') && !aPattern.includes('.')
      const bIsExact = !bPattern.includes('*') && !bPattern.includes('.')
      if (aIsExact && !bIsExact) return -1
      if (!aIsExact && bIsExact) return 1
      return bPattern.length - aPattern.length
    })
  }
  
  inboundRoutes.value = sortBySpecificity(inboundRoutes.value)
  outboundRoutes.value = sortBySpecificity(outboundRoutes.value)

  // Persist the new order to the backend
  try {
    const allRoutes = [...inboundRoutes.value, ...outboundRoutes.value]
    for (let i = 0; i < allRoutes.length; i++) {
      const route = allRoutes[i]
      if (route.id) {
        await systemAPI.updateDialplan(route.id, { dialplan_order: (i + 1) * 10 })
      }
    }
  } catch (e) {
    console.error('Failed to persist route order', e)
  }
}

const loadAllNumbers = async () => {
  await Promise.all([loadNumbers(), loadTenants(), loadNumberGroups()])
}

// Inbound Routes (Global)
const showInboundModal = ref(false)
const editingInbound = ref(false)
const inboundForm = ref({
  dialplan_name: '',
  dialplan_context: 'public',
  enabled: true,
  details: [{ detail_type: 'condition', condition_field: 'destination_number', condition_expression: '', condition_expression_type: 'regex', condition_break: 'on-false' }, { detail_type: 'action', action_application: 'transfer', action_data: '' }]
})

const inboundRoutes = ref([])
const outboundRoutes = ref([])

const loadRoutes = async () => {
  try {
    const response = await systemAPI.listDialplans()
    const allRoutes = response.data.data || []
    
    // Map API format to UI format if needed, but lets try to align them
    // UI expects: name, context, enabled, conditions[], actions[]
    // API returns: Dialplan { Details: [] }
    
    // Helper to map DB model to UI model
    const mapRoute = (r) => {
      const conditions = r.Details.filter(d => d.detail_type === 'condition').map(d => ({
        variable: d.condition_field,
        operator: d.condition_expression_type === 'regex' ? '=~' : (d.condition_expression_type === 'negate' ? '!=' : '=='), // simplified assumptions for now
        value: d.condition_expression
      }))
      const actions = r.Details.filter(d => d.detail_type === 'action').map(d => ({
        app: d.action_application,
        data: d.action_data
      }))
      // If no conditions/actions found (e.g. empty), provide defaults
      if (conditions.length === 0) conditions.push({ variable: 'destination_number', operator: '=~', value: '' })
      if (actions.length === 0) actions.push({ app: 'transfer', data: '' })

      return {
        id: r.id,
        uuid: r.uuid,
        name: r.dialplan_name,
        context: r.dialplan_context,
        enabled: r.enabled,
        conditions,
        actions,
        // Keep original for updates
        _original: r
      }
    }

    inboundRoutes.value = allRoutes.filter(r => r.dialplan_context === 'public').map(mapRoute)
    outboundRoutes.value = allRoutes.filter(r => r.dialplan_context !== 'public').map(mapRoute)

  } catch (e) {
    console.error('Failed to load system routes', e)
  }
}

const addInboundCondition = () => inboundForm.value.details.push({ detail_type: 'condition', condition_field: 'destination_number', condition_expression: '', condition_expression_type: 'regex', condition_break: 'on-false' })
const removeInboundCondition = (i) => {
  const indexToRemove = inboundForm.value.details.findIndex(d => d.detail_type === 'condition' && inboundForm.value.details.indexOf(d) === i);
  if (indexToRemove !== -1) {
    inboundForm.value.details.splice(indexToRemove, 1);
  }
}
const addInboundAction = () => inboundForm.value.details.push({ detail_type: 'action', action_application: 'log', action_data: 'INFO' })
const removeInboundAction = (i) => {
  const indexToRemove = inboundForm.value.details.findIndex(d => d.detail_type === 'action' && inboundForm.value.details.indexOf(d) === i);
  if (indexToRemove !== -1) {
    inboundForm.value.details.splice(indexToRemove, 1);
  }
}

const editInboundRoute = (route) => {
  // Map UI object back to Form object (which should match API model structure ideally)
  const details = []
  route.conditions.forEach(c => details.push({ 
    detail_type: 'condition', 
    condition_field: c.variable, 
    condition_expression: c.value, 
    condition_expression_type: c.operator === '=~' ? 'regex' : (c.operator === '!=' ? 'negate' : 'exact'),
    condition_break: 'on-false' 
  }))
  route.actions.forEach(a => details.push({ 
    detail_type: 'action', 
    action_application: a.app, 
    action_data: a.data 
  }))

  inboundForm.value = {
    id: route.id,
    dialplan_name: route.name,
    dialplan_context: route.context,
    enabled: route.enabled,
    details: details
  }
  editingInbound.value = true
  showInboundModal.value = true
}

const deleteInboundRoute = async (route) => {
  if (confirm(`Delete global route "${route.name}"?`)) {
    try {
      await systemAPI.deleteDialplan(route.id)
      await loadRoutes()
    } catch (e) {
      console.error(e)
      alert('Failed to delete route')
    }
  }
}

const saveInboundRoute = async () => {
    try {
        const payload = {
            dialplan_name: inboundForm.value.dialplan_name,
            dialplan_context: inboundForm.value.dialplan_context,
            enabled: inboundForm.value.enabled,
            details: inboundForm.value.details.map((d, i) => ({ ...d, detail_order: (i+1)*10 }))
        }

        if (editingInbound.value) {
            await systemAPI.updateDialplan(inboundForm.value.id, payload)
        } else {
            await systemAPI.createDialplan(payload)
        }
        await loadRoutes()
        showInboundModal.value = false
        editingInbound.value = false
        inboundForm.value = { dialplan_name: '', dialplan_context: 'public', enabled: true, details: [{ detail_type: 'condition', condition_field: 'destination_number', condition_expression: '', condition_expression_type: 'regex', condition_break: 'on-false' }, { detail_type: 'action', action_application: 'transfer', action_data: '' }] }
    } catch (e) {
        console.error(e)
        alert('Failed to save route')
    }
}

// Outbound Routes (Global)
const showOutboundModal = ref(false)
const editingOutbound = ref(false)
const outboundForm = ref({ name: '', dialplan_context: 'default', enabled: true, pattern: '', strip: 0, prepend: '', gateway: '', continue: true, details: [] })
// Helper for outbound simplified UI (Pattern, Strip, Prepend, Gateway) -> Dialplan Details
// This logic is complex because 'outbound' UI is an abstraction over raw dialplan conditions/actions.
// For now, I'll stick to a simpler implementation or reuse the raw editor if the UI allows.
// Looking at the code, `SystemRoutes.vue` has a specific outbound modal with pattern/strip/prepend/gateway fields.
// I need to map these to Dialplan Details on save, and map back on edit.

const gateways = ref([])

const loadGateways = async () => {
  try {
    const response = await systemAPI.listGateways()
    gateways.value = response.data || []
  } catch (e) {
    console.error('Failed to load gateways', e)
  }
}

const editOutboundRoute = (route) => {
    // Reverse-map from dialplan conditions/actions to simplified outbound form
    const conds = route.conditions || []
    const acts = route.actions || []
    
    // Extract pattern from first condition
    const pattern = conds.length > 0 ? conds[0].value : ''
    
    // Extract gateway, strip, prepend from the bridge action
    let gateway = ''
    let strip = 0
    let prepend = ''
    let continueOnFail = true
    
    const bridgeAction = acts.find(a => a.app === 'bridge')
    if (bridgeAction && bridgeAction.data) {
        // Parse bridge string: sofia/gateway/NAME/PREPEND$1
        const gwMatch = bridgeAction.data.match(/sofia\/gateway\/([^\/]+)/)
        if (gwMatch) gateway = gwMatch[1]
    }
    
    // Look for set actions for strip/prepend
    acts.forEach(a => {
        if (a.app === 'set' && a.data) {
            if (a.data.includes('effective_caller_id')) return
            if (a.data.startsWith('sip_h_X-Prepend=')) prepend = a.data.split('=')[1] || ''
        }
    })
    
    // Check for strip_digits condition data
    conds.forEach(c => {
        if (c.variable === 'strip_digits' && c.value) {
            strip = parseInt(c.value) || 0
        }
    })
    
    outboundForm.value = {
        id: route.id,
        name: route.name,
        pattern: pattern,
        strip: strip,
        prepend: prepend,
        gateway: gateway,
        continue: continueOnFail,
        enabled: route.enabled
    }
    editingOutbound.value = true
    showOutboundModal.value = true
}

const deleteOutboundRoute = async (route) => {
    if (confirm(`Delete global route "${route.name}"?`)) {
        try {
            await systemAPI.deleteDialplan(route.id)
            await loadRoutes()
        } catch (e) {
            console.error(e)
        }
    }
}

const saveOutboundRoute = async () => {
    try {
        const form = outboundForm.value
        
        // Build dialplan details from simplified form
        const details = []
        
        // Condition: destination_number matches pattern
        details.push({
            detail_type: 'condition',
            condition_field: 'destination_number',
            condition_expression: form.pattern,
            condition_expression_type: 'regex',
            condition_break: 'on-false',
            detail_order: 10
        })
        
        // Action: set strip digits if needed
        if (form.strip && parseInt(form.strip) > 0) {
            details.push({
                detail_type: 'action',
                action_application: 'set',
                action_data: `effective_caller_id_number=\${caller_id_number:${form.strip}}`,
                detail_order: 20
            })
        }
        
        // Action: bridge to gateway
        let bridgeStr = `sofia/gateway/${form.gateway}/`
        if (form.prepend) bridgeStr += form.prepend
        bridgeStr += '$1'
        
        details.push({
            detail_type: 'action',
            action_application: 'bridge',
            action_data: bridgeStr,
            detail_order: 30
        })
        
        // Set continue on fail
        if (form.continue) {
            details.splice(details.length - 1, 0, {
                detail_type: 'action',
                action_application: 'set',
                action_data: 'continue_on_fail=true',
                detail_order: 25
            })
        }
        
        const payload = {
            dialplan_name: form.name,
            dialplan_context: 'default',
            enabled: form.enabled !== false,
            details: details
        }
        
        if (editingOutbound.value && form.id) {
            await systemAPI.updateDialplan(form.id, payload)
        } else {
            await systemAPI.createDialplan(payload)
        }
        
        await loadRoutes()
        showOutboundModal.value = false
        editingOutbound.value = false
        outboundForm.value = { name: '', pattern: '', strip: 0, prepend: '', gateway: '', continue: true, enabled: true }
    } catch (e) {
        console.error(e)
        alert('Failed to save outbound route: ' + (e.message || 'Unknown error'))
    }
}

// Initial Load
onMounted(() => {
  loadAllNumbers()
  loadRoutes()
  loadGateways()
  loadMessagingProviders()
})

// Settings
const settings = ref({
  region: 'nanp',
  format: 'e164'
})

const loadSettings = async () => {
  try {
    const response = await systemAPI.getSettings()
    const data = response.data || {}
    if (data.region) settings.value.region = data.region
    if (data.format) settings.value.format = data.format
  } catch (e) {
    // Settings may not exist yet
    console.debug('No system settings found, using defaults')
  }
}

const saveSettings = async () => {
  try {
    await systemAPI.updateSettings(settings.value)
  } catch (e) {
    console.error(e)
    alert('Failed to save settings: ' + (e.message || 'Unknown error'))
  }
}
</script>

<style scoped>
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-actions { display: flex; gap: 8px; }

/* Tabs */
.tabs { display: flex; gap: 2px; border-bottom: 1px solid var(--border-color); }
.tab { padding: 8px 16px; background: transparent; border: 1px solid transparent; border-bottom: none; cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-muted); border-radius: 4px 4px 0 0; }
.tab.active { background: white; border-color: var(--border-color); color: var(--primary-color); margin-bottom: -1px; }
.tab-content { background: white; border: 1px solid var(--border-color); border-top: none; padding: 20px; border-radius: 0 0 var(--radius-md) var(--radius-md); }

/* Route Help */
.route-help { display: flex; align-items: center; gap: 8px; padding: 12px; background: #eff6ff; border-radius: var(--radius-sm); margin-bottom: 16px; color: #1e40af; font-size: 13px; }
.help-icon { width: 16px; height: 16px; }

/* Filter Bar */
.filter-bar { display: flex; gap: 12px; margin-bottom: 16px; align-items: center; }
.search-box { display: flex; align-items: center; gap: 8px; flex: 1; background: var(--bg-app); border: 1px solid var(--border-color); border-radius: var(--radius-sm); padding: 8px 12px; }
.search-icon { width: 16px; height: 16px; color: var(--text-muted); }
.search-input { border: none; background: transparent; flex: 1; font-size: 14px; outline: none; }
.filter-dropdown { padding: 8px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; min-width: 180px; background: white; cursor: pointer; }

/* Routes List */
.routes-list { display: flex; flex-direction: column; gap: 12px; }
.route-card { display: flex; gap: 16px; background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; align-items: flex-start; }
.route-card.disabled { opacity: 0.5; background: var(--bg-app); }

.route-handle { display: flex; flex-direction: column; align-items: center; gap: 4px; color: var(--text-muted); cursor: grab; }
.route-handle:active { cursor: grabbing; }
.grip-icon { width: 16px; height: 16px; }

/* Drag and Drop States */
.route-card.dragging { opacity: 0.5; background: #e0e7ff; border-style: dashed; }
.route-card.dragover { border-color: var(--primary-color); border-width: 2px; background: #eff6ff; }
.route-order { font-size: 11px; font-weight: 700; background: var(--bg-app); padding: 2px 6px; border-radius: 3px; }

.route-main { flex: 1; }
.route-name-row { display: flex; align-items: center; gap: 12px; margin-bottom: 8px; }
.route-name-row h4 { font-size: 14px; font-weight: 600; margin: 0; }

.route-badges { display: flex; gap: 6px; }
.badge { font-size: 10px; padding: 2px 6px; border-radius: 3px; font-weight: 600; display: flex; align-items: center; gap: 4px; }
.badge.context { background: #f3e8ff; color: #7c3aed; }
.badge.gateway { background: #dbeafe; color: #1d4ed8; }
.badge.intl { background: #fee2e2; color: #dc2626; }

.route-conditions { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 8px; }
.condition { display: flex; align-items: center; gap: 4px; font-size: 12px; background: var(--bg-app); padding: 4px 8px; border-radius: 4px; }
.cond-var { color: #7c3aed; font-weight: 500; }
.cond-op { color: var(--text-muted); }
.cond-val { color: #059669; }

.route-actions-display { display: flex; align-items: center; gap: 8px; }
.arrow-icon { width: 16px; height: 16px; color: var(--text-muted); }
.action-chain { display: flex; flex-wrap: wrap; gap: 6px; }
.action-item { display: flex; gap: 4px; font-size: 12px; background: #ecfdf5; padding: 4px 8px; border-radius: 4px; }
.action-app { color: #059669; font-weight: 600; }
.action-data { color: var(--text-main); font-family: monospace; font-size: 11px; }

.route-pattern { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.pattern-label { font-size: 11px; color: var(--text-muted); }
.pattern-regex { background: #1e293b; color: #22d3ee; padding: 4px 8px; border-radius: 4px; font-size: 12px; }
.pattern-desc { font-size: 12px; color: var(--text-muted); }

.route-transforms { display: flex; gap: 12px; }
.transform { font-size: 11px; color: var(--text-muted); }

.route-controls { display: flex; align-items: center; gap: 8px; }

/* Settings Panel */
.settings-panel { max-width: 800px; }
.settings-section { margin-bottom: 24px; }
.settings-section h3 { font-size: 14px; font-weight: 600; margin-bottom: 12px; }
.settings-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

/* Buttons & Inputs (Shared with Routing.vue) */
.btn-primary { background-color: var(--primary-color); color: white; border: none; padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; font-size: var(--text-sm); cursor: pointer; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 8px 16px; border-radius: var(--radius-sm); font-size: var(--text-sm); font-weight: 500; color: var(--text-main); cursor: pointer; }
.btn-small { font-size: 11px; padding: 4px 8px; border: 1px solid var(--border-color); background: white; border-radius: 4px; cursor: pointer; }
.btn-icon { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 4px; }
.btn-icon:hover { color: var(--text-primary); }
.icon-sm { width: 16px; height: 16px; }
.text-bad { color: var(--status-bad); }

.form-group { display: flex; flex-direction: column; gap: 6px; margin-bottom: 12px; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.flex-2 { flex: 2; }
label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }
.input-field.code { font-family: monospace; background: #f8fafc; }
.input-field.small { width: 120px; }
.input-field:focus { outline: none; border-color: var(--primary-color); }
.help-text { font-size: 11px; color: var(--text-muted); }
.divider { height: 1px; background: var(--border-color); margin: 16px 0; }

.checkbox-group { display: flex; flex-direction: column; gap: 8px; }
.checkbox-row { display: flex; align-items: center; gap: 8px; font-size: 13px; cursor: pointer; }

.form-section { margin-bottom: 16px; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.section-header h4 { font-size: 13px; font-weight: 600; margin: 0; }

.conditions-editor, .actions-editor { display: flex; flex-direction: column; gap: 8px; }
.condition-row, .action-row { display: flex; gap: 8px; align-items: center; }
.condition-row .input-field, .action-row .input-field { flex: 1; }

.form-actions { display: flex; justify-content: flex-end; gap: 12px; margin-top: 20px; }

/* Stats Row */
.stats-row {
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
}
.stat-card {
  flex: 1;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
  text-align: center;
}
.stat-card.accent {
  border-color: var(--primary-color);
  background: linear-gradient(135deg, #f8faff 0%, #eef4ff 100%);
}
.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
}
.stat-label {
  font-size: 11px;
  color: var(--text-muted);
  text-transform: uppercase;
  margin-top: 4px;
}
.btn-secondary {
  display: flex;
  align-items: center;
  gap: 6px;
  background: white;
  border: 1px solid var(--border-color);
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}
.btn-secondary:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}
.btn-icon { width: 14px; height: 14px; }

@media (max-width: 768px) {
  .stats-row { flex-wrap: wrap; }
  .stat-card { min-width: 140px; }
  .view-header { flex-direction: column; gap: 12px; align-items: flex-start; }
  .header-actions { width: 100%; flex-wrap: wrap; }
}

/* Sub-tabs */
.sub-tabs { display: flex; gap: 4px; margin-bottom: 16px; }
.sub-tab { padding: 6px 14px; background: var(--bg-app); border: 1px solid var(--border-color); border-radius: 6px; font-size: 12px; font-weight: 600; cursor: pointer; color: var(--text-muted); }
.sub-tab.active { background: var(--primary-color); color: white; border-color: var(--primary-color); }

/* Capability badges */
.cap-badges { display: flex; flex-wrap: wrap; gap: 4px; }
.cap-badge { font-size: 10px; padding: 2px 6px; border-radius: 3px; font-weight: 600; }
.cap-badge.voice { background: #dbeafe; color: #1d4ed8; }
.cap-badge.sms { background: #d1fae5; color: #059669; }
.cap-badge.mms { background: #fce7f3; color: #db2777; }
.cap-badge.fax { background: #e0e7ff; color: #4338ca; }
.cap-badge.e911 { background: #fee2e2; color: #dc2626; }

/* Tenant badge */
.tenant-badge { font-size: 12px; padding: 2px 8px; background: #f3e8ff; color: #7c3aed; border-radius: 4px; font-weight: 500; }

/* Gateway priority list */
.gw-priority-list { display: flex; flex-wrap: wrap; gap: 6px; margin-top: 8px; }
.gw-priority-item { font-size: 11px; background: var(--bg-app); padding: 4px 8px; border-radius: 4px; color: var(--text-main); }

/* Gateway editor in modals */
.gw-editor { display: flex; flex-direction: column; gap: 8px; }
.gw-row { display: flex; gap: 8px; align-items: center; }
.gw-order { font-size: 12px; font-weight: 700; color: var(--text-muted); min-width: 20px; text-align: center; }

/* Empty state */
.empty-state { text-align: center; padding: 40px 20px; color: var(--text-muted); }
.empty-state p { font-size: 14px; }

/* Modals */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-card { background: white; border-radius: var(--radius-md); width: 600px; max-height: 90vh; overflow-y: auto; box-shadow: 0 20px 60px rgba(0,0,0,0.15); }
.modal-card.small { width: 420px; }
.modal-card.large { width: 780px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 20px 24px 16px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { margin: 0; font-size: 16px; }
.modal-body { padding: 20px 24px; }
.modal-actions { display: flex; justify-content: flex-end; gap: 8px; padding: 16px 24px; border-top: 1px solid var(--border-color); }

.text-sm { font-size: 13px; }
</style>
