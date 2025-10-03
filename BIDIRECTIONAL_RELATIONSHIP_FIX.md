# Bidirectional Relationship Display Fix

## Problem Statement
When viewing a Configuration Item (CI) that is the **target** of a relationship, the relationships section showed "No relationships found" instead of showing the incoming relationship.

For example:
- CI A "runs_on" CI B
- When viewing CI A (source), the relationship showed correctly: "runs_on → CI B"
- When viewing CI B (target), it showed "No relationships found" instead of showing that CI A runs_on it

## Root Cause Analysis
The issue was in the `CIDetailsView.vue` component's `loadRelationships` function, which only queried for relationships where the current CI was the **source**:

```javascript
// BEFORE: Only queried outgoing relationships
const response = await relationshipAPI.list({
  source_id: ci.value.id,
  limit: 50
})
```

This missed all incoming relationships where the current CI was the target.

## Solution Implemented

### 1. Backend Verification
Verified that the backend already supports bidirectional querying:
- ✅ `ListRelationshipFilters` struct includes both `SourceID` and `TargetID` fields
- ✅ `relationship_handlers.go` accepts both `source_id` and `target_id` query parameters
- ✅ `repository.go` correctly handles both filters in the SQL query
- ✅ Frontend API service (`api.ts`) supports both parameters

### 2. Frontend Changes
Modified `/home/syam/dev/Pustaka/web/src/views/ci/CIDetailsView.vue`:

#### A. Updated `loadRelationships` Function
```javascript
// AFTER: Query both outgoing and incoming relationships
const [sourceResponse, targetResponse] = await Promise.all([
  relationshipAPI.list({
    source_id: ci.value.id,  // CI as source (outgoing)
    limit: 50
  }),
  relationshipAPI.list({
    target_id: ci.value.id,  // CI as target (incoming)
    limit: 50
  })
])

// Combine both results and remove duplicates
const allRelationships = [
  ...(sourceResponse.data.relationships || []),
  ...(targetResponse.data.relationships || [])
]
const uniqueRelationships = allRelationships.filter((rel, index, arr) =>
  arr.findIndex(r => r.id === rel.id) === index
)
```

#### B. Updated `getRelationshipDisplayText` Function
```javascript
const getRelationshipDisplayText = (rel: Relationship) => {
  const sourceName = rel.source_ci?.name || `Unknown (${rel.source_id?.slice(0, 8)}...)`
  const targetName = rel.target_ci?.name || `Unknown (${rel.target_id?.slice(0, 8)}...)`

  // Check if current CI is the source or target
  if (ci.value && rel.source_id === ci.value.id) {
    // Outgoing relationship: Current CI is the source
    return `${sourceName} → ${targetName}`
  } else {
    // Incoming relationship: Current CI is the target
    return `${sourceName} ← ${targetName}`
  }
}
```

## Technical Details

### API Flow
1. **Frontend**: Makes two parallel API calls to `/api/v1/relationships`
   - One with `?source_id={ciId}` (outgoing relationships)
   - One with `?target_id={ciId}` (incoming relationships)
2. **Backend**: Each call is handled by `relationshipHandlers.ListRelationships`
3. **Repository**: `ListRelationships` method applies appropriate SQL filters:
   ```sql
   WHERE source_id = $1   -- for outgoing
   WHERE target_id = $1   -- for incoming
   ```
4. **Frontend**: Combines results, removes duplicates, fetches CI details, and displays

### Performance Considerations
- **Parallel API Calls**: Uses `Promise.all()` to fetch both source and target relationships concurrently
- **Deduplication**: Removes any duplicate relationships (unlikely but defensive)
- **Limiting**: Each query is limited to 50 results to prevent excessive data transfer
- **Efficient CI Loading**: Only fetches additional CI details when needed

### Display Logic
- **Outgoing Relationships**: Display as "Source → Target"
- **Incoming Relationships**: Display as "Source ← Target" (arrow points toward current CI)

## Testing
The fix includes:
- ✅ **Build Verification**: Frontend compiles successfully with no TypeScript errors
- ✅ **Backend Verification**: Go code compiles successfully
- ✅ **Test Script**: Created `test_bidirectional_relationships.js` for manual verification

## Files Modified
1. `/home/syam/dev/Pustaka/web/src/views/ci/CIDetailsView.vue` - Main fix implementation
2. `/home/syam/dev/Pustaka/test_bidirectional_relationships.js` - Test script (created)
3. `/home/syam/dev/Pustaka/BIDIRECTIONAL_RELATIONSHIP_FIX.md` - Documentation (created)

## Verification Steps
To verify the fix works correctly:

1. Create two CIs (e.g., "web-server-01" and "database-01")
2. Create a relationship: web-server-01 "runs_on" database-01
3. View "web-server-01" details → Should show outgoing relationship: "web-server-01 → database-01"
4. View "database-01" details → Should show incoming relationship: "web-server-01 ← database-01"

Both CIs should now display all their relationships, regardless of whether they are the source or target.