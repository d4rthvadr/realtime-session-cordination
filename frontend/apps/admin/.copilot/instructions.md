---
description: "Follow SyncTime admin dashboard design system for all UI development"
---

# SyncTime Admin Dashboard Instructions

## Design System

Always follow the design system documented in `.copilot/design-system.md`

## Critical Rules

### 1. Buttons MUST be Rounded

**NEVER FORGET**: ALL buttons must have `rounded-full` class

```tsx
✅ <Button className="rounded-full">Action</Button>
❌ <Button>Action</Button>
```

### 2. Color Palette

- **Neutrals**: slate (backgrounds, text, borders)
- **Live/Success**: emerald-100/700
- **Paused/Warning**: amber-100/700
- **Error**: red-600/50
- **Actions**: blue-600

### 3. Responsive Design

Mobile-first approach:

```tsx
// Grid columns
col-span-12 md:col-span-8 lg:col-span-6

// Text sizes
text-sm md:text-base lg:text-lg

// Padding
p-4 sm:p-6 lg:p-8
```

### 4. Component Library

Use shadcn/ui components - never build custom UI from scratch:

- Card, CardHeader, CardTitle, CardContent
- Button (with variants: default, outline, ghost, destructive)
- Badge (with custom classes for status colors)
- Progress, Dialog, etc.

### 5. Status Color Pattern

```tsx
const getStatusColor = (status: string) => {
  switch (status) {
    case "LIVE":
      return "bg-emerald-100 text-emerald-700 border-emerald-200";
    case "PAUSED":
      return "bg-amber-100 text-amber-700 border-amber-200";
    case "ENDED":
      return "bg-slate-100 text-slate-600 border-slate-200";
    default:
      return "bg-blue-100 text-blue-700 border-blue-200";
  }
};
```

### 6. Container Pattern

```tsx
<div className="max-w-[1600px] mx-auto p-4 sm:p-6 lg:p-8">
  {/* All page content */}
</div>
```

### 7. Typography Scale

- Labels: `text-xs font-semibold text-slate-500 uppercase tracking-wider`
- Body: `text-sm text-slate-600`
- Headings: `text-2xl font-bold text-slate-900`
- Large numbers: `text-3xl md:text-4xl font-bold`

## Development Workflow

When adding new features:

1. Check `.copilot/design-system.md` for patterns
2. Use existing components as reference
3. Start mobile-first (col-span-12)
4. Add responsive breakpoints (md:, lg:)
5. Use shadcn/ui components
6. Add `rounded-full` to ALL buttons
7. Test on mobile, tablet, desktop

## Common Components Reference

```tsx
// Page Layout
export default function Page() {
  return (
    <div className="max-w-[1600px] mx-auto p-4 sm:p-6 lg:p-8 space-y-6">
      <Card className="border-slate-200">
        <CardHeader className="pb-3">
          <CardTitle className="text-2xl font-bold text-slate-900">
            Title
          </CardTitle>
        </CardHeader>
        <CardContent>
          <Button className="rounded-full">Action</Button>
        </CardContent>
      </Card>
    </div>
  );
}
```

## Questions?

Refer to `.copilot/design-system.md` for complete guidelines.
