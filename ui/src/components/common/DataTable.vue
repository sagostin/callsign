<template>
  <div class="data-table-container">
    <table class="data-table">
      <thead>
        <tr>
          <th v-for="col in columns" :key="col.key" :style="{ width: col.width }">
            {{ col.label }}
          </th>
          <th v-if="actions" class="actions-header">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(row, index) in data" :key="index">
          <td v-for="col in columns" :key="col.key">
            <slot :name="col.key" :row="row" :value="row[col.key]">
              {{ row[col.key] }}
            </slot>
          </td>
          <td v-if="actions" class="actions-cell">
            <slot name="actions" :row="row"></slot>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
defineProps({
  columns: {
    type: Array, // [{ key, label, width }]
    required: true
  },
  data: {
    type: Array,
    required: true
  },
  actions: {
    type: Boolean,
    default: false
  }
})
</script>

<style scoped>
.data-table-container {
  background: white;
  border-radius: var(--radius-md);
  overflow-x: auto; /* Enable horizontal scroll */
  box-shadow: var(--shadow-sm);
  border: 1px solid transparent; /* No border, just shadow */
}

.data-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
}

th {
  background-color: white; /* No gray header background */
  color: var(--text-muted);
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 16px var(--spacing-lg);
  text-align: left;
  border-bottom: 2px solid var(--bg-app);
}

td {
  padding: 12px var(--spacing-lg);
  border-bottom: 1px solid var(--bg-app); /* Very subtle divider */
  font-size: var(--text-sm);
  color: var(--text-main);
  vertical-align: middle;
  background: white;
}

tr:last-child td {
  border-bottom: none;
}

tr:hover td {
  background-color: var(--bg-app); /* Subtle hover */
}

.actions-header {
  width: 100px;
  text-align: right;
}

.actions-cell {
  text-align: right;
}
</style>
