# Amortization Module Metrics Enhancement Plan

## Overview

Enhance the amortization module to display comprehensive accounting metrics using standard accounting terminology (GVB, NBV, OCC, AD, SV).

**Status**: Planning Phase
**Created**: 2025-12-30
**Priority**: High

---

## Accounting Terminology & Definitions

### Core Metrics

| Acronym | Full Name | Definition | Formula |
|---------|-----------|------------|---------|
| **OCC** | Original Capitalized Cost | Initial purchase cost of assets | Σ(purchase_cost) |
| **GVB** | Gross Book Value | Total capitalized investment including adjustments | OCC + Capitalized Enhancements ± Accounting Corrections |
| **AD** | Accumulated Depreciation | Total depreciation expense recorded to date | Σ(depreciation + write_offs) |
| **NBV** | Net Book Value | Current value after depreciation | GVB - AD |
| **SV** | Salvage Value | Total residual value at end of useful life | Σ(salvage_value) |

### Derived Metrics

| Metric | Formula | Description |
|--------|---------|-------------|
| **Depreciation %** | AD / GVB × 100 | Percentage of asset value consumed |
| **Remaining %** | NBV / GVB × 100 | Percentage of asset value remaining |
| **Adjusted Amount** | GVB - OCC | Net impact of enhancements/corrections |
| **Depreciable Base** | GVB - SV | Total amount to be depreciated |

---

## Current State Analysis

### What We Have:
✅ Purchase Cost per CI
✅ Salvage Value per CI
✅ Current Book Value (NBV) per CI
✅ Adjustment entries in ledger
✅ Monthly depreciation tracking
✅ Accumulated depreciation calculation

### What's Missing:
❌ OCC (Original Capitalized Cost) - Total/Summary level
❌ GVB (Gross Book Value) - Total/Summary level
❌ Proper categorization of adjustments
❌ Visual reference lines for OCC and GVB in charts
❌ Standard accounting terminology in UI

---

## Implementation Plan

### Phase 1: Backend Data Model

#### 1.1 Update Response Models

**File**: `internal/amortization/models.go`

```go
// DepreciationScheduleResponse enhanced with new metrics
type DepreciationScheduleResponse struct {
    Currency              string                   `json:"currency"`
    StartDate             time.Time                `json:"start_date"`
    EndDate               time.Time                `json:"end_date"`

    // New comprehensive metrics
    TotalOriginalCost     float64                  `json:"total_original_cost"`      // OCC
    TotalGrossBookValue   float64                  `json:"total_gross_book_value"`   // GVB
    TotalNetBookValue     float64                  `json:"total_net_book_value"`     // NBV (current)
    TotalSalvageValue     float64                  `json:"total_salvage_value"`      // SV
    TotalAccumulatedDepreciation float64           `json:"total_accumulated_depreciation"` // AD

    Summary               ScheduleSummary          `json:"summary"`
    MonthlyData           []MonthlyScheduleEntry   `json:"monthly_data"`
    ByCIType              []CITypeScheduleSummary  `json:"by_ci_type,omitempty"`
    ByAsset               []AssetScheduleSummary   `json:"by_asset,omitempty"`
}

// MonthlyScheduleEntry enhanced with GVB tracking
type MonthlyScheduleEntry struct {
    Month                   time.Time `json:"month"`
    IsProjected             bool      `json:"is_projected"`
    OpeningBookValue        float64   `json:"opening_book_value"`        // NBV opening
    GrossBookValue          float64   `json:"gross_book_value"`          // GVB for this month
    DepreciationAmount      float64   `json:"depreciation_amount"`
    WriteOffAmount          float64   `json:"write_off_amount"`
    AdjustmentAmount        float64   `json:"adjustment_amount"`        // ± impact to GVB
    AccumulatedDepreciation float64   `json:"accumulated_depreciation"` // Running AD
    ClosingBookValue        float64   `json:"closing_book_value"`        // NBV closing
    ActiveAssetsCount       int       `json:"active_assets_count"`
}

// ScheduleSummary enhanced
type ScheduleSummary struct {
    TotalOriginalCost        float64 `json:"total_original_cost"`        // OCC
    TotalGrossBookValue      float64 `json:"total_gross_book_value"`     // GVB
    TotalNetBookValue        float64 `json:"total_net_book_value"`        // NBV
    TotalDepreciation        float64 `json:"total_depreciation"`          // AD
    TotalWriteOffs           float64 `json:"total_write_offs"`
    TotalAdjustments         float64 `json:"total_adjustments"`           // Net ±
    TotalSalvageValue        float64 `json:"total_salvage_value"`         // SV
    AverageMonthlyExpense    float64 `json:"average_monthly_expense"`
    ProjectedEndValue        float64 `json:"projected_end_value"`
    DepreciationPercentage   float64 `json:"depreciation_percentage"`    // AD/GVB × 100
    RemainingPercentage       float64 `json:"remaining_percentage"`       // NBV/GVB × 100
}
```

