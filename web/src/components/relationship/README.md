# Relationship Type Management System

This document describes the Vue.js frontend components for managing relationship types in the Pustaka CMDB.

## Overview

The relationship type management system provides comprehensive functionality for creating, viewing, editing, and managing relationship types between configuration items. This system integrates with the existing Pustaka architecture and follows established patterns.

## Components

### 1. Store Management

**File:** `/src/stores/relationshipTypes.ts`

- **RelationshipTypeStore**: Pinia store for managing relationship type state
- Provides CRUD operations for relationship types
- Handles loading, caching, and error management
- Computed properties for active relationship types
- Helper methods for finding types by ID or name

### 2. Core Views

#### RelationshipTypeListView
**Path:** `/src/views/relationship-types/RelationshipTypeListView.vue`
**Route:** `/relationship-types`

- Lists all relationship types with pagination
- Search and filtering capabilities
- Shows system vs custom relationship types
- Active/inactive status management
- Responsive table layout with actions

#### RelationshipTypeFormView
**Path:** `/src/views/relationship-types/RelationshipTypeFormView.vue`
**Routes:** `/relationship-types/new`, `/relationship-types/:id/edit`

- Create and edit relationship types
- Form validation for names, labels, and colors
- Live preview of relationship type appearance
- System type protection (system types cannot be edited)
- Color picker and icon support

#### RelationshipTypeDetailsView
**Path:** `/src/views/relationship-types/RelationshipTypeDetailsView.vue`
**Route:** `/relationship-types/:id`

- Comprehensive view of relationship type details
- Usage examples showing forward and reverse relationships
- Metadata and audit information
- Related actions and navigation

### 3. Reusable Components

#### RelationshipTypeSelect
**Path:** `/src/components/relationship/RelationshipTypeSelect.vue`

- Dropdown component for selecting relationship types
- Optional preview showing forward/reverse labels
- Filtering for active/inactive types
- Integration with RelationshipTypeStore
- Customizable styling and labels

#### RelationshipTypeBadge
**Path:** `/src/components/relationship/RelationshipTypeBadge.vue`

- Display component for relationship type labels
- Color-coded badges based on type configuration
- Forward/reverse label support
- Icon integration
- Responsive sizing options

## Features

### 1. Relationship Type Properties

- **Name**: Internal system identifier (underscore_separated)
- **Forward Label**: Display label for source→target relationships
- **Reverse Label**: Display label for target→source relationships
- **Description**: Optional description for usage guidance
- **Color**: Hex color code for visualization
- **Icon**: Optional icon name for enhanced UI
- **Status**: Active/inactive toggle
- **System Type**: Protected system relationship types

### 2. Permission-Based Access Control

The system integrates with Pustaka's RBAC system:

- `relationship-type:read`: View relationship types
- `relationship-type:create`: Create new relationship types
- `relationship-type:update`: Edit existing relationship types
- `relationship-type:delete`: Delete custom relationship types

System relationship types cannot be edited or deleted.

### 3. Search and Filtering

- Search by name, description, or labels
- Filter by active/inactive status
- Pagination support for large datasets
- Real-time search with debouncing

### 4. Integration Points

#### Relationship Management
- Updated relationship form uses RelationshipTypeSelect
- Relationship list shows colored badges
- Enhanced filtering by relationship type

#### Graph Visualization
- Color-coded edges based on relationship type
- Customizable edge labels
- Icon support for relationship types

#### API Integration
- Full CRUD API endpoints
- Error handling and validation
- Optimistic updates where appropriate

## Usage Examples

### Creating a New Relationship Type

```vue
<template>
  <RelationshipTypeFormView />
</template>

<script setup>
import RelationshipTypeFormView from '@/views/relationship-types/RelationshipTypeFormView.vue'
</script>
```

### Using Relationship Type Select

```vue
<template>
  <RelationshipTypeSelect
    v-model="selectedType"
    label="Relationship Type"
    placeholder="Select type"
    :required="true"
    @change="handleTypeChange"
  />
</template>
```

### Displaying Relationship Type Badge

```vue
<template>
  <RelationshipTypeBadge
    :type="relationship.relationship_type"
    :size="'small'"
    :show-icon="true"
  />
</template>
```

## Styling and Theming

- Uses existing Pustaka CSS classes and design system
- Responsive design with mobile-first approach
- Color customization per relationship type
- Badge styling with proper contrast ratios
- Loading states and error handling

## File Structure

```
src/
├── stores/
│   └── relationshipTypes.ts          # Pinia store
├── views/relationship-types/
│   ├── RelationshipTypeListView.vue  # List view
│   ├── RelationshipTypeFormView.vue  # Create/edit form
│   └── RelationshipTypeDetailsView.vue # Details view
├── components/relationship/
│   ├── RelationshipTypeSelect.vue   # Dropdown component
│   ├── RelationshipTypeBadge.vue    # Display badge
│   └── README.md                    # This documentation
└── services/api.ts                   # API service updates
```

## API Integration

The frontend integrates with backend API endpoints:

- `GET /relationship-types` - List relationship types
- `POST /relationship-types` - Create relationship type
- `GET /relationship-types/:id` - Get relationship type details
- `PUT /relationship-types/:id` - Update relationship type
- `DELETE /relationship-types/:id` - Delete relationship type

## Browser Compatibility

- Modern browsers (Chrome 90+, Firefox 88+, Safari 14+)
- ES2022 features supported
- Responsive design works on mobile devices
- Accessible markup with ARIA labels

## Future Enhancements

Potential improvements for future releases:

1. **Bulk Operations**: Mass activate/deactivate relationship types
2. **Import/Export**: CSV or JSON import/export functionality
3. **Usage Analytics**: Show relationship type usage statistics
4. **Templates**: Pre-defined relationship type templates
5. **Validation Rules**: Advanced validation for relationship creation
6. **Audit Trail**: Full audit history for relationship type changes

## Testing

Components should be tested with:

- Unit tests for store operations
- Component tests for form validation
- Integration tests for API calls
- E2E tests for user workflows
- Accessibility testing (WCAG 2.1 AA compliance)