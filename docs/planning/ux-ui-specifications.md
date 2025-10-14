# Pustaka CMDB - UX/UI Specifications

**Version**: 1.0
**Date**: October 14, 2025
**Status**: Final
**Design System**: Tailwind CSS + Headless UI

## Design Philosophy

### Core Principles
1. **Data-First Design**: Information architecture prioritizes data clarity and accessibility
2. **Relationship-Centric**: Visual design emphasizes connections and dependencies
3. **Efficiency-Focused**: Interactions designed for speed and productivity
4. **Progressive Disclosure**: Complex features revealed gradually based on user needs
5. **Accessibility by Default**: WCAG 2.1 AA compliance built into every component

### Visual Design Language
- **Modern & Clean**: Minimalist approach with purposeful use of color and whitespace
- **Professional**: Suitable for enterprise environments while maintaining usability
- **Responsive**: Seamless experience across desktop, tablet, and mobile devices
- **Consistent**: Unified design language across all application interfaces

## Design System Specifications

### Color Palette

#### Primary Colors
```css
/* Brand Colors */
--primary-50: #eff6ff;    /* Lightest */
--primary-100: #dbeafe;
--primary-200: #bfdbfe;
--primary-300: #93c5fd;
--primary-400: #60a5fa;
--primary-500: #3b82f6;   /* Primary */
--primary-600: #2563eb;
--primary-700: #1d4ed8;
--primary-800: #1e40af;
--primary-900: #1e3a8a;   /* Darkest */

/* Secondary Colors */
--secondary-50: #f8fafc;
--secondary-100: #f1f5f9;
--secondary-200: #e2e8f0;
--secondary-300: #cbd5e1;
--secondary-400: #94a3b8;
--secondary-500: #64748b;  /* Secondary */
--secondary-600: #475569;
--secondary-700: #334155;
--secondary-800: #1e293b;
--secondary-900: #0f172a;
```

#### Semantic Colors
```css
/* Success */
--success-50: #f0fdf4;
--success-500: #22c55e;
--success-600: #16a34a;

/* Warning */
--warning-50: #fffbeb;
--warning-500: #f59e0b;
--warning-600: #d97706;

/* Error */
--error-50: #fef2f2;
--error-500: #ef4444;
--error-600: #dc2626;

/* Info */
--info-50: #eff6ff;
--info-500: #3b82f6;
--info-600: #2563eb;
```

#### Relationship Colors (for Graph Visualization)
```css
/* Relationship Type Colors */
--relationship-default: #94a3b8;
--relationship-depends-on: #3b82f6;
--relationship-hosts: #22c55e;
--relationship-connects-to: #f59e0b;
--relationship-manages: #8b5cf6;
--relationship-backups: #06b6d4;
--relationship-monitors: #ef4444;
```

### Typography

#### Font Stack
```css
/* Primary Font */
font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;

/* Monospace Font (for code, technical data) */
font-family: 'JetBrains Mono', 'Fira Code', Consolas, monospace;
```

#### Type Scale
```css
/* Headings */
--text-4xl: 2.25rem;    /* 36px - Page Titles */
--text-3xl: 1.875rem;   /* 30px - Section Headers */
--text-2xl: 1.5rem;     /* 24px - Card Titles */
--text-xl: 1.25rem;     /* 20px - Sub-sections */
--text-lg: 1.125rem;    /* 18px - Large Body */

/* Body Text */
--text-base: 1rem;      /* 16px - Default Body */
--text-sm: 0.875rem;    /* 14px - Small Body */
--text-xs: 0.75rem;     /* 12px - Labels, Captions */
```

### Spacing System

```css
/* Spacing Scale (8px base unit) */
--space-1: 0.25rem;   /* 4px */
--space-2: 0.5rem;    /* 8px */
--space-3: 0.75rem;   /* 12px */
--space-4: 1rem;      /* 16px */
--space-5: 1.25rem;   /* 20px */
--space-6: 1.5rem;    /* 24px */
--space-8: 2rem;      /* 32px */
--space-10: 2.5rem;   /* 40px */
--space-12: 3rem;     /* 48px */
--space-16: 4rem;     /* 64px */
```

### Component Library

#### Button Specifications

##### Primary Button
```css
.btn-primary {
  @apply bg-primary-600 text-white px-4 py-2 rounded-lg font-medium
         hover:bg-primary-700 focus:ring-2 focus:ring-primary-500
         focus:ring-offset-2 disabled:bg-primary-300 disabled:cursor-not-allowed
         transition-colors duration-200;
}
```

