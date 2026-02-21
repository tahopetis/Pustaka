# Warn-But-Allow Validation Testing Guide

This document provides step-by-step instructions for testing the warn-but-allow validation behavior for EA entities.

## Overview

The warn-but-allow pattern allows EA entities to be saved even when validation warnings exist, while still tracking data quality through the `data_quality_score` field.

**Expected Behavior:**
- Entities with `data_quality_score < 100` are saved with warnings
- Response includes `validation_warnings` array
- Critical errors (invalid CI type, missing required fields) still block save (422)
- Non-critical warnings allow save (201/200 with warnings)

---

## Test 1: Create Entity with Missing Recommended Fields

### Objective
Verify entity is saved with warnings when recommended fields are missing.

### API Test

```bash
# First, get a valid lifecycle_status_id
export JWT_TOKEN="<your_jwt_token>"
LIFECYCLE_ID=$(curl -s -X GET "http://localhost:8080/api/v1/lifecycle-statuses" \
  -H "Authorization: Bearer $JWT_TOKEN" | jq -r '.[] | select(.name=="Active") | .id')

# Create entity with only required fields (missing recommended attributes)
curl -X POST http://localhost:8080/api/v1/ea/entities \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Test App Minimal\",
    \"ci_type\": \"EA.Application-BusinessApp\",
    \"lifecycle_status_id\": \"$LIFECYCLE_ID\",
    \"owner\": \"Business Architecture\",
    \"attributes\": {
      \"criticality\": \"high\"
    }
  }" | jq '.'
```

### Expected Response

```json
{
  "id": "<uuid>",
  "name": "Test App Minimal",
  "ci_type": "EA.Application-BusinessApp",
  "owner": "Business Architecture",
  "attributes": {
    "criticality": "high",
    "ea_domain": "Application",
    "ea_owner": "Business Architecture"
  },
  "tags": ["Application"],
  "lifecycle_status": {
    "id": "<uuid>",
    "name": "Active",
    "display_name": "Active"
  },
  "data_quality_score": 75.5,
  "validation_warnings": [
    "Data quality score is 75.5% (recommended: 100%)"
  ],
  "created_at": "2026-02-21T04:30:00Z",
  "updated_at": "2026-02-21T04:30:00Z"
}
```

### Verification

- [ ] Status code is `201 Created`
- [ ] Entity saved to database
- [ ] `data_quality_score < 100`
- [ ] `validation_warnings` array is not empty
- [ ] Warning message includes score percentage

---

## Test 2: Create Entity with All Recommended Fields

### Objective
Verify perfect data quality score when all fields provided.

### API Test

```bash
curl -X POST http://localhost:8080/api/v1/ea/entities \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Test App Complete\",
    \"ci_type\": \"EA.Application-BusinessApp\",
    \"lifecycle_status_id\": \"$LIFECYCLE_ID\",
    \"owner\": \"Business Architecture\",
    \"description\": \"Complete application with all fields\",
    \"attributes\": {
      \"criticality\": \"high\",
      \"description\": \"Business application for testing\",
      \"version\": \"1.0.0\",
      \"support_contact\": \"admin@example.com\"
    },
    \"tags\": [\"test\", \"complete\"]
  }" | jq '.'
```

### Expected Response

```json
{
  "id": "<uuid>",
  "name": "Test App Complete",
  "data_quality_score": 100,
  "validation_warnings": [],
  "created_at": "2026-02-21T04:31:00Z",
  "updated_at": "2026-02-21T04:31:00Z"
}
```

### Verification

- [ ] Status code is `201 Created`
- [ ] `data_quality_score = 100`
- [ ] `validation_warnings` array is empty

---

## Test 3: Create Entity with Invalid CI Type (Critical Error)

### Objective
Verify critical errors block save with 422 response.

### API Test

```bash
curl -X POST http://localhost:8080/api/v1/ea/entities \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Test Invalid\",
    \"ci_type\": \"InvalidCIType\",
    \"lifecycle_status_id\": \"$LIFECYCLE_ID\",
    \"owner\": \"Business Architecture\"
  }" | jq '.'
```

### Expected Response

```json
{
  "error": "Critical validation error: invalid EA CI type"
}
```

### Verification

- [ ] Status code is `422 Unprocessable Entity`
- [ ] Entity NOT saved to database
- [ ] Error message indicates critical validation failure

---

## Test 4: Create Entity with Invalid Enum Value (Critical Error)

### Objective
Verify invalid enum values block save.

### API Test

```bash
curl -X POST http://localhost:8080/api/v1/ea/entities \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Test Invalid Enum\",
    \"ci_type\": \"EA.Application-BusinessApp\",
    \"lifecycle_status_id\": \"$LIFECYCLE_ID\",
    \"owner\": \"Business Architecture\",
    \"attributes\": {
      \"criticality\": \"invalid_value\"
    }
  }" | jq '.'
```

### Expected Response

```json
{
  "error": "Critical validation error: criticality must be one of: critical, high, medium, low"
}
```

### Verification

- [ ] Status code is `422 Unprocessable Entity`
- [ ] Entity NOT saved
- [ ] Error message specifies invalid enum value

---

## Test 5: Update Entity to Improve Data Quality

### Objective
Verify updating entity improves data quality score.

### API Test

```bash
# Get entity ID from Test 1
ENTITY_ID="<entity_id_from_test_1>"

# Update entity with additional attributes
curl -X PUT "http://localhost:8080/api/v1/ea/entities/$ENTITY_ID" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"description\": \"Now with description\",
    \"attributes\": {
      \"criticality\": \"high\",
      \"description\": \"Improved application\",
      \"version\": \"1.0.0\"
    }
  }" | jq '.'
```

### Expected Response