#### 1.2 Repository Calculation Updates

**File**: `internal/amortization/repository.go`

**New Calculations in `GetDepreciationScheduleData`:**

```go
// After loading assets
totalOriginalCost := 0.0     // OCC
totalGrossBookValue := 0.0    // GVB
totalNetBookValue := 0.0       // NBV
totalSalvageValue := 0.0       // SV

for assetRows.Next() {
    // ... existing scanning ...

    totalOriginalCost += asset.PurchaseCost
    totalGrossBookValue += asset.PurchaseCost // Will add adjustments later
    totalNetBookValue += asset.CurrentBookValue
    totalSalvageValue += asset.SalvageValue
}

// Add adjustments to GVB
// Sum all positive adjustments (enhancements) and negative adjustments (corrections)
adjustmentsQuery := `
    SELECT
        COALESCE(SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END), 0) as positive_adj,
        COALESCE(SUM(CASE WHEN amount < 0 THEN ABS(amount) ELSE 0 END), 0) as negative_adj
    FROM amortization_ledger ale
    WHERE ale.entry_type = 'adjustment'
    ` + whereClause

totalGrossBookValue += netAdjustments // Add to original cost

// Running totals for monthly entries
runningGrossBookValue := 0.0       // GVB changes with adjustments
runningNetBookValue := 0.0          // NBV changes with depreciation
runningAccumulatedDepreciation := 0.0
```

### Phase 2: Frontend Type Definitions

#### 2.1 TypeScript Interfaces

**File**: `web/src/types/amortization.ts`

```typescript
export interface DepreciationScheduleResponse {
  currency: string
  start_date: string
  end_date: string

  // New comprehensive metrics
  total_original_cost: number           // OCC
  total_gross_book_value: number        // GVB
  total_net_book_value: number          // NBV
  total_salvage_value: number           // SV
  total_accumulated_depreciation: number // AD

  summary: ScheduleSummary
  monthly_data: MonthlyScheduleEntry[]
  by_ci_type?: CITypeScheduleSummary[]
  by_asset?: AssetScheduleSummary[]
}

export interface MonthlyScheduleEntry {
  month: string
  is_projected: boolean
  opening_book_value: number          // NBV opening
  gross_book_value: number            // GVB for this month
  depreciation_amount: number
  write_off_amount: number
  adjustment_amount: number           // ± impact to GVB
  accumulated_depreciation: number    // Running AD
  closing_book_value: number          // NBV closing
  active_assets_count: number
}

export interface ScheduleSummary {
  total_original_cost: number          // OCC
  total_gross_book_value: number        // GVB
  total_net_book_value: number          // NBV
  total_depreciation: number            // AD
  total_write_offs: number
  total_adjustments: number             // Net ±
  total_salvage_value: number           // SV
  average_monthly_expense: number
  projected_end_value: number
  depreciation_percentage: number       // AD/GVB × 100
  remaining_percentage: number          // NBV/GVB × 100
}
```