**Usage**: Primary actions (Create, Save, Submit)
**States**: Default, Hover, Focus, Disabled, Loading

##### Secondary Button
```css
.btn-secondary {
  @apply bg-white text-primary-600 border border-primary-300 px-4 py-2 rounded-lg font-medium
         hover:bg-primary-50 focus:ring-2 focus:ring-primary-500
         focus:ring-offset-2 disabled:text-gray-300 disabled:border-gray-200
         transition-colors duration-200;
}
```

**Usage**: Secondary actions (Cancel, Reset, Export)

##### Icon Button
```css
.btn-icon {
  @apply p-2 rounded-lg text-gray-600 hover:bg-gray-100
         focus:ring-2 focus:ring-primary-500 focus:ring-offset-2
         transition-colors duration-200;
}
```

**Usage**: Actions in tables, toolbars, headers

#### Form Controls

##### Input Fields
```css
.input-field {
  @apply w-full px-3 py-2 border border-gray-300 rounded-lg
         focus:ring-2 focus:ring-primary-500 focus:border-primary-500
         disabled:bg-gray-50 disabled:text-gray-500;
}
```

**States**: Default, Focus, Error, Disabled, Loading

##### Validation States
```css
.input-error {
  @apply border-error-500 focus:ring-error-500 focus:border-error-500;
}

.input-success {
  @apply border-success-500 focus:ring-success-500 focus:border-success-500;
}
```

##### Select Dropdowns
```css
.select-field {
  @apply w-full px-3 py-2 border border-gray-300 rounded-lg bg-white
         focus:ring-2 focus:ring-primary-500 focus:border-primary-500;
}
```

#### Data Display Components

##### Tables
```css
.data-table {
  @apply w-full bg-white shadow-sm rounded-lg overflow-hidden;
}

.data-table-header {
  @apply bg-gray-50 border-b border-gray-200;
}

.data-table-row {
  @apply border-b border-gray-200 hover:bg-gray-50 transition-colors duration-150;
}
```

**Features**: Sorting, Filtering, Pagination, Row Selection, Bulk Actions

##### Cards
```css
.card {
  @apply bg-white shadow-sm rounded-lg border border-gray-200;
}

.card-header {
  @apply px-6 py-4 border-b border-gray-200;
}

.card-body {
  @apply px-6 py-4;
}
```

##### Status Indicators
```css
.status-active {
  @apply inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800;
}

.status-inactive {
  @apply inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800;
}

.status-pending {
  @apply inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800;
}
```

## Page Layout Specifications

### Global Layout Structure

```
┌─────────────────────────────────────────────────────────┐
│                    Header Bar                            │
├─────────────────────────────────────────────────────────┤
│ Sidebar │                Main Content                   │
│         │                                               │
│ Nav     │                Page Header                    │
│ Menu    │                                               │
│         │                Content Area                   │
│         │                                               │
│         │                                               │
└─────────┴───────────────────────────────────────────────┘
```

### Header Bar (Height: 64px)
- **Logo**: Left-aligned, links to dashboard
- **Global Search**: Center-aligned, full-width search
- **User Menu**: Right-aligned, avatar, notifications, profile
- **Responsive**: Collapses to hamburger menu on mobile

### Sidebar (Width: 256px, Collapsible)
- **Navigation**: Hierarchical menu structure
- **Active State**: Clear visual indication of current page
- **Icons**: Consistent iconography for all menu items
- **Badges**: Notification/count badges where applicable
- **Responsive**: Overlays content on mobile devices

### Main Content Area
- **Max Width**: 1440px with auto margins on large screens
- **Padding**: Consistent spacing using spacing system
- **Responsive**: Adapts to screen size with breakpoints

## Page-Specific UI Specifications

### 1. Dashboard Page

#### Layout Structure
```
┌─────────────────────────────────────────────────────────┐
│                    Dashboard Header                      │
├─────────────────────────────────────────────────────────┤
│  Stats Cards (4-column grid)                            │
├─────────────────────────────────────────────────────────┤
│  Quick Actions (3-column grid)                          │
├─────────────────────────────────────────────────────────┤
│  Recent Activity List                                    │
└─────────────────────────────────────────────────────────┘
```

#### Stats Card Specifications
```css
.stats-card {
  @apply bg-white rounded-lg shadow p-6 border border-gray-200;
}

.stats-icon {
  @apply w-8 h-8 rounded-lg flex items-center justify-center;
}

.stats-value {
  @apply text-2xl font-bold text-gray-900;
}

.stats-label {
  @apply text-sm font-medium text-gray-500;
}
```

