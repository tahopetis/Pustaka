# CI Sidebar Refactor Implementation Plan

## Overview

Refactor the Configuration Items sidebar menu to have a collapsible hierarchy structure with domain-aware CI Type autocomplete selection.

**Target Structure:**
```
Configuration Items (collapsible)
├─ Asset Management CI
└─ Enterprise Architecture CI (collapsible)
   ├─ Strategy
   ├─ Business
   ├─ Application
   ├─ Data
   ├─ Infrastructure
   ├─ Security
   ├─ Technology
   └─ Governance
```

## Worktree Location

```bash
/home/tahopetis/dev/pustaka-ci-sidebar-refactor
```

## Current State Analysis

### Existing Patterns

1. **EA Domain Detection** (`web/src/stores/eaTypes.ts`):
   - EA CI types follow naming: `EA.{Domain}.*` (e.g., `EA.Application.*`)
   - Asset Management CIs are everything NOT starting with `EA.`
   - Store method: `getCiTypesByDomain(domain)` filters by prefix

2. **CI Type Dropdown** (`web/src/views/ci/CIFormView.vue` lines 34-45):
   - Currently uses HTML `<select>` element
   - Shows ALL CI types regardless of context
   - No search/autocomplete functionality

3. **Autocomplete Reference** (`web/src/views/graph/GraphView.vue` lines 24-72):
   - Perfect implementation to reference
   - Debounced search (300ms)
   - Max 5 results
   - Dropdown with CI type badges

## Implementation Tasks

### Task 1: Create CITypeAutocomplete Component

**File**: `web/src/components/base/CITypeAutocomplete.vue`

**Requirements**:
- Autocomplete input with debounced search (300ms delay)
- Domain-aware filtering based on prop
- Max 5 search results displayed
- Show CI type name and description
- Props:
  - `modelValue` (string): Selected CI type name
  - `domain` ('asset' | 'ea' | 'all'): Filter domain
  - `eaDomain` (string, optional): Specific EA domain if domain='ea'
  - `disabled` (boolean): Disable input
  - `placeholder` (string): Input placeholder text
- Emits:
  - `update:modelValue`: When selection changes
  - `selected`: When user selects an option

**Implementation Reference**:
- Copy autocomplete pattern from `GraphView.vue` (lines 24-72, 511-600)
- Use `ciTypeAPI.list()` to fetch CI types
- Filter results based on domain prop:
  - `domain='asset'`: Filter out types starting with `EA.`
  - `domain='ea'`: Only show types starting with `EA.{eaDomain}`
  - `domain='all'`: Show all types

**Template Structure**:
```vue
<template>
  <div class="relative">
    <input
      :value="displayValue"
      @input="onInput"
      @focus="showDropdown = true"
      @blur="hideDropdown"
      type="text"
      :placeholder="placeholder"
      :disabled="disabled"
      class="form-input"
    />
    <div v-if="showDropdown && (searchResults.length > 0 || searching)" class="autocomplete-dropdown">
      <!-- Results list -->
    </div>
  </div>
</template>
```

---

### Task 2: Refactor AppSidebar Component

**File**: `web/src/components/layout/AppSidebar.vue`

**Changes**:

1. **Add Collapsible State**:
```vue
<script setup>
const ciMenuOpen = ref(false)
const eaMenuOpen = ref(false)
</script>
```