### Phase 3: Frontend UI - Summary Cards

#### 3.1 Card Layout

**File**: `web/src/views/amortization/ReportsView.vue`

```vue
<!-- 5 Metrics Cards in a Row -->
<div class="metrics-grid">
  <!-- OCC Card -->
  <div class="metric-card metric-card-gray">
    <div class="metric-icon">
      <Icon name="shopping-cart" />
    </div>
    <div class="metric-content">
      <div class="metric-label">Original Capitalized Cost</div>
      <div class="metric-value">{{ formatCurrency(scheduleData?.total_original_cost || 0) }}</div>
      <div class="metric-sublabel">OCC</div>
    </div>
  </div>

  <!-- GVB Card -->
  <div class="metric-card metric-card-purple">
    <div class="metric-icon">
      <Icon name="trending-up" />
    </div>
    <div class="metric-content">
      <div class="metric-label">Gross Book Value</div>
      <div class="metric-value">{{ formatCurrency(scheduleData?.total_gross_book_value || 0) }}</div>
      <div class="metric-sublabel">GVB</div>
    </div>
  </div>

  <!-- NBV Card (existing) -->
  <div class="metric-card metric-card-blue">
    <div class="metric-icon">
      <Icon name="book" />
    </div>
    <div class="metric-content">
      <div class="metric-label">Net Book Value</div>
      <div class="metric-value">{{ formatCurrency(scheduleData?.total_net_book_value || 0) }}</div>
      <div class="metric-sublabel">NBV</div>
    </div>
  </div>

  <!-- AD Card -->
  <div class="metric-card metric-card-orange">
    <div class="metric-icon">
      <Icon name="chart-line" />
    </div>
    <div class="metric-content">
      <div class="metric-label">Accumulated Depreciation</div>
      <div class="metric-value">{{ formatCurrency(scheduleData?.total_accumulated_depreciation || 0) }}</div>
      <div class="metric-sublabel">AD</div>
      <div class="metric-badge">{{ scheduleData?.summary?.depreciation_percentage || 0 }}%</div>
    </div>
  </div>

  <!-- SV Card (existing, repositioned) -->
  <div class="metric-card metric-card-red">
    <div class="metric-icon">
      <Icon name="anchor" />
    </div>
    <div class="metric-content">
      <div class="metric-label">Salvage Value</div>
      <div class="metric-value">{{ formatCurrency(scheduleData?.total_salvage_value || 0) }}</div>
      <div class="metric-sublabel">SV</div>
    </div>
  </div>
</div>
```

#### 3.2 Card Styling

**File**: `web/src/views/amortization/ReportsView.vue` (CSS)

```css
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 1rem;
  margin-bottom: 2rem;
}

.metric-card {
  background: white;
  padding: 1.25rem;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  gap: 1rem;
  position: relative;
  overflow: hidden;
}

.metric-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
}

.metric-card-gray::before { background: #6b7280; }
.metric-card-purple::before { background: #8b5cf6; }
.metric-card-blue::before { background: #3b82f6; }
.metric-card-orange::before { background: #f97316; }
.metric-card-red::before { background: #ef4444; }

.metric-icon {
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
}

.metric-card-gray .metric-icon { background: #f3f4f6; color: #6b7280; }
.metric-card-purple .metric-icon { background: #ede9fe; color: #8b5cf6; }
.metric-card-blue .metric-icon { background: #dbeafe; color: #3b82f6; }
.metric-card-orange .metric-icon { background: #ffedd5; color: #f97316; }
.metric-card-red .metric-icon { background: #fee2e2; color: #ef4444; }

.metric-content {
  flex: 1;
}

.metric-label {
  font-size: 0.75rem;
  font-weight: 500;
  color: #6b7280;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.25rem;
}

.metric-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: #1f2937;
  line-height: 1.2;
}

.metric-sublabel {
  font-size: 0.7rem;
  font-weight: 600;
  color: #9ca3af;
  margin-top: 0.25rem;
}

.metric-badge {
  position: absolute;
  top: 0.75rem;
  right: 0.75rem;
  padding: 0.25rem 0.5rem;
  background: #ffedd5;
  color: #c2410c;
  font-size: 0.7rem;
  font-weight: 600;
  border-radius: 9999px;
}
```

