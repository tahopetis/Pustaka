# CSV Import Testing Guide

This document provides step-by-step instructions for manually testing the EA entity CSV import functionality.

## Prerequisites

1. **Docker Services Running:**
   ```bash
   docker compose up -d
   ```

2. **Verify Services:**
   - Frontend: http://localhost:3000
   - API: http://localhost:8080
   - API Health: http://localhost:8080/health

3. **Admin User:**
   - Login as admin user (admin/Admin@123)
   - Export JWT_TOKEN for API tests:
     ```bash
     # Login via browser or API
     export JWT_TOKEN="<your_jwt_token>"
     ```

---

## Test 1: Template Download

### Objective
Verify CSV template generation for a specific CI type.

### Steps

1. **Download Template via API:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/ea/import/template \
     -H "Authorization: Bearer $JWT_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"ci_type":"EA.Application-BusinessApp"}' \
     --output template.csv
   ```

2. **Verify Template:**
   ```bash
   cat template.csv
   ```

3. **Expected Output:**
   - CSV with header row containing: Name*, CI_Type*, Domain*, Lifecycle_Status, Owner, Team, Tags
   - Additional columns for required attributes of the CI type
   - Example row with sample data

4. **Verify Columns Match CI Type:**
   - Check that required attribute columns match EA.Application-BusinessApp schema
   - Verify example data is valid

---

## Test 2: Create Test CSV File

### Objective
Create a test CSV file with valid and invalid rows for validation testing.

### Steps

1. **Copy Template:**
   ```bash
   cp template.csv test-import.csv
   ```

2. **Edit CSV** (using your preferred editor):
   - Add **5 valid rows** with complete data
   - Add **2 rows with errors**:
     - Row 6: Missing Name field
     - Row 7: Invalid lifecycle_status (e.g., "InvalidStatus")

3. **Example CSV Structure:**
   ```csv
   Name,CI_Type,Domain,Lifecycle_Status,Owner,Team,Tags,Attributes
   Customer Portal,EA.Application-BusinessApp,Application,Active,Business Architecture,Enterprise Architecture,customer,web,"{""criticality"":""high""}"
   Order System,EA.Application-BusinessApp,Application,Active,Business Architecture,Enterprise Architecture,critical,backend,"{""criticality"":""critical""}"
   Payment Gateway,EA.Application-BusinessApp,Application,Active,Business Architecture,Enterprise Architecture,payment,external,"{""criticality"":""critical""}"
   Inventory App,EA.Application-BusinessApp,Application,Active,Business Architecture,Enterprise Architecture,internal,logistics,"{""criticality"":""high""}"
   User Auth,EA.Application-BusinessApp,Application,Active,Business Architecture,Enterprise Infrastructure,security,auth,"{""criticality"":""critical""}"
   ,EA.Application-BusinessApp,Application,Active,Business Architecture,Enterprise Architecture,internal,logistics,"{""criticality"":""high""}"
   Test App,EA.Application-BusinessApp,Application,InvalidStatus,Business Architecture,Enterprise Architecture,test,sample,"{""criticality"":""medium""}"
   ```

---

## Test 3: Upload and Validate CSV

### Objective
Verify CSV validation detects and reports errors correctly.

### Steps

1. **Upload CSV for Validation:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/ea/import/validate \
     -H "Authorization: Bearer $JWT_TOKEN" \
     -F "file=@test-import.csv" \
     -F "ci_type=EA.Application-BusinessApp"
   ```

2. **Expected Response:**
   ```json
   {
     "total_rows": 7,
     "error_count": 2,
     "success_count": 5,
     "errors": [
       {
         "row_number": 6,
         "field_name": "Name",
         "error_message": "Name is required",
         "severity": "error"
       },
       {
         "row_number": 7,
         "field_name": "Lifecycle_Status",
         "error_message": "Lifecycle_Status 'InvalidStatus' does not exist",
         "severity": "error"
       }
     ]
   }
   ```

3. **Verify:**
   - total_rows = 7
   - error_count = 2
   - errors array contains 2 errors
   - Error messages are clear and actionable

---

## Test 4: Download Error CSV

### Objective
Verify error CSV download functionality.

### Steps

1. **Download Error CSV:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/ea/import/errors/download \
     -H "Authorization: Bearer $JWT_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "errors": [
         {"row_number": 6, "field_name": "Name", "error_message": "Name is required"},
         {"row_number": 7, "field_name": "Lifecycle_Status", "error_message": "Lifecycle_Status \"InvalidStatus\" does not exist"}
       ]
     }' \
     --output errors.csv
   ```

2. **Verify Error CSV:**
   ```bash
   cat errors.csv
   ```

3. **Expected Format:**
   ```csv
   RowNumber,FieldName,ErrorMessage,ExpectedFormat,ActualValue
   6,Name,"Name is required",,
   7,Lifecycle_Status,"Lifecycle_Status 'InvalidStatus' does not exist",,
   ```

---

## Test 5: Fix CSV and Import

### Objective
Verify that fixing errors allows successful import.

### Steps

1. **Fix CSV Errors:**
   - Edit test-import.csv
   - Add Name for row 6
   - Change InvalidStatus to "Active" for row 7

2. **Re-validate:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/ea/import/validate \
     -H "Authorization: Bearer $JWT_TOKEN" \
     -F "file=@test-import-fixed.csv" \
     -F "ci_type=EA.Application-BusinessApp"
   ```

3. **Expected Response:**
   ```json
   {
     "total_rows": 7,
     "error_count": 0,
     "success_count": 7,
     "errors": []
   }
   ```

