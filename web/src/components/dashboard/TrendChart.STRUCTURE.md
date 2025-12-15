# TrendChart Component Structure

## Visual Layout

```
┌─────────────────────────────────────────────────────────────┐
│                      TrendChart Component                    │
│                                                              │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                    SVG Canvas                         │  │
│  │                                                       │  │
│  │  ┌─ Y Axis                                          │  │
│  │  │                                                   │  │
│  │  │  ╱╲  Series 1 (Blue)                            │  │
│  │  │ ╱  ╲╱╲                                           │  │
│  │  │╱      ╲  ┌─ Gradient Fill                       │  │
│  │  │        ╲╱                                        │  │
│  │  │    ╱╲   Series 2 (Green)                        │  │
│  │  │   ╱  ╲╱                                          │  │
│  │  │  ╱                                               │  │
│  │  └──┬───┬───┬───┬───┬───┬── X Axis (Dates)        │  │
│  │     Jan Feb Mar Apr May Jun                        │  │
│  │                                                     │  │
│  │  ┌─ Hover Line (dashed, on mouseover)             │  │
│  │  │  ● ← Hover circles at data points              │  │
│  │  │  ●                                              │  │
│  │  │                                                 │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌────────────────────────────────────────────────────┐     │
│  │  Legend (for multi-series)                         │     │
│  │  ■ Series 1   ■ Series 2   ■ Series 3             │     │
│  └────────────────────────────────────────────────────┘     │
│                                                              │
│  ┌────────────────────┐                                     │
│  │ Tooltip (on hover) │                                     │
│  │ January 15, 2025   │                                     │
│  │ ● Series 1: 42     │                                     │
│  │ ● Series 2: 38     │                                     │
│  └────────────────────┘                                     │
└─────────────────────────────────────────────────────────────┘
```

## Component Tree

```
TrendChart.vue
│
├── <div class="trend-chart-container">                  // Main container
│   │
│   ├── <svg viewBox="...">                              // SVG canvas
│   │   │
│   │   ├── <defs>                                       // Gradient definitions
│   │   │   └── <linearGradient> (for each series)
│   │   │
│   │   └── <g> (chart group with margins)
│   │       │
│   │       ├── <g class="x-axis">                       // X-axis (dates)
│   │       ├── <g class="y-axis">                       // Y-axis (values)
│   │       ├── <g class="grid">                         // Grid lines
│   │       │
│   │       ├── <g class="lines-group">                  // D3-rendered content
│   │       │   └── (for each series)
│   │       │       ├── <path class="area">              // Gradient fill
│   │       │       └── <path class="line">              // Trend line
│   │       │
│   │       ├── <rect class="interaction-overlay">       // Mouse events
│   │       │
│   │       ├── <line v-if="hoveredDate">                // Hover vertical line
│   │       │
│   │       └── <g class="hover-circles">                // Hover circles
│   │           └── <circle> (for each series at hover point)
│   │
│   ├── <div class="legend" v-if="showLegend">           // Legend
│   │   └── <div class="legend-item"> (for each series)
│   │       ├── <div class="legend-color">
│   │       └── <span> (series name)
│   │
│   └── <div class="tooltip" v-if="hoveredDate">         // Tooltip
│       ├── <div> (formatted date)
│       └── <div> (for each series value)
```

## Data Flow

```
Props (data: ChartSeries[])
    ↓
parsedData (computed)                    // Parse date strings to Date objects
    ↓
allDates (computed)                      // Extract unique dates
    ↓
yDomain (computed)                       // Calculate min/max for Y-axis
    ↓
initScales()                             // Create D3 scales
    ↓
    ├── xScale (d3.scaleTime)            // Date scale for X-axis
    └── yScale (d3.scaleLinear)          // Linear scale for Y-axis
    ↓
renderChart()
    ↓
    ├── renderXAxis()                    // D3 axis with date formatting
    ├── renderYAxis()                    // D3 axis with tick values
    ├── renderGrid()                     // Horizontal grid lines
    └── renderLines()                    // Lines + gradient fills
        ↓
        ├── lineGenerator (d3.line)      // Generate line path
        └── areaGenerator (d3.area)      // Generate area path
```

## Interaction Flow

```
User hovers over chart
    ↓
handleMouseMove(event)
    ↓
    ├── Calculate mouse X position
    ├── Use xScale.invert() to get date
    └── Find closest data point
    ↓
Update reactive state:
    ├── hoveredDate (Date)
    ├── hoverLineX (number)
    ├── tooltipX (number)
    └── tooltipY (number)
    ↓
Computed properties react:
    ├── hoveredPoints                    // Points at hovered date
    └── tooltipData                      // Data for tooltip display
    ↓
Template re-renders:
    ├── Show hover line
    ├── Show hover circles
    └── Show tooltip
```