### Phase 4: Frontend UI - Chart Reference Lines

#### 4.1 Chart Drawing Updates

**File**: `web/src/views/amortization/ReportsView.vue`

```typescript
const drawSimpleChart = () => {
  // ... existing setup ...

  const occ = scheduleData.value?.total_original_cost || 0
  const gvb = scheduleData.value?.total_gross_book_value || 0
  const sv = scheduleData.value?.total_salvage_value || 0

  // Draw OCC reference line (gray dashed)
  if (occ > 0 && occ < maxValue) {
    const occY = padding.top + chartHeight - (occ / maxValue) * chartHeight

    const occLine = document.createElementNS(ns, 'line')
    occLine.setAttribute('x1', padding.left.toString())
    occLine.setAttribute('y1', occY.toString())
    occLine.setAttribute('x2', (width - padding.right).toString())
    occLine.setAttribute('y2', occY.toString())
    occLine.setAttribute('stroke', '#6b7280')
    occLine.setAttribute('stroke-width', '2')
    occLine.setAttribute('stroke-dasharray', '12,4,4,4') // Long dash pattern
    svg.appendChild(occLine)

    const occLabel = document.createElementNS(ns, 'text')
    occLabel.setAttribute('x', (width - padding.right - 10).toString())
    occLabel.setAttribute('y', (occY - 8).toString())
    occLabel.setAttribute('text-anchor', 'end')
    occLabel.setAttribute('font-size', '10')
    occLabel.setAttribute('font-weight', '600')
    occLabel.setAttribute('fill', '#6b7280')
    occLabel.textContent = `OCC: ${formatCurrency(occ)}`
    svg.appendChild(occLabel)
  }

  // Draw GVB reference line (purple dashed)
  if (gvb > 0 && gvb < maxValue) {
    const gvbY = padding.top + chartHeight - (gvb / maxValue) * chartHeight

    const gvbLine = document.createElementNS(ns, 'line')
    gvbLine.setAttribute('x1', padding.left.toString())
    gvbLine.setAttribute('y1', gvbY.toString())
    gvbLine.setAttribute('x2', (width - padding.right).toString())
    gvbLine.setAttribute('y2', gvbY.toString())
    gvbLine.setAttribute('stroke', '#8b5cf6')
    gvbLine.setAttribute('stroke-width', '2')
    gvbLine.setAttribute('stroke-dasharray', '8,4')
    svg.appendChild(gvbLine)

    const gvbLabel = document.createElementNS(ns, 'text')
    gvbLabel.setAttribute('x', (width - padding.right - 10).toString())
    gvbLabel.setAttribute('y', (gvb - 8).toString())
    gvbLabel.setAttribute('text-anchor', 'end')
    gvbLabel.setAttribute('font-size', '10')
    gvbLabel.setAttribute('font-weight', '600')
    gvbLabel.setAttribute('fill', '#8b5cf6')
    gvbLabel.textContent = `GVB: ${formatCurrency(gvb)}`
    svg.appendChild(gvbLabel)
  }

  // ... existing SV line (red) ...
  // ... existing NBV line (blue) ...
  // ... existing AD line (orange) ...
}
```

#### 4.2 Updated Legend