```json
{
  "id": "$ENTITY_ID",
  "name": "Test App Minimal",
  "data_quality_score": 90,
  "validation_warnings": [
    "Data quality score is 90.0% (recommended: 100%)"
  ],
  "updated_at": "2026-02-21T04:32:00Z"
}
```

### Verification

- [ ] Status code is `200 OK`
- [ ] `data_quality_score` improved (from 75.5 to 90+)
- [ ] `updated_at` timestamp changed
- [ ] Warnings reduced or eliminated

---

## Test 6: Frontend Form Display of Warnings

### Objective
Verify frontend displays warnings and data quality score.

### Steps

1. **Navigate to Entity Form:**
   - Open http://localhost:3000/entities/new
   - Select CI Type: "EA.Application-BusinessApp"

2. **Fill Required Fields Only:**
   - Name: "Test Frontend Warning"
   - Owner: "Business Architecture"
   - Lifecycle Status: "Active"
   - Attributes → Criticality: "high"

3. **Submit Form:**
   - Click "Create"
   - **Verify:** Form submits successfully (not blocked)

4. **Check Response Display:**
   - **Verify:** Success notification shows: "Entity created with data quality score: 75.5%"
   - **Verify:** Warning badge displays (yellow/orange for score < 80)
   - **Verify:** Warning messages listed below form fields

5. **Add Recommended Fields:**
   - Description: "Test application"
   - Version: "1.0"
   - Support Contact: "admin@example.com"

6. **Submit Again:**
   - Click "Update"
   - **Verify:** Success notification shows: "Entity updated with data quality score: 100%"
   - **Verify:** Green badge for perfect score

### Verification

- [ ] Form submits with warnings (not blocked)
- [ ] Data quality score badge visible (color-coded)
- [ ] Warning messages displayed in UI
- [ ] Score improves when fields added
- [ ] Green badge for score = 100

---

## Test 7: Data Quality Score Color Coding

### Objective
Verify score badge colors match thresholds.

### Test Matrix

| Score Range | Color | Badge Class |
|-------------|-------|-------------|
| 100         | Green | `bg-green-100 text-green-800` |
| 80-99       | Green | `bg-green-100 text-green-800` |
| 60-79       | Yellow | `bg-yellow-100 text-yellow-800` |
| < 60        | Red | `bg-red-100 text-red-800` |

### API Test

Create entities with different completeness levels and verify badge colors in frontend.

---

## Test 8: Verify Database Storage

### Objective
Verify data_quality_score persisted correctly.

### Database Query

```bash
docker exec -it pustaka-postgres psql -U pustaka -d pustaka -c "
  SELECT
    id,
    name,
    ci_type,
    data_quality_score,
    attributes
  FROM configuration_items
  WHERE ci_type LIKE 'EA.%'
  ORDER BY created_at DESC
  LIMIT 5;
"
```

### Expected Output

```
                  id                  |        Name        |              ci_type              | data_quality_score |                           attributes
--------------------------------------+--------------------+----------------------------------+--------------------+--------------------------------------------------------------
 <uuid>                                | Test App Complete  | EA.Application-BusinessApp       |              100 | {"criticality": "high", "description": "..." }
 <uuid>                                | Test App Minimal   | EA.Application-BusinessApp       |             75.5 | {"criticality": "high", "ea_domain": "Application" }
```

### Verification

- [ ] `data_quality_score` stored in database
- [ ] Scores match API responses
- [ ] Attributes include EA metadata

---

## Test Checklist

Complete all items to verify warn-but-allow functionality:

### Backend API
- [ ] Create with missing fields returns 201 with warnings
- [ ] Create with all fields returns 201 with no warnings
- [ ] Create with invalid CI type returns 422
- [ ] Create with invalid enum returns 422
- [ ] Update with additional fields improves score
- [ ] Response includes `data_quality_score` field
- [ ] Response includes `validation_warnings` array
- [ ] Critical errors block save (422)
- [ ] Non-critical warnings allow save (201)

### Frontend
- [ ] Data quality score badge visible on entity form
- [ ] Badge color-coded (green ≥80, yellow ≥60, red <60)
- [ ] Warning messages displayed below affected fields
- [ ] Form submits successfully with warnings
- [ ] Success notification shows score percentage
- [ ] Score improves when fields added
- [ ] Form blocked only on critical errors

### Database
- [ ] `data_quality_score` persisted correctly
- [ ] Scores match API responses
- [ ] Query by score range works

---

## Troubleshooting

### Issue: Entity saves with score = 100 when fields missing
- **Check:** CI type schema (are required attributes defined?)
- **Check:** Validation logic in `internal/ea/validation.go`
- **Check:** Score calculation in `CalculateDataQualityScore`

### Issue: Form blocked on warnings (should allow save)
- **Check:** Frontend form validation logic
- **Check:** Error vs warning handling
- **Check:** Response status code (should be 201, not 422)

### Issue: Warnings not displayed in UI
- **Check:** Frontend response parsing
- **Check:** `validation_warnings` array in response
- **Check:** DynamicFormBuilder warning display logic

### Issue: Score not updating after edit
- **Check:** Update endpoint returns new score
- **Check:** Frontend refreshes entity data after update
- **Check:** Database value changed

---

## Related Documentation

- Plan: .planning/phases/02-entity-management/02-07-PLAN.md
- Handlers: internal/api/handlers/ea_handlers.go (CreateEAEntity, UpdateEAEntity)
- Service: internal/ea/service.go (CreateEntity, UpdateEntity)
- Validation: internal/ea/validation.go (ValidateEntityAttributes, CalculateDataQualityScore)
- Frontend Form: web/src/views/ea/EntityFormView.vue
- Form Builder: web/src/components/ea/DynamicFormBuilder.vue
