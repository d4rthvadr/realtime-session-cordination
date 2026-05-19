# Copilot Configuration

This directory contains design system documentation and instructions for GitHub Copilot to maintain consistency across the SyncTime admin dashboard.

## Files

### `design-system.md`

Complete design system documentation including:

- Color palette and usage
- Button styles (all must be `rounded-full`)
- Typography scale
- Responsive design patterns
- Component patterns (shadcn/ui)
- Code organization
- Accessibility guidelines
- Common pitfalls to avoid

**Use this as the authoritative reference for all UI development.**

### `instructions.md`

Quick reference and critical rules for Copilot:

- Essential design rules
- Common patterns
- Component templates
- Status color patterns

## Usage

When developing new features or modifying existing ones:

1. **Check design-system.md first** for comprehensive guidelines
2. Reference existing components for patterns
3. Follow mobile-first responsive approach
4. Use shadcn/ui components
5. Never forget `rounded-full` on buttons!

## Key Principles

1. **Consistency**: Use established patterns
2. **Accessibility**: Include ARIA labels, focus states
3. **Responsiveness**: Mobile-first with sm/md/lg breakpoints
4. **Simplicity**: Use shadcn/ui, don't reinvent components
5. **Performance**: Semantic HTML, proper React patterns

## Color Quick Reference

```
Neutrals: slate-50, slate-200, slate-500, slate-600, slate-900
Live/Success: emerald-50, emerald-100, emerald-700
Paused/Warning: amber-50, amber-100, amber-700
Error: red-50, red-600
Actions: blue-600
```

## Button Rule

**ALL buttons must include `rounded-full` class - no exceptions!**

```tsx
✅ <Button className="rounded-full">Action</Button>
❌ <Button>Action</Button>
```

## Repository Memory

Design system essentials are also saved in `/memories/repo/design-system.md` for quick access.