#### Quick Actions Section
- **Layout**: 3-column responsive grid
- **Buttons**: Primary and secondary styles
- **Icons**: Consistent 16px icons
- **Hover States**: Subtle elevation changes

### 2. CI List Page

#### Layout Structure
```
┌─────────────────────────────────────────────────────────┐
│  Page Header (Title + Actions)                           │
├─────────────────────────────────────────────────────────┤
│  Filters Bar (Horizontal)                                │
├─────────────────────────────────────────────────────────┤
│  Bulk Actions Bar (when items selected)                  │
├─────────────────────────────────────────────────────────┤
│  Data Table                                              │
├─────────────────────────────────────────────────────────┤
│  Pagination Controls                                     │
└─────────────────────────────────────────────────────────┘
```

#### Filter Bar Specifications
```css
.filter-bar {
  @apply flex flex-wrap gap-4 p-4 bg-gray-50 border-b border-gray-200;
}

.filter-group {
  @apply flex items-center gap-2;
}

.filter-label {
  @apply text-sm font-medium text-gray-700;
}

.filter-input {
  @apply min-w-0 flex-1;
}
```

#### Table Specifications
- **Columns**: Configurable column visibility
- **Sorting**: Click headers to sort (ascending/descending)
- **Selection**: Checkbox selection with bulk actions
- **Row Actions**: Dropdown menu for each row
- **Pagination**: 20, 50, 100 items per page

### 3. CI Details Page

#### Layout Structure
```
┌─────────────────────────────────────────────────────────┐
│  Page Header (Breadcrumb + Actions)                     │
├─────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────────────────────────┐   │
│  │   CI Info   │  │      Relationships Tab         │   │
│  │   Panel     │  │                               │   │
│  │             │  │  Graph Visualization           │   │
│  │             │  │                               │   │
│  └─────────────┘  └─────────────────────────────────┘   │
├─────────────────────────────────────────────────────────┤
│  Tabs (Details | History | Relationships | Audit)        │
├─────────────────────────────────────────────────────────┤
│  Tab Content Area                                       │
└─────────────────────────────────────────────────────────┘
```

#### CI Info Panel Specifications
```css
.ci-info-panel {
  @apply bg-white rounded-lg shadow-sm border border-gray-200 p-6;
}

.ci-title {
  @apply text-xl font-bold text-gray-900 mb-4;
}

.ci-attribute {
  @apply py-2 border-b border-gray-100 last:border-b-0;
}

.ci-attribute-label {
  @apply text-sm font-medium text-gray-500;
}

.ci-attribute-value {
  @apply text-sm text-gray-900;
}
```

### 4. Graph Visualization Page

#### Layout Structure
```
┌─────────────────────────────────────────────────────────┐
│  Graph Controls Bar (Filters + Actions)                  │
├─────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────────────────────────┐   │
│  │   Legend    │  │        Graph Canvas             │   │
│  │   Panel     │  │                               │   │
│  │             │  │                               │   │
│  └─────────────┘  └─────────────────────────────────┘   │
├─────────────────────────────────────────────────────────┤
│  Graph Details Panel (Selected Node Info)                │
└─────────────────────────────────────────────────────────┘
```

#### Graph Visualization Specifications
```css
.graph-canvas {
  @apply w-full h-96 bg-gray-50 rounded-lg border border-gray-200;
}

.graph-controls {
  @apply flex items-center justify-between p-4 bg-white border-b border-gray-200;
}

.graph-legend {
  @apply bg-white rounded-lg shadow-sm border border-gray-200 p-4;
}
```

#### Node Specifications
- **Size**: Based on CI importance or connection count
- **Color**: Based on CI type (Server, Application, Database)
- **Shape**: Circular with icons
- **Hover**: Highlight connections and show details
- **Selection**: Expand to show detailed information

#### Edge Specifications
- **Color**: Based on relationship type
- **Width**: Based on relationship strength or importance
- **Style**: Solid for active, dashed for inactive
- **Arrows**: Directional indicators
- **Labels**: Relationship type on hover

### 5. Forms (CI Create/Edit)

#### Layout Structure
```
┌─────────────────────────────────────────────────────────┐
│  Form Header (Title + Save/Cancel)                       │
├─────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────────────────────────┐   │
│  │   Form      │  │      Preview Panel              │   │
│  │   Fields    │  │                               │   │
│  │             │  │  Live Preview of CI Data       │   │
│  │             │  │                               │   │
│  └─────────────┘  └─────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

#### Form Field Specifications
```css
.form-section {
  @apply mb-6;
}

