---
description: "Design system and UI conventions for the SyncTime admin dashboard. Follow these patterns for consistent styling, component usage, and responsive design across all features."
applyTo: "frontend/apps/admin/**/*.{tsx,ts,css}"
---

# SyncTime Admin Dashboard - Design System & Style Guide

## Overview

This design system defines the visual language, component patterns, and styling conventions for the SyncTime real-time session coordination admin dashboard. All UI development must follow these guidelines to ensure consistency.

## Core Technologies

- **Framework**: Next.js 14.2+ (App Router)
- **Styling**: Tailwind CSS 3.4+
- **Component Library**: shadcn/ui (Radix UI primitives)
- **Icons**: lucide-react
- **State**: React hooks + Zustand for global state

## Color Palette

### Primary Colors

```
Neutral Base:
- Background: bg-slate-50 (page backgrounds)
- Cards: bg-white
- Text Primary: text-slate-900
- Text Secondary: text-slate-600
- Text Muted: text-slate-500
- Borders: border-slate-200

Action/Primary:
- Primary Action: text-blue-600, bg-blue-600
- Primary Hover: hover:text-slate-900
- Focus: focus:border-slate-500, focus:ring-2, focus:ring-slate-200
```

### Status Colors

```
LIVE/Success:
- bg-emerald-50, text-emerald-700, border-emerald-200
- bg-emerald-100 (badges)

PAUSED/Warning:
- bg-amber-50, text-amber-700, border-amber-200
- bg-amber-100 (badges)

ENDED/Neutral:
- bg-slate-50, text-slate-600, border-slate-200
- bg-slate-100 (badges)

Error/Destructive:
- bg-red-50, text-red-700, border-red-200
- text-red-600 (error messages)

Info:
- bg-blue-50, text-blue-700, border-blue-200
- bg-blue-100 (badges)
```

### Special Accents

```
Purple (Agenda/Progress):
- bg-purple-50, text-purple-700

Dark Theme Elements:
- bg-slate-900 (special cards like AttendeeStats)
- text-white (on dark backgrounds)
```

## Button Design

### Core Rule: ALWAYS use `rounded-full` for ALL buttons

```tsx
// ✅ CORRECT - All buttons must be rounded-full
<Button className="rounded-full">Action</Button>
<Button variant="outline" className="rounded-full">Cancel</Button>
<Button variant="ghost" size="icon" className="rounded-full">
  <Icon className="w-5 h-5" />
</Button>

// ❌ WRONG - Never use default rounded
<Button>Action</Button>
<Button className="rounded-lg">Action</Button>
```

### Button Variants

```tsx
// Primary action buttons
<Button className="rounded-full">Create Session</Button>

// Secondary/Cancel buttons
<Button variant="outline" className="rounded-full">Cancel</Button>

// Icon-only buttons (ghost variant)
<Button variant="ghost" size="icon" className="rounded-full hover:bg-slate-100">
  <Bell className="w-5 h-5 text-slate-600" />
</Button>

// Destructive actions
<Button variant="destructive" className="rounded-full">End Session</Button>
```

### Button Heights

```
Default: h-10 sm:h-11 (responsive)
Icon buttons: size="icon" (automatic sizing)
Large action buttons: h-11 md:h-12
```

## Typography

### Font Stack

```css
font-family:
  ui-sans-serif,
  system-ui,
  -apple-system,
  Segoe UI,
  sans-serif;
```

### Text Sizes (Responsive)

```
Headlines: text-2xl font-bold text-slate-900
Section Titles: text-xl md:text-2xl font-bold
Body Text: text-sm (default)
Labels: text-xs font-semibold text-slate-500 uppercase tracking-wider
Large Numbers: text-3xl md:text-4xl font-bold
Timer Display: text-5xl sm:text-7xl md:text-8xl lg:text-9xl font-bold
```

### Font Weights

```
Regular: font-normal (default)
Medium: font-medium (labels)
Semibold: font-semibold (nav links, small headings)
Bold: font-bold (titles, headings, numbers)
```

## Spacing & Layout

### Container Width