2. **Replace "Configuration Items" Link** (lines 30-50) with:
```vue
<!-- Configuration Items (Collapsible) -->
<div>
  <button
    @click="ciMenuOpen = !ciMenuOpen"
    class="w-full group flex items-center px-2 py-2 text-sm font-medium rounded-md"
    :class="isCurrentRoute('/ci') || hasCIRoute()
      ? 'bg-blue-100 text-blue-900'
      : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'"
  >
    <svg class="mr-3 h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
    </svg>
    <span class="flex-1 text-left">Configuration Items</span>
    <svg :class="ciMenuOpen ? 'rotate-90' : ''" class="w-4 h-4 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
    </svg>
  </button>

  <!-- Sub-menu (shown when ciMenuOpen = true) -->
  <div v-if="ciMenuOpen" class="ml-6 mt-1 space-y-1">

    <!-- Asset Management CI -->
    <router-link
      :to="{ name: 'CreateCI', query: { context: 'asset' } }"
      class="group flex items-center px-2 py-2 text-sm font-medium rounded-md"
      :class="isCIContext('asset')
        ? 'bg-blue-100 text-blue-900'
        : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'"
    >
      <svg class="mr-3 h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
      </svg>
      Asset Management CI
    </router-link>

    <!-- Enterprise Architecture CI (Collapsible) -->
    <div>
      <button
        @click="eaMenuOpen = !eaMenuOpen"
        class="w-full group flex items-center px-2 py-2 text-sm font-medium rounded-md"
        :class="isEAContext()
          ? 'bg-blue-100 text-blue-900'
          : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'"
      >
        <svg class="mr-3 h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
        </svg>
        <span class="flex-1 text-left">Enterprise Architecture CI</span>
        <svg :class="eaMenuOpen ? 'rotate-90' : ''" class="w-4 h-4 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </button>

      <!-- EA Domains Sub-menu -->
      <div v-if="eaMenuOpen" class="ml-6 mt-1 space-y-1">
        <router-link
          v-for="domain in eaDomains"
          :key="domain.value"
          :to="{ name: 'CreateCI', query: { context: 'ea', domain: domain.value } }"
          class="group flex items-center px-2 py-2 text-sm font-medium rounded-md"
          :class="isEADomain(domain.value)
            ? 'bg-blue-100 text-blue-900'
            : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'"
        >
          <svg class="mr-3 h-3 w-3" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
          {{ domain.label }}
        </router-link>
      </div>
    </div>

  </div>
</div>
```

3. **Add Computed Properties & Helpers**:
```vue
<script setup>
const eaDomains = [
  { value: 'strategy', label: 'Strategy' },
  { value: 'business', label: 'Business' },
  { value: 'application', label: 'Application' },
  { value: 'data', label: 'Data' },
  { value: 'infrastructure', label: 'Infrastructure' },
  { value: 'security', label: 'Security' },
  { value: 'technology', label: 'Technology' },
  { value: 'governance', label: 'Governance' }
]

const isCIContext = (context: string) => {
  return route.query.context === context
}

const isEAContext = () => {
  return route.query.context === 'ea'
}

const isEADomain = (domain: string) => {
  return route.query.context === 'ea' && route.query.domain === domain
}

const hasCIRoute = () => {
  return route.path.startsWith('/ci') || route.path.startsWith('/entities')
}
</script>
```

---

### Task 3: Update CIFormView

**File**: `web/src/views/ci/CIFormView.vue`

**Changes**:

1. **Read Query Parameters** (add to script setup):
```vue
<script setup>
const context = computed(() => route.query.context as string | undefined)
const domain = computed(() => route.query.domain as string | undefined)
</script>
```

2. **Update Page Title** (lines 5-10):
```vue
<h1 class="page-title">
  {{
    isEdit
      ? 'Edit Configuration Item'
      : context.value === 'ea' && domain.value
        ? `Create ${domain.value.charAt(0).toUpperCase() + domain.value.slice(1)} Entity`
        : context.value === 'asset'
          ? 'Create Asset Management CI'
          : 'Create Configuration Item'
  }}
</h1>
```

3. **Replace CI Type Dropdown** (lines 33-46) with:
```vue
<!-- Replace this section -->
<CITypeAutocomplete
  v-model="form.ci_type"
  :domain="context.value === 'ea' ? 'ea' : context.value === 'asset' ? 'asset' : 'all'"
  :ea-domain="domain.value"
  placeholder="Search CI types..."
  :disabled="loading || isEdit"
  @change="onCITypeChange"
/>
```

4. **Filter CI Types** (update `loadCITypes` function around line 594):
```javascript
const loadCITypes = async () => {
  try {
    const response = await ciTypeAPI.list()
    let allTypes = response.data.ci_types || []

    // Filter based on context
    if (context.value === 'asset') {
      // Show only non-EA types
      allTypes = allTypes.filter((type: CIType) => !type.name.startsWith('EA.'))
    } else if (context.value === 'ea' && domain.value) {
      // Show only EA types for specific domain
      const domainPrefix = `EA.${domain.value.charAt(0).toUpperCase() + domain.value.slice(1)}`
      allTypes = allTypes.filter((type: CIType) => type.name.startsWith(domainPrefix))
    }

    ciTypes.value = allTypes
  } catch (error) {
    console.error('Failed to load CI types:', error)
    showErrorToast('Failed to load CI types')
  }
}
```

5. **Import Component** (add to imports):
```vue
<script setup>
import CITypeAutocomplete from '@/components/base/CITypeAutocomplete.vue'
</script>
```

---

### Task 4: Add Router Convenience Routes (Optional)

**File**: `web/src/router/index.ts`