.form-section-title {
  @apply text-lg font-medium text-gray-900 mb-4 pb-2 border-b border-gray-200;
}

.form-field {
  @apply mb-4;
}

.form-field-label {
  @apply block text-sm font-medium text-gray-700 mb-1;
}

.form-field-help {
  @apply text-xs text-gray-500 mt-1;
}

.form-field-error {
  @apply text-xs text-error-600 mt-1;
}
```

#### Dynamic Attributes
- **Field Type**: Renders appropriate input based on attribute type
- **Validation**: Real-time validation with inline error messages
- **Conditional Fields**: Show/hide based on other field values
- **Autocomplete**: For predefined values and lookups

## Interaction Design Specifications

### Navigation Patterns

#### Breadcrumb Navigation
```css
.breadcrumb {
  @apply flex items-center space-x-2 text-sm;
}

.breadcrumb-item {
  @apply text-gray-500 hover:text-gray-700;
}

.breadcrumb-current {
  @apply text-gray-900 font-medium;
}
```

#### Tab Navigation
```css
.tab-list {
  @apply flex space-x-8 border-b border-gray-200;
}

.tab-button {
  @apply py-2 px-1 border-b-2 font-medium text-sm
         border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300;
}

.tab-button-active {
  @apply border-primary-500 text-primary-600;
}
```

### Search and Filter Interactions

#### Global Search
- **Trigger**: Focus on search or Ctrl+K shortcut
- **Autocomplete**: Real-time suggestions with highlighting
- **Categories**: Group results by type (CI, Relationships, Users)
- **Keyboard Navigation**: Arrow keys to navigate, Enter to select

#### Advanced Filters
- **Collapsible**: Expandable filter sections
- **Multi-select**: Checkbox and tag-based selection
- **Date Ranges**: Calendar picker integration
- **Clear All**: Quick filter reset functionality

### Data Manipulation Interactions

#### Bulk Selection
- **Selection Methods**: Individual checkboxes, select all, range selection
- **Visual Feedback**: Highlighted rows, selection count
- **Bulk Actions**: Context-sensitive action menu
- **Keyboard Shortcuts**: Ctrl+A to select all, Delete to remove

#### Drag and Drop
- **Relationship Creation**: Drag from source CI to target CI
- **Visual Feedback**: Ghost element, drop zones, connection lines
- **Validation**: Real-time validation of relationship constraints
- **Confirmation**: Confirm dialog for relationship creation

### Loading and Feedback States

#### Loading States
```css
.loading-skeleton {
  @apply animate-pulse bg-gray-200 rounded;
}

.loading-spinner {
  @apply animate-spin rounded-full border-2 border-gray-300 border-t-primary-600;
}
```

#### Progress Indicators
- **Linear Progress**: For file uploads and bulk operations
- **Circular Progress**: For inline loading states
- **Step Progress**: For multi-step wizards
- **Percentage**: Clear completion percentage

#### Success/Error Feedback
- **Toast Notifications**: Non-intrusive feedback messages
- **Inline Validation**: Real-time form validation feedback
- **Confirmation Dialogs**: For destructive actions
- **Empty States**: Helpful guidance when no data exists

## Responsive Design Specifications

### Breakpoint System
```css
/* Mobile First Approach */
/* Extra Small Devices (phones) */
@media (max-width: 639px) { }

/* Small Devices (tablets) */
@media (min-width: 640px) { }

/* Medium Devices (landscape tablets, small desktops) */
@media (min-width: 768px) { }

/* Large Devices (desktops) */
@media (min-width: 1024px) { }

/* Extra Large Devices (large desktops) */
@media (min-width: 1280px) { }