## File Organization

```
web/src/components/dashboard/
│
├── TrendChart.vue                       // Main component (16 KB)
├── TrendChart.example.md                // Usage examples (6.4 KB)
├── TrendChart.test.example.ts           // Test examples (2.5 KB)
└── TrendChart.STRUCTURE.md              // This file
```

## Key Methods

### Rendering Pipeline
1. `initScales()` - Initialize D3 scales based on data
2. `renderXAxis()` - Render X-axis with D3
3. `renderYAxis()` - Render Y-axis with D3
4. `renderGrid()` - Render grid lines
5. `renderLines()` - Render trend lines and gradient fills
6. `renderChart()` - Orchestrate all rendering

### Interaction Handlers
1. `handleMouseMove()` - Track mouse position and update hover state
2. `handleMouseLeave()` - Clear hover state
3. `formatTooltipDate()` - Format date for tooltip display

### Lifecycle Management
1. `onMounted()` - Setup ResizeObserver and initial render
2. `onUnmounted()` - Cleanup ResizeObserver
3. `watch(data)` - Re-render on data changes
4. `watch(height)` - Re-render on height changes
5. `handleResize()` - Handle container resize

## Props to Template Flow

```typescript
// Props
data: ChartSeries[]
height: number (default: 300)
showLegend: boolean (default: true)
showGrid: boolean (default: true)

// Template refs
containerRef → <div class="trend-chart-container">
svgRef → <svg>
chartRef → <g> (main chart group)
xAxisRef → <g class="x-axis">
yAxisRef → <g class="y-axis">
gridRef → <g class="grid">
linesRef → <g class="lines-group">
tooltipRef → <div class="tooltip">

// Computed for template
width → SVG viewBox width
innerWidth → Chart area width (minus margins)
innerHeight → Chart area height (minus margins)
ariaLabel → Accessibility label
parsedData → Data with parsed dates
hoveredPoints → Data points at hover position
tooltipData → Formatted data for tooltip
```

## SVG Coordinate System

```
(0, 0) ─────────────────────→ X
  │
  │  ┌────── margin.left ──────┐
  │  │                         │
  │  │  ┌──────────────────┐   │
  │  │  │                  │   │  ← innerHeight
  │  │  │  Chart Area      │   │
  │  │  │  (data space)    │   │
  │  │  │                  │   │
  │  │  └──────────────────┘   │
  │  │     ← innerWidth         │
  │  └─────────────────────────┘
  │
  ↓ Y
```

## Color Scheme

### Gradients
- Top: `series.color` at 30% opacity
- Bottom: `series.color` at 5% opacity
- Direction: Vertical (top to bottom)

### Lines
- Stroke: `series.color` (100% opacity)
- Width: 2px
- Curve: `d3.curveMonotoneX` (smooth)

### Axes and Grid
- Axis lines: `#D1D5DB` (gray-300)
- Axis text: `#6B7280` (gray-500)
- Grid lines: `#F3F4F6` (gray-100) at 70% opacity

### Hover Elements
- Hover line: `#9CA3AF` (gray-400), dashed
- Hover circles: `series.color` with 2px stroke
- Tooltip background: `#111827` (gray-900)
- Tooltip text: `#FFFFFF` (white)

## Performance Optimization Points

1. **Memoization**: Computed properties cache expensive calculations
2. **D3 Enter/Update/Exit**: Minimal DOM manipulation
3. **ResizeObserver**: Only re-render when size actually changes
4. **Deep Watch**: Only re-render when data deeply changes
5. **Lazy Rendering**: Chart only renders when mounted and visible
6. **Virtual DOM**: Vue 3 optimizes template updates

## Accessibility Tree

```
[img] TrendChart: Line chart showing trends for...
├── [graphics-document] SVG canvas
│   ├── [group] Chart group
│   │   ├── [group] X-axis
│   │   │   └── [text] Date labels
│   │   ├── [group] Y-axis
│   │   │   └── [text] Value labels
│   │   └── [group] Lines group
│   │       └── [path] Trend lines (for each series)
│   └── [graphics-symbol] Gradient definitions
├── [list] Legend (if multi-series)
│   └── [listitem] Series items
└── [tooltip] Hover tooltip (when active)
    └── [text] Date and values
```