4. **Execute Import:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/ea/import/execute \
     -H "Authorization: Bearer $JWT_TOKEN" \
     -F "file=@test-import-fixed.csv" \
     -F "ci_type=EA.Application-BusinessApp"
   ```

5. **Expected Response:**
   ```json
   {
     "success_count": 7,
     "error_count": 0,
     "total_rows": 7
   }
   ```

---

## Test 6: Verify Imported Entities

### Objective
Verify entities were created in the database.

### Steps

1. **List Entities via API:**
   ```bash
   curl -X GET "http://localhost:8080/api/v1/ea/entities?ci_type=EA.Application-BusinessApp" \
     -H "Authorization: Bearer $JWT_TOKEN" | jq
   ```

2. **Expected Output:**
   - JSON array with 7+ entities (5 from test + 7 from import)
   - Each entity has: id, name, ci_type, attributes, tags, data_quality_score

3. **Verify Specific Entity:**
   ```bash
   curl -X GET "http://localhost:8080/api/v1/ea/entities?search=Customer Portal" \
     -H "Authorization: Bearer $JWT_TOKEN" | jq
   ```

4. **Expected:**
   - Entity with name "Customer Portal"
   - ci_type = "EA.Application-BusinessApp"
   - attributes.criticality = "high"

---

## Test 7: Frontend Wizard Flow

### Objective
Verify the 4-step import wizard in the UI.

### Steps

1. **Open Browser:**
   - Navigate to: http://localhost:3000/entities/import

2. **Step 1 - Upload:**
   - Select CI Type: "EA.Application-BusinessApp"
   - Download template (verify template downloads)
   - Upload test-import-fixed.csv
   - Click "Next: Preview"
   - **Verify:** File name displayed, next button enabled

3. **Step 2 - Preview:**
   - **Verify:** Table shows first 10 rows of CSV
   - **Verify:** Column headers match CSV
   - **Verify:** Data displayed correctly
   - Click "Validate"
   - **Verify:** Loading indicator shown

4. **Step 3 - Validate:**
   - **Verify:** Validation results displayed
   - **Verify:** Success message with entity count
   - **Verify:** If errors present, error list shown
   - **Verify:** "Download Error CSV" button available if errors exist
   - Click "Import X Entities"
   - **Verify:** Importing... indicator shown

5. **Step 4 - Import Complete:**
   - **Verify:** Success icon displayed
   - **Verify:** Message "Successfully imported X entities"
   - **Verify:** "View Imported Entities" button navigates to entity list
   - **Verify:** "Import More" button resets wizard

---

## Test Checklist

Complete all items to verify CSV import functionality:

- [ ] Template download works for all EA CI types
- [ ] Template columns match CI type schema
- [ ] Valid CSV uploads and previews correctly
- [ ] Preview shows first 10 rows accurately
- [ ] Validation errors detected (missing required, invalid enums, non-existent references)
- [ ] Error summary displays row number, field, and message
- [ ] Error CSV downloads with correct format
- [ ] Fixed CSV re-validates successfully
- [ ] Import executes without errors
- [ ] Success message shows accurate entity count
- [ ] Entities created in database
- [ ] Entity attributes populated correctly
- [ ] Tags parsed from comma-separated values
- [ ] Frontend wizard shows all 4 steps
- [ ] Step navigation disabled until prerequisites met
- [ ] Import progress indicator works
- [ ] Error handling works for malformed CSV
- [ ] Large file handling (32MB limit)

---

## Troubleshooting

### Issue: Template download fails
- **Check:** JWT_TOKEN is valid and not expired
- **Check:** CI type exists in database (query ea_ci_type_definitions table)
- **Check:** API is accessible (curl http://localhost:8080/health)

### Issue: Validation passes but import fails
- **Check:** Database connection (docker logs pustaka-api)
- **Check:** Required foreign keys exist (lifecycle_statuses, ea_teams)
- **Check:** API logs for specific error messages

### Issue: Frontend wizard not loading
- **Check:** Frontend container running (docker ps)
- **Check:** Browser console for JavaScript errors
- **Check:** API accessible from frontend (network tab in devtools)

### Issue: Import is slow
- **Check:** CSV row count (should be < 1000 rows for manual testing)
- **Check:** Database performance (docker logs pustaka-postgres)
- **Check:** API container resources (docker stats pustaka-api)

---

## Test Data Cleanup

After testing, remove test entities:

```bash
# Get list of entities
curl -X GET "http://localhost:8080/api/v1/ea/entities?ci_type=EA.Application-BusinessApp" \
  -H "Authorization: Bearer $JWT_TOKEN" | jq '.entities[].id'

# Delete each test entity (replace <id> with actual UUID)
curl -X DELETE "http://localhost:8080/api/v1/ea/entities/<id>?force=true" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

Or access via frontend:
1. Navigate to http://localhost:3000/entities/business
2. Filter by CI Type: "EA.Application-BusinessApp"
3. Select test entities and bulk delete

---

## Success Criteria

CSV import functionality is verified when:

1. All API tests pass (template, validate, import, error download)
2. All frontend wizard steps work correctly
3. Error detection and reporting is accurate
4. Entities are created correctly in database
5. Validation errors are properly surfaced to users
6. Error CSV format is correct and downloadable
7. Large files are handled gracefully
8. All checklist items completed

---

## Related Documentation

- Plan: .planning/phases/02-entity-management/02-07-PLAN.md
- Import Service: internal/ea/import_service.go
- Import Handlers: internal/api/handlers/ea_handlers.go (import endpoints)
- Frontend Wizard: web/src/views/ea/ImportWizardView.vue
- Validation Logic: internal/ea/validation.go