```tsx
// All pages should use this max-width container
<div className="max-w-[1600px] mx-auto p-4 sm:p-6 lg:p-8">
```

### Responsive Padding

```
Mobile: p-4, px-4, py-3
Tablet: sm:p-6, sm:px-6, sm:py-4
Desktop: lg:p-8, lg:px-8
```

### Gap Values

```
Tight: gap-1 sm:gap-2
Normal: gap-3, gap-4
Loose: gap-6, gap-8
```

### Card Padding

```
Small cards: p-4 sm:p-6
Large cards: p-6 md:p-8
Compact content: p-3
```

## Responsive Design

### Mobile-First Approach

Always start with mobile layout, then add responsive classes:

```tsx
// Grid layouts - mobile to desktop
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">

// Column spans in bento grids
<Card className="col-span-12 md:col-span-8 lg:col-span-6">

// Text sizing
<h1 className="text-2xl md:text-3xl lg:text-4xl">

// Padding
<div className="p-4 sm:p-6 lg:p-8">

// Visibility
<div className="hidden md:flex">
<div className="flex md:hidden">
```

### Breakpoints

```
sm: 640px   (tablets)
md: 768px   (small laptops)
lg: 1024px  (desktops)
xl: 1280px  (large screens)
2xl: 1536px (extra large)
```

## Component Patterns

### Cards

```tsx
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";

<Card className="border-slate-200">
  <CardHeader className="pb-3">
    <CardTitle className="text-2xl font-bold text-slate-900">
      Title Here
    </CardTitle>
  </CardHeader>
  <CardContent>{/* Content */}</CardContent>
</Card>;
```

### Badges (Status Indicators)

```tsx
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

// Status-based badges
<Badge className={cn(
  "text-xs font-semibold border",
  status === "LIVE" && "bg-emerald-100 text-emerald-700 border-emerald-200",
  status === "PAUSED" && "bg-amber-100 text-amber-700 border-amber-200",
  status === "ENDED" && "bg-slate-100 text-slate-600 border-slate-200"
)}>
  {status}
</Badge>

// With pulse animation for live status
<Badge className="bg-emerald-50 text-emerald-700 border-emerald-200">
  <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse mr-2"></span>
  LIVE
</Badge>
```

### Inputs & Forms

```tsx
// Text inputs
<input
  value={value}
  onChange={(e) => setValue(e.target.value)}
  className="w-full rounded-md border border-slate-300 px-3 py-2 text-slate-900 outline-none ring-0 focus:border-slate-500 focus:ring-2 focus:ring-slate-200"
  placeholder="Enter value..."
/>

// Labels
<label className="block">
  <span className="mb-1.5 block text-sm font-medium text-slate-700">
    Label Text
  </span>
  <input {...props} />
</label>

// Error messages
<p className="text-sm text-red-600">{error}</p>
```

### Navigation Links

```tsx
import { cn } from "@/lib/utils";
import { usePathname } from "next/navigation";

const pathname = usePathname();
const isActive = pathname === href;

<Link
  href={href}
  className={cn(
    "text-sm font-semibold transition-colors border-b-2 pb-1",
    isActive
      ? "text-blue-600 border-blue-600"
      : "text-slate-600 border-transparent hover:text-slate-900",
  )}
>
  {label}
</Link>;
```

### Loading States

```tsx
// Spinner with clock icon
<Clock className="w-12 h-12 mx-auto mb-4 animate-spin text-slate-300" />

// Loading text
<p className="text-center text-slate-500">Loading...</p>
```

### Empty States

```tsx
<div className="text-center py-12 text-slate-500">
  <IconComponent className="w-12 h-12 mx-auto mb-4 text-slate-300" />
  <p className="mb-4">No items found</p>
  <Button variant="outline" className="rounded-full">
    Create First Item
  </Button>
</div>
```

## Bento Grid Layout

Use CSS Grid with responsive column spans:

