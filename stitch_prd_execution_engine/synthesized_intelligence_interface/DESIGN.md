---
name: Synthesized Intelligence Interface
colors:
  surface: '#0b1326'
  surface-dim: '#0b1326'
  surface-bright: '#31394d'
  surface-container-lowest: '#060e20'
  surface-container-low: '#131b2e'
  surface-container: '#171f33'
  surface-container-high: '#222a3d'
  surface-container-highest: '#2d3449'
  on-surface: '#dae2fd'
  on-surface-variant: '#c7c4d8'
  inverse-surface: '#dae2fd'
  inverse-on-surface: '#283044'
  outline: '#918fa1'
  outline-variant: '#464555'
  surface-tint: '#c3c0ff'
  primary: '#c3c0ff'
  on-primary: '#1d00a5'
  primary-container: '#4f46e5'
  on-primary-container: '#dad7ff'
  inverse-primary: '#4d44e3'
  secondary: '#6bd8cb'
  on-secondary: '#003732'
  secondary-container: '#29a195'
  on-secondary-container: '#00302b'
  tertiary: '#bcc7de'
  on-tertiary: '#263143'
  tertiary-container: '#566175'
  on-tertiary-container: '#d1dcf4'
  error: '#ffb4ab'
  on-error: '#690005'
  error-container: '#93000a'
  on-error-container: '#ffdad6'
  primary-fixed: '#e2dfff'
  primary-fixed-dim: '#c3c0ff'
  on-primary-fixed: '#0f0069'
  on-primary-fixed-variant: '#3323cc'
  secondary-fixed: '#89f5e7'
  secondary-fixed-dim: '#6bd8cb'
  on-secondary-fixed: '#00201d'
  on-secondary-fixed-variant: '#005049'
  tertiary-fixed: '#d8e3fb'
  tertiary-fixed-dim: '#bcc7de'
  on-tertiary-fixed: '#111c2d'
  on-tertiary-fixed-variant: '#3c475a'
  background: '#0b1326'
  on-background: '#dae2fd'
  surface-variant: '#2d3449'
  background-principal: '#0F172A'
  surface-secondary: '#1E293B'
  text-primary: '#F8FAFC'
  text-secondary: '#94A3B8'
  accent-indigo: '#6366F1'
  accent-teal: '#14B8A6'
  status-success: '#22C55E'
  status-warning: '#F59E0B'
  status-error: '#EF4444'
  border-subtle: '#334155'
typography:
  headline-lg:
    fontFamily: Hanken Grotesk
    fontSize: 32px
    fontWeight: '600'
    lineHeight: '1.2'
    letterSpacing: -0.02em
  headline-md:
    fontFamily: Hanken Grotesk
    fontSize: 24px
    fontWeight: '600'
    lineHeight: '1.3'
  headline-sm:
    fontFamily: Hanken Grotesk
    fontSize: 18px
    fontWeight: '600'
    lineHeight: '1.4'
  body-lg:
    fontFamily: Hanken Grotesk
    fontSize: 16px
    fontWeight: '400'
    lineHeight: '1.6'
  body-md:
    fontFamily: Hanken Grotesk
    fontSize: 14px
    fontWeight: '400'
    lineHeight: '1.5'
  body-sm:
    fontFamily: Hanken Grotesk
    fontSize: 13px
    fontWeight: '400'
    lineHeight: '1.5'
  code-sm:
    fontFamily: JetBrains Mono
    fontSize: 12px
    fontWeight: '400'
    lineHeight: '1.6'
  label-caps:
    fontFamily: JetBrains Mono
    fontSize: 11px
    fontWeight: '600'
    lineHeight: '1'
    letterSpacing: 0.05em
rounded:
  sm: 0.125rem
  DEFAULT: 0.25rem
  md: 0.375rem
  lg: 0.5rem
  xl: 0.75rem
  full: 9999px
spacing:
  unit: 4px
  gutter: 16px
  margin-desktop: 24px
  sidebar-width: 260px
  context-panel-width: 320px
  max-content-width: 1200px
---

## Brand & Style

The design system is engineered for high-stakes enterprise environments where Retrieval-Augmented Generation (RAG) serves as a mission-critical tool. The brand personality is **Sober, Technical, and Authoritative**, eschewing the whimsical "magic" often associated with AI in favor of a "scientific instrument" aesthetic. It prioritizes the relationship between data, evidence, and synthesis.

The visual style is a refined **Corporate Modern** approach with **Minimalist** leanings. It utilizes a flat hierarchy that relies on surgical precision in spacing and subtle tonal shifts rather than dramatic shadows or decorative elements. The interface should feel like a high-end IDE or a research terminal—focused on throughput, clarity, and verifiability.

