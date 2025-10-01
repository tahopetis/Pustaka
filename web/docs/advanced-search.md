# Advanced CI Search

The Pustaka CMDB now includes powerful JSONB-aware search functionality that allows users to search within CI attributes, not just basic CI fields.

## Features

### Basic Search
- **Quick Search**: Search by CI name, type, and tags
- **CI Type Filtering**: Filter by specific CI types
- **Sorting**: Sort by name, type, creation date, or update date
- **Ordering**: Ascending or descending order

### Advanced Search

Click "Show Advanced" to access powerful attribute-based search capabilities:

#### Attribute Search
1. **Select CI Type**: Choose a CI type to search its attributes
2. **Dynamic Filters**: The search interface automatically generates filters based on the CI type's schema
3. **Type-Specific Inputs**:
   - **String fields**: Text search with partial matching
   - **Enum fields**: Dropdown selection of predefined values
   - **Integer fields**: Range filters (min/max values)
   - **Boolean fields**: True/False/Any selection
   - **Array/Object fields**: JSON content search

#### Tag Search
- Add multiple tags to filter CIs
- Tags are combined with AND logic (CI must have all specified tags)

#### Active Filters Display
- Visual representation of all active search filters
- Individual filter removal capability
- Filter count and result summary

## Search Examples

### Example 1: Find Servers with Specific Configuration
```
CI Type: Server
Attributes:
- CPU Cores: 8-16 (range)
- Memory GB: ≥ 32
- OS Type: "Ubuntu"
- Environment: "production"
```

### Example 2: Search Network Devices
```
CI Type: Network Device
Attributes:
- Device Type: "router"
- Management IP: "192.168.*" (pattern search)
- Is Managed: true
```

### Example 3: Find Applications by Version
```
CI Type: Application
Attributes:
- Version: "1.2.*"
- Language: "Java"
- Database: "PostgreSQL"
Tags: "critical", "web"
```

## Technical Implementation

### Frontend Components

#### AdvancedSearch.vue
- Dynamic form generation based on CI type schemas
- Real-time filter updates
- Support for all attribute types (string, integer, boolean, array, object)
- Range filters for numeric attributes
- Enum dropdowns for constrained values
- JSON search for complex data types

#### CIListView.vue
- Integration with advanced search component
- Clean filter management
- Result pagination
- Search state management

### API Integration

The search functionality sends structured queries to the backend:

```typescript
// Example search request
{
  search: "web-server",
  ci_type: "Server",
  sort: "name",
  order: "asc",
  tags: ["production", "critical"],
  attributes: {
    cpu_cores: { min: 8, max: 16 },
    memory_gb: { min: 32 },
    os_type: "Ubuntu",
    environment: "production"
  },
  page: 1,
  limit: 20
}
```

### Backend Support

The backend API supports:
- Basic text search across CI names and descriptions
- Tag filtering with multiple tags
- JSONB attribute queries with type-specific operators
- Range queries for numeric attributes
- Pattern matching for string attributes
- Exact matching for enum values

## User Interface

### Search Controls
- **Collapsible Advanced Section**: Toggle between basic and advanced search
- **Dynamic Form Generation**: Filters appear based on selected CI type
- **Real-time Updates**: Results update as filters are applied
- **Filter Management**: Easy addition and removal of individual filters

### Visual Feedback
- **Loading States**: Clear indication during search operations
- **Result Counts**: Display total number of matching results
- **Active Filter Badges**: Visual representation of applied filters
- **Clear All Option**: Reset all search criteria at once

## Performance Considerations

### Client-Side
- Debounced search inputs to reduce API calls
- Efficient filter state management
- Lazy loading of CI type schemas

### Server-Side (Expected Implementation)
- Indexed JSONB queries for performance
- Query optimization for complex attribute searches
- Pagination support for large result sets
- Caching of frequently used CI type schemas

## Future Enhancements

### Planned Features
- **Saved Searches**: Save and reuse complex search criteria
- **Search History**: Recent search queries
- **Export Results**: Export search results to CSV/Excel
- **Search Templates**: Pre-configured search templates for common use cases
- **Full-Text Search**: Enhanced text search with relevance scoring
- **Graph Search**: Search based on CI relationships

### Advanced Filtering
- **Date Range Filters**: For date/time attributes
- **Regex Search**: Advanced pattern matching
- **Logical Operators**: AND/OR/NOT combinations for complex queries
- **Cross-Type Search**: Search across multiple CI types simultaneously

## Usage Tips

1. **Start Broad**: Begin with basic search, then refine with advanced filters
2. **Use CI Type Selection**: Selecting a CI type enables attribute-specific filters
3. **Combine Filters**: Use multiple attribute filters for precise results
4. **Save Common Searches**: Note frequently used filter combinations
5. **Use Range Filters**: For numeric attributes, ranges are more effective than exact values
6. **Tag Strategy**: Use tags consistently to enable effective filtering

## Troubleshooting

### Common Issues
- **No Results**: Try removing filters or broadening search criteria
- **Slow Performance**: Reduce the number of active filters or use pagination
- **Missing Attributes**: Ensure the correct CI type is selected for attribute search

### Best Practices
- Use specific CI types when doing attribute searches
- Combine basic and advanced search for best results
- Use appropriate filter types for each attribute type
- Clear filters between different search scenarios