```tsx
<div className="grid grid-cols-12 gap-4">
  {/* Large widget - 8 columns on medium+ */}
  <Card className="col-span-12 md:col-span-8">{/* Timer Widget */}</Card>

  {/* Sidebar widget - 4 columns on medium+ */}
  <Card className="col-span-12 md:col-span-4">{/* Stats Widget */}</Card>

  {/* Half-width widgets */}
  <Card className="col-span-12 md:col-span-6">{/* Agenda */}</Card>
  <Card className="col-span-12 md:col-span-6">{/* Another Widget */}</Card>

  {/* Quarter-width widgets on large screens */}
  <Card className="col-span-6 md:col-span-3">{/* Status Card */}</Card>
</div>
```

## Accessibility

### ARIA Labels

```tsx
// Icon buttons must have aria-label
<Button
  variant="ghost"
  size="icon"
  aria-label="Pause session"
  className="rounded-full"
>
  <Pause className="w-5 h-5" />
</Button>
```

### Focus States

All interactive elements must have visible focus states:

```
focus:border-slate-500
focus:ring-2
focus:ring-slate-200
focus:outline-none
```

## Animations

### Subtle Transitions

```
transition-colors (for hover states)
transition-all (for complex changes)
duration-200 ease-out (default timing)
```

### Pulse for Live Status

```tsx
<span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
```

### Hover States

```
hover:bg-slate-50 (cards)
hover:bg-slate-100 (buttons)
hover:text-slate-900 (links)
```

## Code Organization

### Component Structure

```tsx
"use client"; // If using hooks or client-side logic

import { useState } from "react";
import Link from "next/link";
// UI components
import { Card, Button } from "@/components/ui/*";
// Icons
import { Icon } from "lucide-react";
// Utils
import { cn } from "@/lib/utils";

interface ComponentProps {
  // Props with types
}

export default function ComponentName({ prop }: ComponentProps) {
  // Hooks first
  const [state, setState] = useState();

  // Functions
  const handleAction = () => {};

  // Render
  return <Card className="...">{/* Content */}</Card>;
}
```

### File Naming

```
Components: PascalCase.tsx (TimerWidget.tsx)
Pages: lowercase.tsx (page.tsx, layout.tsx)
Utils: camelCase.ts (backend.ts, session.ts)
```

## Common Pitfalls to Avoid

1. ❌ **Never** forget `rounded-full` on buttons
2. ❌ **Never** use inline colors - always use Tailwind classes
3. ❌ **Never** use fixed widths - always responsive
4. ❌ **Never** use magic numbers - use Tailwind spacing scale
5. ❌ **Never** skip mobile breakpoint - always start mobile-first
6. ❌ **Never** use `className=` without considering hover/focus states
7. ❌ **Never** create custom UI components - use shadcn/ui first

## Checklist for New Features

- [ ] All buttons have `rounded-full`
- [ ] Proper status colors used (emerald/amber/red/slate)
- [ ] Mobile-first responsive design
- [ ] Consistent spacing (4, 6, 8 padding pattern)
- [ ] Text uses proper sizing (text-sm, text-xs for labels)
- [ ] Cards use border-slate-200
- [ ] Hover states defined
- [ ] Focus states for accessibility
- [ ] Loading states implemented
- [ ] Empty states designed
- [ ] Error messages in red-600
- [ ] Uses shadcn/ui components where applicable
- [ ] Icon buttons have aria-label

## Quick Reference

```tsx
// Page Container
<div className="max-w-[1600px] mx-auto p-4 sm:p-6 lg:p-8 space-y-6">

// Section Title
<h2 className="text-2xl font-bold text-slate-900">Title</h2>

// Label
<span className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
  LABEL
</span>

// Primary Button
<Button className="rounded-full">
  <Icon className="w-4 h-4 mr-2" />
  Action
</Button>

// Status Badge
<Badge className="bg-emerald-100 text-emerald-700 border-emerald-200">
  LIVE
</Badge>

// Card
<Card className="border-slate-200">
  <CardContent className="p-4 sm:p-6">
    {/* Content */}
  </CardContent>
</Card>
```

## Questions?

When in doubt:

1. Check existing components in `/components` or `/app/dashboard`
2. Use shadcn/ui documentation for component patterns
3. Follow mobile-first responsive approach
4. Default to slate colors for neutrals
5. Always use `rounded-full` for buttons!