/* 2XL Devices (very large desktops) */
@media (min-width: 1536px) { }
```

### Mobile Adaptations

#### Navigation
- **Hamburger Menu**: Collapsible sidebar on mobile
- **Bottom Navigation**: Primary navigation on mobile
- **Swipe Gestures**: Navigation between related items

#### Tables
- **Card View**: Transform tables to cards on mobile
- **Horizontal Scroll**: For complex tables with essential columns
- **Simplified Actions**: Action buttons with icons only

#### Forms
- **Full Width**: Form fields use full available width
- **Stacked Layout**: Single column layout on mobile
- **Mobile Keyboards**: Appropriate input types for mobile keyboards

## Accessibility Specifications

### WCAG 2.1 AA Compliance

#### Color Contrast
- **Normal Text**: 4.5:1 contrast ratio minimum
- **Large Text**: 3:1 contrast ratio minimum
- **Interactive Elements**: 3:1 contrast ratio minimum
- **Color Independence**: Information not conveyed by color alone

#### Keyboard Navigation
- **Tab Order**: Logical tab order through interactive elements
- **Focus Indicators**: Visible focus state for all interactive elements
- **Skip Links**: Skip to main content, skip to navigation
- **Keyboard Shortcuts**: Common shortcuts documented and available

#### Screen Reader Support
```html
<!-- Semantic HTML5 Structure -->
<main role="main" aria-label="Main content">
  <nav aria-label="Main navigation">
    <ul role="menubar">
      <li role="none">
        <a href="/dashboard" role="menuitem" aria-current="page">
          Dashboard
        </a>
      </li>
    </ul>
  </nav>

  <section aria-labelledby="ci-list-heading">
    <h1 id="ci-list-heading">Configuration Items</h1>
    <!-- Table with proper headers -->
    <table role="table" aria-label="Configuration items list">
      <thead>
        <tr>
          <th scope="col">Name</th>
          <th scope="col">Type</th>
          <th scope="col">Status</th>
        </tr>
      </thead>
      <!-- Table body -->
    </table>
  </section>
</main>
```

#### Form Accessibility
```html
<!-- Properly Labeled Form Fields -->
<div class="form-field">
  <label for="ci-name" class="form-field-label">
    CI Name
    <span class="required" aria-label="required">*</span>
  </label>
  <input
    id="ci-name"
    type="text"
    class="form-input"
    aria-required="true"
    aria-describedby="ci-name-help ci-name-error"
  >
  <div id="ci-name-help" class="form-field-help">
    Enter a unique name for the configuration item
  </div>
  <div id="ci-name-error" class="form-field-error" role="alert" aria-live="polite">
    <!-- Error messages appear here -->
  </div>
</div>
```

## Animation and Micro-interactions

### Motion Principles
- **Purposeful**: Animations serve a functional purpose
- **Fast**: Transitions complete quickly (200-300ms)
- **Natural**: Follow natural physical laws
- **Accessible**: Respect prefers-reduced-motion settings

### Transition Specifications
```css
/* Standard Transitions */
.transition-standard {
  transition: all 0.2s ease-in-out;
}

.transition-colors {
  transition: color 0.2s ease-in-out, background-color 0.2s ease-in-out;
}

.transition-transform {
  transition: transform 0.2s ease-in-out;
}

/* Respect user preferences */
@media (prefers-reduced-motion: reduce) {
  .transition-standard,
  .transition-colors,
  .transition-transform {
    transition: none;
  }
}
```

### Micro-interactions

#### Button Interactions
- **Hover**: Subtle color change and elevation
- **Active**: Brief scale transform (0.98) on click
- **Loading**: Spinner with button text change
- **Success**: Brief checkmark animation

#### Card Interactions
- **Hover**: Slight elevation increase and shadow enhancement
- **Focus**: Visible outline for keyboard navigation
- **Selection**: Background color change with checkmark

#### Data Loading
- **Skeleton Screens**: Content placeholders during loading
- **Progressive Loading**: Content appears as it loads
- **Staggered Animations**: Items appear with slight delays

## Component Testing Specifications

### Visual Regression Testing
- **Screenshot Testing**: Automated screenshots for all components
- **Responsive Testing**: Test at all breakpoints
- **Cross-browser Testing**: Chrome, Firefox, Safari, Edge
- **Device Testing**: iOS Safari, Chrome Mobile

### Accessibility Testing
- **Automated Testing**: Axe-core integration
- **Keyboard Navigation**: Manual keyboard testing
- **Screen Reader Testing**: NVDA, VoiceOver testing
- **Color Contrast**: Automated contrast checking

### Performance Testing
- **Load Time**: Page load time under 3 seconds
- **Interaction Time**: Button interactions under 100ms
- **Animation Performance**: 60fps animations
- **Bundle Size**: Optimize bundle size under 1MB

## Design Documentation

### Component Library Documentation
- **Storybook Integration**: Interactive component showcase
- **Usage Guidelines**: When and how to use each component
- **Props Documentation**: Complete API documentation
- **Design Tokens**: Centralized design system variables

### Pattern Library
- **Common Patterns**: Reusable UI patterns documented
- **Layout Templates**: Common layout patterns
- **User Flows**: Step-by-step user journey documentation
- **Design Decisions**: Rationale behind design choices

---

*This UX/UI specification document provides comprehensive guidance for implementing the Pustaka CMDB interface. It ensures consistency, accessibility, and usability across all user interactions while maintaining professional enterprise standards.*