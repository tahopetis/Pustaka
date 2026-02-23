# Quick Start Instructions

## Worktree Setup

The worktree has already been created at:
```bash
/home/tahopetis/dev/pustaka-ci-sidebar-refactor
```

### Navigate to Worktree
```bash
cd /home/tahopetis/dev/pustaka-ci-sidebar-refactor
```

### Verify Current Branch
```bash
git branch
# Should show: * feature/ci-sidebar-refactor
```

---

## Implementation Steps

### Step 1: Create CITypeAutocomplete Component
```bash
# Create the component file
cat > web/src/components/base/CITypeAutocomplete.vue << 'EOF'
<template>
  <!-- Copy content from PLAN.md Task 1 -->
</template>
EOF
```

**Reference**: See `PLAN.md` Task 1 for full implementation details.

---

### Step 2: Refactor AppSidebar
```bash
# Edit the sidebar
nano web/src/components/layout/AppSidebar.vue
# OR use your preferred editor
```

**Reference**: See `PLAN.md` Task 2 for exact code to replace/add.

---

### Step 3: Update CIFormView
```bash
# Edit the CI form
nano web/src/views/ci/CIFormView.vue
```

**Reference**: See `PLAN.md` Task 3 for exact changes needed.

---

### Step 4: (Optional) Update Router
```bash
# Edit router
nano web/src/router/index.ts
```

**Reference**: See `PLAN.md` Task 4 for convenience routes.

---

## Testing Your Changes

### Option A: Test with Docker (Recommended)

```bash
# From worktree directory
cd /home/tahopetis/dev/pustaka-ci-sidebar-refactor

# Stop existing containers
docker compose down

# Rebuild and start
docker compose up --build -d

# Check logs
docker logs -f pustaka-sidebar-refactor-frontend
docker logs -f pustaka-sidebar-refactor-api

# Access app
# Frontend: http://localhost:3001
# API: http://localhost:8081
```

### Option B: Test with Dev Server

```bash
# Terminal 1 - Backend
cd /home/tahopetis/dev/pustaka-ci-sidebar-refactor
go run cmd/api/main.go

# Terminal 2 - Frontend
cd /home/tahopetis/dev/pustaka-ci-sidebar-refactor/web
npm run dev
```

---

## Verification Checklist

Test each item from `PLAN.md` Testing Checklist:

### Sidebar Tests
- [ ] Click "Configuration Items" - should expand/collapse
- [ ] Click "Asset Management CI" - should go to `/ci/new?context=asset`
- [ ] Click "Enterprise Architecture CI" - should expand/collapse
- [ ] Verify all 8 EA domains appear
- [ ] Click each EA domain - should navigate with correct query params

### CI Form Tests
- [ ] Asset context: Only shows non-EA CI types
- [ ] EA Strategy context: Only shows `EA.Strategy*` types
- [ ] EA Application context: Only shows `EA.Application*` types
- [ ] Autocomplete shows max 5 results
- [ ] Search/filter works in autocomplete
- [ ] Create CI/entity works successfully

---

## Git Workflow

### Check Changes
```bash
git status
git diff
```

### Commit When Complete
```bash
git add .
git commit -m "refactor(ci): reorganize sidebar with collapsible CI menus

- Add collapsible Configuration Items menu
- Separate Asset Management and EA contexts
- Add 8 EA domain sub-menus under EA CI
- Create CITypeAutocomplete component with domain filtering
- Update CIFormView to support context-based CI type filtering

BREAKING CHANGE: Query params now used for CI context
"
```

### Push to Remote (if needed)
```bash
git push -u origin feature/ci-sidebar-refactor
```

---

## Troubleshooting

### Issue: Autocomplete not filtering correctly
**Check**: Verify domain prop matches expected values ('asset', 'ea', 'all')
**Check**: Ensure CI type names follow `EA.*` pattern

### Issue: Sidebar not collapsible
**Check**: Ensure state variables (`ciMenuOpen`, `eaMenuOpen`) are defined
**Check**: Verify click handlers are bound correctly

### Issue: Query params not working
**Check**: Ensure `vue-router` is reading `route.query` correctly
**Check**: Verify router links use `:to` with query object syntax

### Issue: Docker not picking up changes
**Solution**:
```bash
docker compose down
docker compose up --build -d
```

---

## Rollback

If you need to start over:
```bash
# Discard all changes
git reset --hard HEAD

# Or switch back to main
git checkout main
```

---

## Next Steps After Implementation

1. **Test thoroughly** using the checklist
2. **Get user feedback** on the new menu structure
3. **Refine autocomplete** UX if needed
4. **Consider adding** keyboard navigation to autocomplete
5. **Document** the new menu structure in user docs

---

## Support Files

- `PLAN.md` - Detailed implementation plan with code
- `INSTRUCTIONS.md` - This file
- Current worktree: `/home/tahopetis/dev/pustaka-ci-sidebar-refactor`