**Add after line 46** (after EditCI route):
```javascript
// Asset Management CI create
{
  path: '/ci/asset/new',
  name: 'CreateAssetCI',
  component: () => import('@/views/ci/CIFormView.vue'),
  meta: { requiresAuth: true, requiresPermission: 'ci:create' }
},
// EA Entity create by domain
{
  path: '/ci/ea/:domain/new',
  name: 'CreateEACIDomain',
  component: () => import('@/views/ci/CIFormView.vue'),
  meta: { requiresAuth: true, requiresPermission: 'ci:create' }
},
```

Then update sidebar links to use these routes instead:
```vue
<!-- Asset Management -->
<router-link
  :to="{ name: 'CreateAssetCI' }"
  ...
>

<!-- EA Domains -->
<router-link
  :to="{ name: 'CreateEACIDomain', params: { domain: domain.value } }"
  ...
>
```

---

## Implementation Order

Execute in this order for minimal conflicts:

1. **Task 1**: Create `CITypeAutocomplete.vue` (independent)
2. **Task 2**: Refactor `AppSidebar.vue` (independent)
3. **Task 3**: Update `CIFormView.vue` (depends on Task 1)
4. **Task 4**: Update router (optional, can be done anytime)

---

## Testing Checklist

After implementation, verify:

### Sidebar
- [ ] "Configuration Items" is collapsible
- [ ] "Asset Management CI" link navigates to `/ci/new?context=asset`
- [ ] "Enterprise Architecture CI" is collapsible
- [ ] All 8 EA domains appear as sub-items under EA
- [ ] Each EA domain link navigates to `/ci/new?context=ea&domain=X`
- [ ] Active route highlighting works correctly
- [ ] Collapsible state persists during navigation

### CI Form - Asset Context
- [ ] Visiting `/ci/new?context=asset` shows "Create Asset Management CI" title
- [ ] CI Type autocomplete only shows non-EA types
- [ ] Search in autocomplete filters only non-EA types
- [ ] Max 5 results shown in autocomplete
- [ ] Creating CI works correctly

### CI Form - EA Context
- [ ] Visiting `/ci/new?context=ea&domain=strategy` shows "Create Strategy Entity" title
- [ ] CI Type autocomplete only shows `EA.Strategy*` types
- [ ] Search in autocomplete filters only `EA.Strategy*` types
- [ ] Each EA domain shows correct CI types
- [ ] Creating entity works correctly

### CI Form - Default (No Context)
- [ ] Visiting `/ci/new` shows "Create Configuration Item" title
- [ ] CI Type autocomplete shows all CI types
- [ ] No filtering applied
- [ ] Creating CI works correctly

---

## Rollback Plan

If issues occur:
```bash
# Switch back to main branch
cd /home/tahopetis/dev/pustaka-ci-sidebar-refactor
git checkout main

# Or view changes
git diff main

# To re-apply work
git checkout feature/ci-sidebar-refactor
```

---

## Files to Modify

1. `web/src/components/base/CITypeAutocomplete.vue` - **NEW FILE**
2. `web/src/components/layout/AppSidebar.vue` - **MODIFY**
3. `web/src/views/ci/CIFormView.vue` - **MODIFY**
4. `web/src/router/index.ts` - **OPTIONAL MODIFY**

---

## Technical Notes

### Domain Detection Logic
```javascript
// Asset Management: CI types NOT starting with "EA."
const isAsset = (ciTypeName) => !ciTypeName.startsWith('EA.')

// EA Domain: CI types starting with "EA.{Domain}"
const isEADomain = (ciTypeName, domain) => {
  const prefix = `EA.${domain.charAt(0).toUpperCase() + domain.slice(1)}`
  return ciTypeName.startsWith(prefix)
}
```

### EA Domain Mapping
| URL Domain Value | CI Type Prefix | Example |
|-----------------|----------------|---------|
| strategy | EA.Strategy | EA.Strategy.Capability |
| business | EA.Business | EA.Business.Process |
| application | EA.Application | EA.Application.Service |
| data | EA.Data | EA.Data.Entity |
| infrastructure | EA.Infrastructure | EA.Infrastructure.Server |
| security | EA.Security | EA.Security.Control |
| technology | EA.Technology | EA.Technology.Product |
| governance | EA.Governance | EA.Governance.Policy |

---

## Estimated Effort

- Task 1 (CITypeAutocomplete): 1-2 hours
- Task 2 (AppSidebar): 1 hour
- Task 3 (CIFormView): 1 hour
- Task 4 (Router): 30 minutes (optional)
- Testing & Refinement: 1 hour

**Total**: ~4-5 hours
