<template>
  <div class="data-table-wrapper">
    <div class="data-table-container" :class="{ 'has-actions': actions }">
      <table class="data-table">
        <thead>
          <tr>
            <th 
              v-for="col in columns" 
              :key="col.key" 
              :style="{ width: col.width }"
              :class="{ 
                'text-left': col.align === 'left' || !col.align,
                'text-center': col.align === 'center',
                'text-right': col.align === 'right',
                'sortable': col.sortable,
                'sorted': sortKey === col.key
              }"
              @click="col.sortable && handleSort(col.key)"
            >
              <div class="th-content">
                {{ col.label }}
                <span v-if="col.sortable" class="sort-indicator">
                  <ChevronUpIcon v-if="sortKey === col.key && sortOrder === 'asc'" class="sort-icon" />
                  <ChevronDownIcon v-else-if="sortKey === col.key && sortOrder === 'desc'" class="sort-icon" />
                  <span v-else class="sort-placeholder">⇅</span>
                </span>
              </div>
            </th>
            <th v-if="actions" class="actions-header">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, index) in sortedData" :key="row.id || index" class="data-row">
            <td 
              v-for="col in columns" 
              :key="col.key"
              :class="{ 
                'text-left': col.align === 'left' || !col.align,
                'text-center': col.align === 'center',
                'text-right': col.align === 'right'
              }"
              :data-label="col.label"
            >
              <slot :name="col.key" :row="row" :value="row[col.key]">
                <span class="cell-content">{{ formatValue(row[col.key], col.format) }}</span>
              </slot>
            </td>
            <td v-if="actions" class="actions-cell">
              <div class="actions-wrapper">
                <slot name="actions" :row="row"></slot>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      
      <!-- Empty State -->
      <div v-if="sortedData.length === 0" class="empty-state">
        <div class="empty-content">
          <InboxIcon class="empty-icon" />
          <p class="empty-text">No data available</p>
        </div>
      </div>
    </div>
    
    <!-- Pagination -->
    <div v-if="pagination && totalPages > 1" class="pagination">
      <button 
        class="pagination-btn" 
        :disabled="currentPage === 1"
        @click="currentPage--"
        aria-label="Previous page"
      >
        <ChevronLeftIcon class="pagination-icon" />
      </button>
      
      <div class="pagination-info">
        <span class="page-numbers">{{ currentPage }} / {{ totalPages }}</span>
        <span class="page-total">({{ data.length }} total)</span>
      </div>
      
      <button 
        class="pagination-btn" 
        :disabled="currentPage === totalPages"
        @click="currentPage++"
        aria-label="Next page"
      >
        <ChevronRightIcon class="pagination-icon" />
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { 
  ChevronUp as ChevronUpIcon, 
  ChevronDown as ChevronDownIcon,
  ChevronLeft as ChevronLeftIcon,
  ChevronRight as ChevronRightIcon,
  Inbox as InboxIcon
} from 'lucide-vue-next'

const props = defineProps({
  columns: {
    type: Array,
    required: true
    // [{ key, label, width, align, sortable, format }]
  },
  data: {
    type: Array,
    required: true
  },
  actions: {
    type: Boolean,
    default: false
  },
  pagination: {
    type: Boolean,
    default: false
  },
  pageSize: {
    type: Number,
    default: 10
  }
})

// Sorting state
const sortKey = ref('')
const sortOrder = ref('asc')
const currentPage = ref(1)

// Handle sorting
const handleSort = (key) => {
  if (sortKey.value === key) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortOrder.value = 'asc'
  }
}

// Format cell value
const formatValue = (value, format) => {
  if (value === null || value === undefined) return '—'
  if (format === 'date') {
    return new Date(value).toLocaleDateString()
  }
  if (format === 'datetime') {
    return new Date(value).toLocaleString()
  }
  if (format === 'number') {
    return Number(value).toLocaleString()
  }
  if (format === 'currency') {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(value)
  }
  return value
}

// Sorted and paginated data
const sortedData = computed(() => {
  let result = [...props.data]
  
  // Sort
  if (sortKey.value) {
    result.sort((a, b) => {
      const aVal = a[sortKey.value]
      const bVal = b[sortKey.value]
      
      if (aVal === null || aVal === undefined) return 1
      if (bVal === null || bVal === undefined) return -1
      
      if (typeof aVal === 'string') {
        const comparison = aVal.localeCompare(bVal)
        return sortOrder.value === 'asc' ? comparison : -comparison
      }
      
      if (aVal < bVal) return sortOrder.value === 'asc' ? -1 : 1
      if (aVal > bVal) return sortOrder.value === 'asc' ? 1 : -1
      return 0
    })
  }
  
  // Paginate
  if (props.pagination) {
    const start = (currentPage.value - 1) * props.pageSize
    const end = start + props.pageSize
    return result.slice(start, end)
  }
  
  return result
})