```typescript
// Add to legend section (all in one row)
const legendY = height - 25
const legendX = padding.left

// OCC (Gray dashed)
const legendOCCLine = document.createElementNS(ns, 'line')
legendOCCLine.setAttribute('x1', legendX.toString())
legendOCCLine.setAttribute('y1', legendY.toString())
legendOCCLine.setAttribute('x2', (legendX + 30).toString())
legendOCCLine.setAttribute('y2', legendY.toString())
legendOCCLine.setAttribute('stroke', '#6b7280')
legendOCCLine.setAttribute('stroke-width', '2')
legendOCCLine.setAttribute('stroke-dasharray', '12,4,4,4')
svg.appendChild(legendOCCLine)

const legendTextOCC = document.createElementNS(ns, 'text')
legendTextOCC.setAttribute('x', (legendX + 38).toString())
legendTextOCC.setAttribute('y', (legendY + 4).toString())
legendTextOCC.setAttribute('font-size', '11')
legendTextOCC.setAttribute('fill', '#6b7280')
legendTextOCC.textContent = 'OCC'
svg.appendChild(legendTextOCC)

// GVB (Purple dashed)
const legendGVBLine = document.createElementNS(ns, 'line')
legendGVBLine.setAttribute('x1', (legendX + 80).toString())
legendGVBLine.setAttribute('y1', legendY.toString())
legendGVBLine.setAttribute('x2', (legendX + 110).toString())
legendGVBLine.setAttribute('y2', legendY.toString())
legendGVBLine.setAttribute('stroke', '#8b5cf6')
legendGVBLine.setAttribute('stroke-width', '2')
legendGVBLine.setAttribute('stroke-dasharray', '8,4')
svg.appendChild(legendGVBLine)

const legendTextGVB = document.createElementNS(ns, 'text')
legendTextGVB.setAttribute('x', (legendX + 118).toString())
legendTextGVB.setAttribute('y', (legendY + 4).toString())
legendTextGVB.setAttribute('font-size', '11')
legendTextGVB.setAttribute('fill', '#6b7280')
legendTextGVB.textContent = 'GVB'
svg.appendChild(legendTextGVB)

// ... existing legend items ...
```

### Phase 5: Monthly Data Table Enhancement

#### 5.1 Table Columns

**File**: `web/src/views/amortization/ReportsView.vue`

```vue
<table class="data-table">
  <thead>
    <tr>
      <th>Month</th>
      <th>Type</th>
      <th>OCC</th>
      <th>GVB</th>
      <th>Adjustments</th>
      <th>Depreciation</th>
      <th>AD</th>
      <th>NBV</th>
      <th>Assets</th>
    </tr>
  </thead>
  <tbody>
    <tr v-for="entry in scheduleData?.monthly_data || []" :key="entry.month">
      <td>{{ formatMonth(entry.month) }}</td>
      <td>
        <span :class="entry.is_projected ? 'badge badge-blue' : 'badge badge-gray'">
          {{ entry.is_projected ? 'Projected' : 'Actual' }}
        </span>
      </td>
      <td class="amount text-gray">{{ formatCurrency(entry.original_cost || 0) }}</td>
      <td class="amount text-purple">{{ formatCurrency(entry.gross_book_value) }}</td>
      <td class="amount" :class="entry.adjustment_amount > 0 ? 'text-green' : entry.adjustment_amount < 0 ? 'text-red' : ''">
        {{ formatCurrency(entry.adjustment_amount) }}
      </td>
      <td class="amount">{{ formatCurrency(entry.depreciation_amount) }}</td>
      <td class="amount text-orange">{{ formatCurrency(entry.accumulated_depreciation) }}</td>
      <td class="amount text-blue">{{ formatCurrency(entry.closing_book_value) }}</td>
      <td>{{ entry.active_assets_count }}</td>
    </tr>
  </tbody>
</table>
```

---

## Color Scheme

| Metric | Color | Hex | Purpose |
|--------|-------|-----|---------|
| OCC | Gray | #6b7280 | Baseline, neutral |
| GVB | Purple | #8b5cf6 | Capitalized investment |
| NBV | Blue | #3b82f6 | Current value |
| AD | Orange | #f97316 | Consumed value |
| SV | Red | #ef4444 | Residual value |
| Positive Adjustment | Green | #10b981 | Enhancement |
| Negative Adjustment | Red | #ef4444 | Correction/Write-down |