Key visual principles:
- **Utility over Ornament:** Every element must serve a functional purpose.
- **Data Density:** High information density is maintained through compact typography and efficient use of negative space.
- **Technical Integrity:** Monospaced accents and status-driven color logic emphasize the system's "under-the-hood" reliability.

## Colors

The color palette is anchored in a "Professional Dark" scheme. The foundation is **Deep Charcoal** (`#0F172A`), providing a low-strain environment for extended research sessions. 

- **Primary & Secondary:** A sophisticated pairing of Indigo and Blue-Green (Teal) is used for interactive states and primary actions. These are used sparingly to guide the eye without causing visual fatigue.
- **Functional Grays:** Text follows a strict hierarchy—Warm White for primary readability and Neutral Gray for metadata and secondary labels.
- **Status Logic:** Semantic colors (Green, Amber, Red) are desaturated to fit the sober aesthetic while remaining distinct for rapid diagnostic scanning.
- **Surface Strategy:** Depth is communicated through tonal stepping (e.g., using a slightly lighter charcoal for cards) rather than elevation.

## Typography

This design system uses **Hanken Grotesk** as its primary typeface, chosen for its sharp, contemporary geometry and exceptional legibility at small sizes. 

- **Primary Body:** Optimized for long-form consumption of AI-generated responses. Line heights are slightly generous to prevent "text-walling."
- **Technical Metrics:** **JetBrains Mono** is utilized for all technical data, including vector scores, timestamps, and log snippets. This differentiates "system data" from "generated content."
- **Information Density:** The scale is intentionally compact. Labels use uppercase monospaced styling to provide a structural feel to the interface sections.
- **Scale:** On mobile devices, `headline-lg` should scale down to `24px` to maintain visual balance within tight margins.

## Layout & Spacing

The layout employs a **Fluid-Fixed Hybrid Grid**. 
- **Sidebar:** A fixed-width left navigation for persistence.
- **Workspace:** A flexible center column that expands and contracts, but maintains a maximum readable width of 1200px for the chat container to prevent line-length issues.
- **Contextual Panel:** A right-aligned panel dedicated to sources and metadata, collapsible to maximize workspace when evidence review is not the primary task.

Spacing follows a strict 4px base unit. 
- **Comfortable-Professional:** 24px margins between major containers.
- **Compact (Diagnostic):** 8px padding within table rows and metadata lists to maximize data visibility.
- **Breakpoints:** On tablets, the context panel transitions to a bottom-sheet or hidden drawer. On mobile, the interface collapses into a single-column stack with a focus on the input and response flow.

## Elevation & Depth

This system avoids traditional material elevation (large shadows) to maintain its sober, technical character. Depth is communicated via:

1.  **Tonal Layering:** The primary background is the darkest layer. Secondary containers (sidebars, cards) use lighter charcoal values to "float" above the base.
2.  **Subtle Outlines:** Components are defined by 1px solid borders in `border-subtle` (`#334155`). This provides clear structural definition without the visual "weight" of shadows.
3.  **Active Focus:** When an element is focused or active (like a selected source), a 1px border of `primary_color` or `secondary_color` is used instead of a glow or shadow.
4.  **Zero-Shadow Philosophy:** Shadows are reserved exclusively for temporary floating elements like tooltips or context menus, where they are rendered with 0% offset and a tight, low-opacity blur.

## Shapes

The shape language is **Soft (0.25rem)**. This "discrete" roundedness maintains a professional, architectural feel while slightly softening the technical edges of the dark-mode UI.

- **Inputs and Buttons:** Use `rounded` (4px).
- **Cards and Modals:** Use `rounded-lg` (8px).
- **Status Indicators:** Use `rounded-xl` (12px) or full pills for badges to distinguish them from structural UI components.
- **Dividers:** 1px solid lines are the primary tool for separating content blocks within a single surface.

## Components

- **Buttons:** Primary buttons use a solid `primary_color` or `secondary_color` fill with white text. Secondary buttons are "ghost" style with subtle borders. All transitions should be instantaneous or use a very fast 150ms linear ease.
- **Source Cards:** A critical component for RAG. These should display a monospaced "relevance score," a title, and a snippet. They use `surface-secondary` backgrounds with a hover state that changes the border color to the indigo accent.
- **Streaming Cursor:** A vertical bar at the end of generated text, pulsing slightly in the secondary accent color to indicate active processing.
- **Input Field:** A large, multi-line text area for queries. It should lack a traditional border, instead sitting in a container with a slightly darker or lighter fill than the workspace to appear "embedded."
- **Status Badges:** Compact, using the monospaced label style. They indicate system health (e.g., "GPU: ACTIVE", "INDEX: READY") with a small colored dot prefix.
- **Skeletons:** Content loading is handled by subtle, low-contrast pulsing rectangles that match the text line-heights, ensuring no layout shift during the "Retrieving" phase.