const totalPages = computed(() => Math.ceil(props.data.length / props.pageSize))
</script>

<style scoped>
.data-table-wrapper {
  width: 100%;
}

.data-table-container {
  background: var(--bg-card);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
  overflow: hidden;
}

.data-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
}

/* Header Styles */
th {
  background: var(--bg-hover);
  color: var(--text-secondary);
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  text-transform: uppercase;
  letter-spacing: var(--tracking-wider);
  padding: var(--spacing-3) var(--spacing-4);
  text-align: left;
  border-bottom: 1px solid var(--border-color);
  white-space: nowrap;
  user-select: none;
}

.th-content {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
}

th.sortable {
  cursor: pointer;
  transition: color var(--transition-fast);
}

th.sortable:hover {
  color: var(--text-primary);
}

th.sorted {
  color: var(--primary-color);
}

.sort-indicator {
  display: inline-flex;
  align-items: center;
  color: var(--text-muted);
}

.sort-icon {
  width: 14px;
  height: 14px;
}

.sort-placeholder {
  font-size: 10px;
  opacity: 0.5;
}

/* Cell Styles */
td {
  padding: var(--spacing-3) var(--spacing-4);
  font-size: var(--text-sm);
  color: var(--text-main);
  border-bottom: 1px solid var(--border-light);
  vertical-align: middle;
  background: var(--bg-card);
}

.cell-content {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.data-row {
  transition: background var(--transition-fast);
}

.data-row:hover td {
  background: var(--bg-hover);
}

.data-row:last-child td {
  border-bottom: none;
}

/* Alignment */
.text-left { text-align: left; }
.text-center { text-align: center; }
.text-right { text-align: right; }

/* Actions */
.actions-header {
  width: 100px;
  text-align: right;
}

.actions-cell {
  text-align: right;
}

.actions-wrapper {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: var(--spacing-1);
}

/* Empty State */
.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-12) var(--spacing-6);
  background: var(--bg-card);
}

.empty-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-3);
  color: var(--text-muted);
}

.empty-icon {
  width: 48px;
  height: 48px;
  opacity: 0.5;
}

.empty-text {
  font-size: var(--text-sm);
  color: var(--text-secondary);
}

/* Pagination */
.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-4);
  padding: var(--spacing-4);
  background: var(--bg-card);
  border-top: 1px solid var(--border-light);
}

.pagination-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.pagination-btn:hover:not(:disabled) {
  background: var(--bg-hover);
  border-color: var(--border-hover);
  color: var(--text-primary);
}

.pagination-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.pagination-icon {
  width: 16px;
  height: 16px;
}

.pagination-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  font-size: var(--text-sm);
}

.page-numbers {
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.page-total {
  font-size: var(--text-xs);
  color: var(--text-muted);
}

/* Mobile Responsive Styles */
@media (max-width: 768px) {
  .data-table-container {
    border-radius: var(--radius-md);
  }
  
  th {
    padding: var(--spacing-2) var(--spacing-3);
    font-size: var(--text-2xs);
  }
  
  td {
    padding: var(--spacing-3);
    font-size: var(--text-sm);
  }
  
  /* Stackable table on mobile */
  .data-table,
  .data-table tbody,
  .data-table tr,
  .data-table td {
    display: block;
    width: 100%;
  }
  
  .data-table thead {
    display: none;
  }
  
  .data-row {
    margin-bottom: var(--spacing-3);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    overflow: hidden;
  }
  
  .data-row:last-child {
    margin-bottom: 0;
  }
  
  .data-row td {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--spacing-2) var(--spacing-3);
    border-bottom: 1px solid var(--border-light);
  }
  
  .data-row td:last-child {
    border-bottom: none;
  }
  
  .data-row td::before {
    content: attr(data-label);
    font-weight: var(--font-semibold);
    color: var(--text-secondary);
    font-size: var(--text-xs);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wider);
  }
  
  .data-row td.actions-cell::before {
    display: none;
  }
  
  .cell-content {
    white-space: normal;
    text-align: right;
  }
  
  .actions-cell {
    justify-content: flex-end;
    background: var(--bg-hover);
  }
  
  .actions-wrapper {
    width: 100%;
    justify-content: flex-end;
  }
  
  /* Pagination mobile */
  .pagination {
    padding: var(--spacing-3);
  }
}

@media (max-width: 480px) {
  .empty-state {
    padding: var(--spacing-8) var(--spacing-4);
  }
  
  .empty-icon {
    width: 40px;
    height: 40px;
  }
}
</style>