---

## API Changes

### New Response Fields

```json
{
  "currency": "USD",
  "total_original_cost": 170000.00,
  "total_gross_book_value": 175000.00,
  "total_net_book_value": 95000.00,
  "total_salvage_value": 5000.00,
  "total_accumulated_depreciation": 80000.00,
  "summary": {
    "depreciation_percentage": 45.71,
    "remaining_percentage": 54.29
  }
}
```

---

## Database Considerations

### No Schema Changes Required

The existing `amortization_ledger` table already has:
- `entry_type` ('depreciation', 'adjustment', 'write_off')
- `amount` (can be positive or negative for adjustments)
- `book_value_before` and `book_value_after`

We can calculate GVB by:
1. Summing all `purchase_cost` from `configuration_items` → OCC
2. Adding all adjustment amounts where `entry_type = 'adjustment'` → Net adjustments
3. OCC + Net Adjustments = GVB

---

## Testing Checklist

### Backend Tests
- [ ] OCC calculation matches sum of purchase costs
- [ ] GVB correctly adds positive adjustments
- [ ] GVB correctly subtracts negative adjustments
- [ ] Monthly GVB tracks adjustments over time
- [ ] AD percentage calculation correct
- [ ] Remaining percentage calculation correct

### Frontend Tests
- [ ] All 5 metric cards display correctly
- [ ] OCC reference line appears in chart
- [ ] GVB reference line appears in chart
- [ ] Legend shows all 5 metrics
- [ ] Table shows all new columns
- [ ] Colors match specification
- [ ] Tooltips show correct data

### Integration Tests
- [ ] API returns all new fields
- [ ] Frontend correctly parses response
- [ ] Chart renders with 3 reference lines
- [ ] Metrics are consistent across cards, chart, and table

---

## Migration Strategy

### No Breaking Changes
- All existing fields remain in the API
- New fields are additive only
- Frontend gracefully handles missing values (defaults to 0)

### Rollout Plan
1. **Phase 1**: Backend implementation (no UI changes yet)
2. **Phase 2**: Frontend type definitions
3. **Phase 3**: Summary cards (expand to 5 cards)
4. **Phase 4**: Chart reference lines (OCC and GVB)
5. **Phase 5**: Table enhancements (new columns)

---

## Success Criteria

✅ Users can see the complete asset value hierarchy
✅ Standard accounting terminology (OCC, GVB, NBV, AD, SV) is used throughout
✅ Visual reference lines in chart show the "value envelope"
✅ Summary cards provide quick metrics overview
✅ Table shows detailed monthly breakdown
✅ All metrics are mathematically consistent:
   - GVB = OCC + Net Adjustments
   - NBV = GVB - AD
   - Depreciation % = AD / GVB × 100
   - Remaining % = NBV / GVB × 100

---

## Future Enhancements

1. **Adjustment Categorization**
   - Add adjustment_type field to ledger
   - Distinguish between "enhancement" and "correction"
   - Show breakdown in UI

2. **Historical GVB Tracking**
   - Track GVB changes over time
   - Show when adjustments were made
   - Visual indicators for adjustment events

3. **Asset-Level Metrics**
   - Show OCC/GVB/NBV per asset in detail view
   - Compare assets by depreciation percentage

4. **Export Enhancement**
   - Include OCC/GVB in CSV exports
   - Add calculated columns to reports

---

## References

- **GAAP**: ASC 360 - Property, Plant and Equipment
- **IFRS**: IAS 16 - Property, Plant and Equipment
- **FASB**: Statement of Financial Accounting Standards No. 144

---

## Notes

- GVB is also known as "Carrying Amount" before depreciation
- NBV is also known as "Written Down Value" or "Carrying Amount"
- In some contexts, "Book Value" alone refers to NBV (after depreciation)
- Adjustments that increase asset value are "Capital Expenditures" (CapEx)
- Adjustments that decrease asset value are "Impairments" or "Write-downs"

