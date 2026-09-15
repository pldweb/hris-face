# Graph Report - .  (2026-09-15)

## Corpus Check
- 28 files · ~110,750 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1604 nodes · 2227 edges · 140 communities (116 shown, 24 thin omitted)
- Extraction: 99% EXTRACTED · 1% INFERRED · 0% AMBIGUOUS · INFERRED: 18 edges (avg confidence: 0.67)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- Face Recognition Service
- Web Package Dependencies
- MCP Server Tools
- Antislop Copywriting Rules (agent)
- Antislop Copywriting Rules (.agents)
- Attendance Service Core
- Admin Monitoring Page
- Admin Master Data Page
- Antislop Mobile Layout (agent)
- Antislop Mobile Layout (.agents)
- Admin Employee List
- Employee Service
- Check-in Page
- Employee HTTP Handlers
- Antislop Accessibility Rules (agent)
- Antislop Accessibility Rules (.agents)
- Leave Service
- Report Service
- Admin Leave Review
- Face Enrollment Page
- Report HTTP Handlers
- TS Node Config
- Attendance HTTP Handlers
- Admin Face Scan
- Employee History Page
- Employee Leave Page
- Antislop Hard Gate (agent)
- Antislop Hard Gate (.agents)
- Correction Service
- Antislop Code Comments (agent)
- Antislop Code Comments (.agents)
- Correction HTTP Handlers
- Auth Token Service
- Leave HTTP Handlers
- TS App Config
- Employee API Client
- Profile Edit Modal
- Master Data Handlers
- Master Data Service
- Antislop Purpose Gate (agent)
- Antislop Purpose Gate (.agents)
- Auth HTTP Handlers
- Enrollment Service
- E2E API Test Helpers
- Corrections UI
- Master Data API Client
- Contrast Check Script (agent)
- Contrast Check Script (.agents)
- Attendance Photo Store
- PRD Overview Sections
- Inline Master Select
- Slop Layout Patterns (agent)
- Slop Visual Patterns (agent)
- Slop Layout Patterns (.agents)
- Slop Visual Patterns (.agents)
- Login Throttle
- Core DB Schema
- Employee Nav And Team
- Contrast MCP Server (agent)
- Antislop Quality Locks (agent)
- Slop Decorative Patterns (agent)
- Contrast MCP Server (.agents)
- Antislop Quality Locks (.agents)
- Slop Decorative Patterns (.agents)
- MCP Documentation
- Face Service Docs
- MCP E2E Test
- Web App Bootstrap
- Admin Filters And Review
- API Config Loading
- PRD Frontend Direction
- PRD Functional Requirements
- Web Auth Client
- Master Data Loaders
- Admin Layout Shell
- Antislop Core Skill (agent)
- Slop Pattern Catalog (agent)
- Slop Dashboard Patterns (agent)
- Antislop Core Skill (.agents)
- Slop Pattern Catalog (.agents)
- Slop Dashboard Patterns (.agents)
- Local Run Script
- All Pages Browser Test
- Antislop UI Motion (agent)
- Antislop UI Motion (.agents)
- PRD Deployment Plan
- Product Brief
- Login Page
- Craftsmanship Standard (agent)
- Craftsmanship Standard (.agents)
- Enrollment Handlers
- PRD Architecture
- Project README
- Admin Browser E2E
- Leave E2E Test
- Antislop Delivery Gate (agent)
- Antislop Delivery Gate (.agents)
- Face Service Client
- Anti-spoof Model Export
- E2E Test Docs
- Demo Data Seeder
- Admin UI Test
- HR Editing Test
- HR Editing UI Test
- Attendance API Client
- Liveliness Toolkit (agent)
- Slop Structural Patterns (agent)
- Liveliness Toolkit (.agents)
- Slop Structural Patterns (.agents)
- Rate Limit Middleware
- PRD Main Flows
- Browser E2E Test
- Admin CRUD Test
- Enrollment API Client
- Agent Instructions
- Leave Requests Migration
- Request Logger
- Root TS Config
- Refresh Tokens Migration
- Leave Migration Down
- Profile Migration Down
- Profile Migration Up
- Deploy Script
- Check-in Design Contract
- Web README
- Isolated Client Node
- Attendance Service Core
- Attendance Service Core
- Attendance Service Core
- Mutex Type
- Isolated Client Node
- Go Module

## God Nodes (most connected - your core abstractions)
1. `Service` - 19 edges
2. `Group 1: Hard Gate (absolute, no exceptions)` - 18 edges
3. `Group 1: Hard Gate (absolute, no exceptions)` - 18 edges
4. `RegisterRoutes()` - 17 edges
5. `PRD — HRIS Absensi Berbasis Face Recognition` - 16 edges
6. `Service` - 16 edges
7. `compilerOptions` - 15 edges
8. `tzDayjs()` - 15 edges
9. `Client` - 14 edges
10. `Group 2: Purpose-Gate (technique allowed, purpose required)` - 13 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `FaceEngine`  [EXTRACTED]
  test/e2e/benchmark_accuracy.py → face/app/face_engine.py
- `main()` --calls--> `FaceEngine`  [EXTRACTED]
  test/e2e/benchmark_liveness.py → face/app/face_engine.py
- `main()` --calls--> `LivenessDetector`  [EXTRACTED]
  test/e2e/benchmark_liveness.py → face/app/liveness.py
- `analyze()` --references--> `post()`  [EXTRACTED]
  face/app/main.py → test/e2e/test_challenge.py
- `main()` --calls--> `New()`  [INFERRED]
  api/cmd/api/main.go → api/internal/mcpserver/tools.go

## Import Cycles
- None detected.

## Communities (140 total, 24 thin omitted)

### Community 0 - "Face Recognition Service"
Cohesion: 0.05
Nodes (40): eye_openness(), face_signature(), ndarray, Eye openness, used for the blink challenge (docs/PRD.md F5).  Passive anti-spoof, Mean vertical eye opening as a fraction of face width.      Normalising by face, An 8x8 grayscale thumbnail of the face crop.      The blink challenge needs to p, DetectedFace, FaceEngine (+32 more)

### Community 1 - "Web Package Dependencies"
Cohesion: 0.05
Nodes (40): @ant-design/icons-vue, ant-design-vue, axios, dayjs, @mediapipe/tasks-vision, pinia, @types/node, typescript (+32 more)

### Community 2 - "MCP Server Tools"
Cohesion: 0.10
Nodes (32): envOr(), main(), Context, Service, mapHeader(), NewClient(), departmentIDByName(), errResult() (+24 more)

### Community 3 - "Antislop Copywriting Rules (agent)"
Cohesion: 0.05
Nodes (38): Actorless Passive, All-Caps Emphasis, antislop-copywriting, Aphorism Formulas, Boldface Overuse, Chatbot Closers, Copywriting Skill Checklist, Draft, audit, final (+30 more)

### Community 4 - "Antislop Copywriting Rules (.agents)"
Cohesion: 0.05
Nodes (38): Actorless Passive, All-Caps Emphasis, antislop-copywriting, Aphorism Formulas, Boldface Overuse, Chatbot Closers, Copywriting Skill Checklist, Draft, audit, final (+30 more)

### Community 5 - "Attendance Service Core"
Cohesion: 0.14
Nodes (18): Service, Context, meanPairwiseVariation(), Service, Context, NewService(), randomKey(), ChallengeResult (+10 more)

### Community 6 - "Admin Monitoring Page"
Cohesion: 0.08
Nodes (31): AttendanceFilter, AttendanceRow, DailySummary, DayRecord, deleteAttendance(), exportReport(), fetchToday(), listAttendances() (+23 more)

### Community 7 - "Admin Master Data Page"
Cohesion: 0.07
Nodes (28): deleteDepartment(), approveDevice(), deleteLocation(), DAY_LABELS, deletingDepartmentId, deletingLocationId, departmentForm, departmentModal (+20 more)

### Community 8 - "Antislop Mobile Layout (agent)"
Cohesion: 0.07
Nodes (29): 100vh Sections, antislop-layoutmobile, Bottom Nav That Eats Content, Breakpoint Driven by Device List, Breakpoints, Columns That Don't Collapse, Desktop-Only Layout, Desktop-Sized Everything (+21 more)

### Community 9 - "Antislop Mobile Layout (.agents)"
Cohesion: 0.07
Nodes (29): 100vh Sections, antislop-layoutmobile, Bottom Nav That Eats Content, Breakpoint Driven by Device List, Breakpoints, Columns That Don't Collapse, Desktop-Only Layout, Desktop-Sized Everything (+21 more)

### Community 10 - "Admin Employee List"
Cohesion: 0.07
Nodes (27): deactivateEmployee(), importEmployeesCsv(), columns, deactivatingId, departmentOptions, departments, editingId, employees (+19 more)

### Community 11 - "Employee Service"
Cohesion: 0.14
Nodes (15): Context, Pool, Time, NewService(), randomPassword(), CreateEmployeeInput, CreateEmployeeResult, Department (+7 more)

### Community 12 - "Check-in Page"
Cohesion: 0.09
Nodes (27): submitAttendance(), submitChallenge(), fetchMe(), actionLabel, CameraError, canvasEl, captureBurst(), captureJpeg() (+19 more)

### Community 13 - "Employee HTTP Handlers"
Cohesion: 0.25
Nodes (22): approveDeviceHandler(), createDepartmentHandler(), createHandler(), deactivateHandler(), deleteDepartmentHandler(), HandlerFunc, IRoutes, Service (+14 more)

### Community 14 - "Antislop Accessibility Rules (agent)"
Cohesion: 0.10
Nodes (20): antislop-human, Broken Tab Order, Color & Contrast, Color-Only Feedback, Focus & States, How to use this skill, Human Skill Checklist, Keyboard (+12 more)

### Community 15 - "Antislop Accessibility Rules (.agents)"
Cohesion: 0.10
Nodes (20): antislop-human, Broken Tab Order, Color & Contrast, Color-Only Feedback, Focus & States, How to use this skill, Human Skill Checklist, Keyboard (+12 more)

### Community 16 - "Leave Service"
Cohesion: 0.22
Nodes (11): countWorkdays(), Context, Pool, Time, NewService(), scanLeaveRows(), Balance, LeaveRequest (+3 more)

### Community 17 - "Report Service"
Cohesion: 0.21
Nodes (11): Context, Location, Pool, Time, NewService(), DailySummary, DayRecord, Filter (+3 more)

### Community 18 - "Admin Leave Review"
Cohesion: 0.12
Nodes (20): decideLeaveRequest(), listAdminLeaveRequests(), updateLeaveRequest(), columns, confirmReview(), editForm, editing, editOpen (+12 more)

### Community 19 - "Face Enrollment Page"
Cohesion: 0.10
Nodes (17): cameraError, cameraReady, canCapture, canSubmit, canvasEl, captureJpeg(), captures, currentPoseLabel (+9 more)

### Community 20 - "Report HTTP Handlers"
Cohesion: 0.34
Nodes (19): boolLabel(), deleteHandler(), exportHandler(), exportXlsxHandler(), filterFrom(), Context, HandlerFunc, IRoutes (+11 more)

### Community 21 - "TS Node Config"
Cohesion: 0.10
Nodes (19): ES2023, node, vite.config.ts, compilerOptions, allowImportingTsExtensions, erasableSyntaxOnly, lib, module (+11 more)

### Community 22 - "Attendance HTTP Handlers"
Cohesion: 0.19
Nodes (15): main(), main(), mustEnv(), challengeHandler(), HandlerFunc, IRoutes, Service, markHandler() (+7 more)

### Community 23 - "Admin Face Scan"
Cohesion: 0.13
Nodes (15): apiClient, FaceScanResult, scanFace(), cameraError, canScan, canvasEl, captureJpeg(), onScan() (+7 more)

### Community 24 - "Employee History Page"
Cohesion: 0.12
Nodes (16): createCorrection(), fetchMyHistory(), Dayjs, DayjsArgs, corrections, days, form, load() (+8 more)

### Community 25 - "Employee Leave Page"
Cohesion: 0.16
Nodes (17): createLeaveRequest(), fetchMyLeaveBalance(), LeaveBalance, LeaveRequest, LeaveStatus, LeaveType, listMyLeaveRequests(), balance (+9 more)

### Community 26 - "Antislop Hard Gate (agent)"
Cohesion: 0.11
Nodes (18): Group 1: Hard Gate (absolute, no exceptions), R-02 — Copywriting, R-03 — Mobile Responsiveness, R-17 — Data & Numbers, R-18 — Testimonials, R-23 — Clarification & Visual Assets, R-24 — Navigation, R-25 — Color Contrast (+10 more)

### Community 27 - "Antislop Hard Gate (.agents)"
Cohesion: 0.11
Nodes (18): Group 1: Hard Gate (absolute, no exceptions), R-02 — Copywriting, R-03 — Mobile Responsiveness, R-17 — Data & Numbers, R-18 — Testimonials, R-23 — Clarification & Visual Assets, R-24 — Navigation, R-25 — Color Contrast (+10 more)

### Community 28 - "Correction Service"
Cohesion: 0.21
Nodes (9): Context, Pool, Time, NewService(), New(), Correction, Service, Config (+1 more)

### Community 29 - "Antislop Code Comments (agent)"
Cohesion: 0.12
Nodes (16): antislop-code, Code Comment Checklist, Comments That Add Nothing, Decorative Emoji, Decorative Separators, Empty Labels, End Markers, How It Should Read (+8 more)

### Community 30 - "Antislop Code Comments (.agents)"
Cohesion: 0.12
Nodes (16): antislop-code, Code Comment Checklist, Comments That Add Nothing, Decorative Emoji, Decorative Separators, Empty Labels, End Markers, How It Should Read (+8 more)

### Community 31 - "Correction HTTP Handlers"
Cohesion: 0.21
Nodes (15): createHandler(), HandlerFunc, IRoutes, Service, listHandler(), myListHandler(), RegisterRoutes(), reviewHandler() (+7 more)

### Community 32 - "Auth Token Service"
Cohesion: 0.25
Nodes (10): Context, Duration, Pool, RegisteredClaims, hashToken(), NewService(), randomToken(), Claims (+2 more)

### Community 33 - "Leave HTTP Handlers"
Cohesion: 0.36
Nodes (15): createHandler(), Context, HandlerFunc, IRoutes, Service, listHandler(), myBalanceHandler(), myListHandler() (+7 more)

### Community 34 - "TS App Config"
Cohesion: 0.12
Nodes (15): src/**/*.ts, src/**/*.tsx, src/**/*.vue, vite/client, @vue/tsconfig/tsconfig.dom.json, compilerOptions, allowArbitraryExtensions, erasableSyntaxOnly (+7 more)

### Community 35 - "Employee API Client"
Cohesion: 0.13
Nodes (15): createDepartment(), createEmployee(), CreateEmployeeInput, CreateEmployeeResult, Department, Employee, EmployeeStatus, ImportResult (+7 more)

### Community 36 - "Profile Edit Modal"
Cohesion: 0.15
Nodes (15): fetchProfile(), Profile, updateProfile(), auth, changingPassword, emit, form, load() (+7 more)

### Community 37 - "Master Data Handlers"
Cohesion: 0.35
Nodes (14): assignScheduleHandler(), createLocationHandler(), createScheduleHandler(), deleteLocationHandler(), HandlerFunc, IRoutes, Service, listLocationsHandler() (+6 more)

### Community 38 - "Master Data Service"
Cohesion: 0.27
Nodes (6): Context, Pool, NewService(), Location, Schedule, Service

### Community 39 - "Antislop Purpose Gate (agent)"
Cohesion: 0.15
Nodes (13): Group 2: Purpose-Gate (technique allowed, purpose required), R-01 — Color & Gradients, R-04 — Icons, R-06 — Typography, R-07 — Background, R-08 — Button Arrows, R-09 — Badges, R-10 — Glassmorphism (+5 more)

### Community 40 - "Antislop Purpose Gate (.agents)"
Cohesion: 0.15
Nodes (13): Group 2: Purpose-Gate (technique allowed, purpose required), R-01 — Color & Gradients, R-04 — Icons, R-06 — Typography, R-07 — Background, R-08 — Button Arrows, R-09 — Badges, R-10 — Glassmorphism (+5 more)

### Community 41 - "Auth HTTP Handlers"
Cohesion: 0.36
Nodes (12): clearRefreshCookie(), Context, Duration, HandlerFunc, IRoutes, Service, loginHandler(), logoutHandler() (+4 more)

### Community 42 - "Enrollment Service"
Cohesion: 0.27
Nodes (7): Client, Context, Pool, NewService(), PhotoRejection, RejectedError, Service

### Community 43 - "E2E API Test Helpers"
Cohesion: 0.18
Nodes (8): _checkout(), face(), prepare_faces(), End-to-end test against the real Go API + real Postgres/pgvector.  Runs against, Like request() but returns the raw body -- the CSV export is not JSON., Two LFW identities (two photos each) plus a third as the unknown face.      Enro, request(), request_raw()

### Community 44 - "Corrections UI"
Cohesion: 0.19
Nodes (10): Correction, listCorrections(), listMyCorrections(), reviewCorrection(), columns, items, loading, note (+2 more)

### Community 45 - "Master Data API Client"
Cohesion: 0.17
Nodes (11): createLocation(), createSchedule(), Device, listDevices(), Location, Schedule, updateLocation(), updateSchedule() (+3 more)

### Community 46 - "Contrast Check Script (agent)"
Cohesion: 0.32
Nodes (11): contrast_ratio(), linearize(), luminance(), main(), parse_hex(), parse_pairing(), parse_reference_rows(), Turn 'White on #333333' or '#555555 on black' into two RGB tuples. (+3 more)

### Community 47 - "Contrast Check Script (.agents)"
Cohesion: 0.32
Nodes (11): contrast_ratio(), linearize(), luminance(), main(), parse_hex(), parse_pairing(), parse_reference_rows(), Turn 'White on #333333' or '#555555 on black' into two RGB tuples. (+3 more)

### Community 48 - "Attendance Photo Store"
Cohesion: 0.24
Nodes (7): Duration, Time, NewPhotoStore(), TestPhotoStoreDisabled(), TestPhotoStoreRetention(), PhotoStore, T

### Community 49 - "PRD Overview Sections"
Cohesion: 0.17
Nodes (12): 12. Risiko Produk, 13. Rencana Rilis, 14. Pertanyaan Terbuka, 1. Ringkasan, 2. Masalah, 3.1 Hasil pengukuran (2026-09-07), 3. Tujuan & Metrik Sukses, 4. Skala Target (+4 more)

### Community 50 - "Inline Master Select"
Cohesion: 0.20
Nodes (9): busy, cancel(), draft, emit, Mode, model, props, save() (+1 more)

### Community 51 - "Slop Layout Patterns (agent)"
Cohesion: 0.18
Nodes (11): 4-Column Template Footer, Bento Grid, Copy-Paste Feature Cards, Demo Without a Product, "How It Works" Always 3 Steps, Layout & Components, Monotonous Template Layout, "Most Popular" Pricing Card (+3 more)

### Community 52 - "Slop Visual Patterns (agent)"
Cohesion: 0.18
Nodes (11): Background Grid, Dark Mode Default for No Reason, Excessive Accent Color, Excessive Border Radius, Excessive Glassmorphism, Generic Blue-Purple Gradient, Glow Everywhere, Overly Soft Shadows (+3 more)

### Community 53 - "Slop Layout Patterns (.agents)"
Cohesion: 0.18
Nodes (11): 4-Column Template Footer, Bento Grid, Copy-Paste Feature Cards, Demo Without a Product, "How It Works" Always 3 Steps, Layout & Components, Monotonous Template Layout, "Most Popular" Pricing Card (+3 more)

### Community 54 - "Slop Visual Patterns (.agents)"
Cohesion: 0.18
Nodes (11): Background Grid, Dark Mode Default for No Reason, Excessive Accent Color, Excessive Border Radius, Excessive Glassmorphism, Generic Blue-Purple Gradient, Glow Everywhere, Overly Soft Shadows (+3 more)

### Community 55 - "Login Throttle"
Cohesion: 0.29
Nodes (6): Duration, Mutex, Time, newThrottle(), attemptRecord, throttle

### Community 56 - "Core DB Schema"
Cohesion: 0.40
Nodes (10): attendance_corrections, attendances, audit_logs, departments, devices, employees, face_embeddings, users (+2 more)

### Community 57 - "Employee Nav And Team"
Cohesion: 0.18
Nodes (8): auth, canSeeTeam, initials, isAdmin, links, profileOpen, route, router

### Community 58 - "Contrast MCP Server (agent)"
Cohesion: 0.38
Nodes (9): _channel(), check_contrast(), contrast_ratio(), _error(), main(), relative_luminance(), _reply(), _send() (+1 more)

### Community 59 - "Antislop Quality Locks (agent)"
Cohesion: 0.20
Nodes (10): Group 3: Quality Locks (consistency), R-05 — Layout & Page Structure, R-11 — Border Radius, R-15 — CTA (Call to Action), R-16 — Copywriting & Buzzwords, R-20 — Visual Identity, R-21 — Dark Mode, R-29 — Color Palette (+2 more)

### Community 60 - "Slop Decorative Patterns (agent)"
Cohesion: 0.20
Nodes (10): AI Capsule Badges, Colored Left Stripe, Decorative Elements, Emoji as Decoration, Fake Terminal Window, Generic AI Icons, Generic AI Typography, Illustrations With No Connection (+2 more)

### Community 61 - "Contrast MCP Server (.agents)"
Cohesion: 0.38
Nodes (9): _channel(), check_contrast(), contrast_ratio(), _error(), main(), relative_luminance(), _reply(), _send() (+1 more)

### Community 62 - "Antislop Quality Locks (.agents)"
Cohesion: 0.20
Nodes (10): Group 3: Quality Locks (consistency), R-05 — Layout & Page Structure, R-11 — Border Radius, R-15 — CTA (Call to Action), R-16 — Copywriting & Buzzwords, R-20 — Visual Identity, R-21 — Dark Mode, R-29 — Color Palette (+2 more)

### Community 63 - "Slop Decorative Patterns (.agents)"
Cohesion: 0.20
Nodes (10): AI Capsule Badges, Colored Left Stripe, Decorative Elements, Emoji as Decoration, Fake Terminal Window, Generic AI Icons, Generic AI Typography, Illustrations With No Connection (+2 more)

### Community 64 - "MCP Documentation"
Cohesion: 0.20
Nodes (9): Build, Catatan operasional, Claude Code, Claude Desktop, Konfigurasi, MCP Server — HRIS Face Attendance, Sifatnya: hanya membaca, Tes (+1 more)

### Community 65 - "Face Service Docs"
Cohesion: 0.20
Nodes (6): Accuracy, Anti-spoof models, Face service, Run, Setup, Two things that will silently break this model

### Community 66 - "MCP E2E Test"
Cohesion: 0.24
Nodes (4): call_tool(), MCP, Drives the MCP server over stdio the way an AI client does: real JSON-RPC on std, Minimal MCP stdio client: newline-delimited JSON-RPC 2.0.

### Community 67 - "Web App Bootstrap"
Cohesion: 0.27
Nodes (4): i18n, router, attendanceStatusColor, themeTokens

### Community 68 - "Admin Filters And Review"
Cohesion: 0.22
Nodes (10): tzDayjs(), confirmReview(), openEdit(), defaultRange(), onApplyFilter(), onEditOpen(), onRekapBulanIni(), onResetFilter() (+2 more)

### Community 69 - "API Config Loading"
Cohesion: 0.42
Nodes (8): atoiDefault(), getenv(), Location, Load(), mustEnv(), mustLocation(), splitCSV(), Config

### Community 70 - "PRD Frontend Direction"
Cohesion: 0.22
Nodes (9): 10.1 Tesis layar absen, 10.2 Sistem visual, 10.3 Layar absen — komposisi, 10.4 Layar admin, 10.5 State yang wajib ada (bukan opsional), 10.6 Motion, 10.7 Aksesibilitas & i18n, 10.8 Dependency frontend (+1 more)

### Community 71 - "PRD Functional Requirements"
Cohesion: 0.22
Nodes (9): 7. Kebutuhan Fungsional, F1 — Autentikasi & Otorisasi, F2 — Master Data, F3 — Face Enrollment, F4 — Absensi, F5 — Anti-Spoofing (wajib rilis 1), F6 — Dashboard & Laporan, F7 — Koreksi & Audit (+1 more)

### Community 72 - "Web Auth Client"
Cohesion: 0.31
Nodes (5): refreshAccessToken(), useAuthStore, auth, homePath, route

### Community 73 - "Master Data Loaders"
Cohesion: 0.31
Nodes (9): listDepartments(), listEmployees(), updateDepartment(), listLocations(), listSchedules(), loadAll(), reloadMasterLists(), load() (+1 more)

### Community 74 - "Admin Layout Shell"
Cohesion: 0.22
Nodes (7): auth, initials, pageTitle, profileOpen, route, router, selectedKeys

### Community 75 - "Antislop Core Skill (agent)"
Cohesion: 0.25
Nodes (7): antislop, Core Principle, First-Run Install Wizard, Functional Patterns, Part 2: Mandatory Rules (R-01 to R-38, grouped), Two Usage Modes, What This Is (and What It Isn't)

### Community 76 - "Slop Pattern Catalog (agent)"
Cohesion: 0.25
Nodes (8): Accessibility, Copywriting & Content, Decorative Elements, Functionality & Content, Identity & Originality, Layout & Components, Part 1: AI Slop Patterns (Warning Signs), Visual & Color

### Community 77 - "Slop Dashboard Patterns (agent)"
Cohesion: 0.25
Nodes (8): App & Dashboard, Charts Without a Question, Default Dashboard Shell, Filler Activity Feed, Filler Data in Fields and Columns, Generic Table Columns, Placeholder Empty and Loading States, Stat Cards With Invented Numbers

### Community 78 - "Antislop Core Skill (.agents)"
Cohesion: 0.25
Nodes (7): antislop, Core Principle, First-Run Install Wizard, Functional Patterns, Part 2: Mandatory Rules (R-01 to R-38, grouped), Two Usage Modes, What This Is (and What It Isn't)

### Community 79 - "Slop Pattern Catalog (.agents)"
Cohesion: 0.25
Nodes (8): Accessibility, Copywriting & Content, Decorative Elements, Functionality & Content, Identity & Originality, Layout & Components, Part 1: AI Slop Patterns (Warning Signs), Visual & Color

### Community 80 - "Slop Dashboard Patterns (.agents)"
Cohesion: 0.25
Nodes (8): App & Dashboard, Charts Without a Question, Default Dashboard Shell, Filler Activity Feed, Filler Data in Fields and Columns, Generic Table Columns, Placeholder Empty and Loading States, Stat Cards With Invented Numbers

### Community 81 - "Local Run Script"
Cohesion: 0.25
Nodes (6): DATABASE_URL, FACE_SERVICE_URL, JWT_SECRET, LISTEN_ADDR, OFFICE_IP_ALLOWLIST, run-local.sh script

### Community 82 - "All Pages Browser Test"
Cohesion: 0.29
Nodes (4): check(), { chromium }, HR, visit()

### Community 83 - "Antislop UI Motion (agent)"
Cohesion: 0.29
Nodes (6): antislop-ui, Endless Pulses and Loops, How to use this skill, Motion, Template Animations Stacked, UI Skill Checklist

### Community 84 - "Antislop UI Motion (.agents)"
Cohesion: 0.29
Nodes (6): antislop-ui, Endless Pulses and Loops, How to use this skill, Motion, Template Animations Stacked, UI Skill Checklist

### Community 85 - "PRD Deployment Plan"
Cohesion: 0.29
Nodes (7): 11.1 Komponen, 11.2 Caddy, 11.3 systemd, 11.4 Deploy, 11.5 Backup & operasional, 11.6 Risiko VPS satu mesin, 11. Deployment — VPS Tanpa Docker

### Community 86 - "Product Brief"
Cohesion: 0.29
Nodes (7): Apa ini, Batasan yang mengikat, Bukan tujuan (rilis 1), Mekanisme unik, Pengguna & scene nyata, PRODUCT.md — HRIS Absensi Face Recognition, Yang bikin hasil poles terasa salah

### Community 87 - "Login Page"
Cohesion: 0.29
Nodes (5): auth, errorMessage, form, router, submitting

### Community 88 - "Craftsmanship Standard (agent)"
Cohesion: 0.33
Nodes (6): C-1 — Intentionality, C-2 — Functional Completeness, C-3 — Content-Driven Composition, C-4 — Resilience, C-5 — Evidence Over Claims, The Craftsmanship Standard

### Community 89 - "Craftsmanship Standard (.agents)"
Cohesion: 0.33
Nodes (6): C-1 — Intentionality, C-2 — Functional Completeness, C-3 — Content-Driven Composition, C-4 — Resilience, C-5 — Evidence Over Claims, The Craftsmanship Standard

### Community 90 - "Enrollment Handlers"
Cohesion: 0.47
Nodes (5): HandlerFunc, IRoutes, Service, RegisterRoutes(), uploadHandler()

### Community 91 - "PRD Architecture"
Cohesion: 0.33
Nodes (6): 9. Arsitektur, API (garis besar), Kenapa 3 bahasa, Model, Pencocokan — ArcFace + Cosine Similarity (final), Skema data inti

### Community 92 - "Project README"
Cohesion: 0.33
Nodes (6): Coba sendiri (satu perintah), Deploy ke VPS, Development, HRIS Absensi — Face Recognition, Status, Struktur

### Community 94 - "Leave E2E Test"
Cohesion: 0.33
Nodes (3): prev_weekday(), Verifies the leave/permission/sick (cuti/izin/sakit) module: create, list, balan, Most recent date (<= d) with the given Python weekday (Mon=0..Sun=6).

### Community 95 - "Antislop Delivery Gate (agent)"
Cohesion: 0.40
Nodes (5): Block 1: Hard Gate (absolute), Block 2: Purpose-Gate (technique allowed, reason required), Block 3: Liveliness (required to be alive, not just clean), Block 4: Craftsmanship & Quality Locks, Delivery Gate (Mandatory)

### Community 96 - "Antislop Delivery Gate (.agents)"
Cohesion: 0.40
Nodes (5): Block 1: Hard Gate (absolute), Block 2: Purpose-Gate (technique allowed, reason required), Block 3: Liveliness (required to be alive, not just clean), Block 4: Craftsmanship & Quality Locks, Delivery Gate (Mandatory)

### Community 97 - "Face Service Client"
Cohesion: 0.60
Nodes (3): New(), AnalyzeResult, Client

### Community 98 - "Anti-spoof Model Export"
Cohesion: 0.50
Nodes (4): fetch(), main(), Path, Exports the official Silent-Face-Anti-Spoofing weights to ONNX.  Provenance matt

### Community 99 - "E2E Test Docs"
Cohesion: 0.40
Nodes (4): End-to-end tests and accuracy benchmarks, Running, Test identities, What these tests still do not cover

### Community 100 - "Demo Data Seeder"
Cohesion: 0.40
Nodes (3): Fills a freshly-seeded database with realistic demo data so every screen has som, A date `offset_days` back from today, skipped back off weekends., workday()

### Community 104 - "Attendance API Client"
Cohesion: 0.40
Nodes (3): CheckInError, CheckInErrorCode, CheckInResult

### Community 105 - "Liveliness Toolkit (agent)"
Cohesion: 0.50
Nodes (4): Design Read (how the dials are set), Levers (how the dials become visual decisions), Part 3: Liveliness Toolkit, Three Dials (required)

### Community 106 - "Slop Structural Patterns (agent)"
Cohesion: 0.50
Nodes (4): Dead Navigation, Non-Functional Controls, Sections That Fill a Template, Structural & Flow

### Community 107 - "Liveliness Toolkit (.agents)"
Cohesion: 0.50
Nodes (4): Design Read (how the dials are set), Levers (how the dials become visual decisions), Part 3: Liveliness Toolkit, Three Dials (required)

### Community 108 - "Slop Structural Patterns (.agents)"
Cohesion: 0.50
Nodes (4): Dead Navigation, Non-Functional Controls, Sections That Fill a Template, Structural & Flow

### Community 109 - "Rate Limit Middleware"
Cohesion: 0.50
Nodes (3): Duration, HandlerFunc, RateLimit()

### Community 110 - "PRD Main Flows"
Cohesion: 0.50
Nodes (4): 6.1 Enrollment (pendaftaran wajah), 6.2 Check-in / Check-out, 6.3 Koreksi, 6. Alur Utama

### Community 113 - "Enrollment API Client"
Cohesion: 0.50
Nodes (3): EnrollOutcome, PhotoRejection, submitEnrollment()

## Knowledge Gaps
- **729 isolated node(s):** `loginRequest`, `createRequest`, `reviewRequest`, `refresh_tokens`, `deploy.sh script` (+724 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **24 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `tzDayjs()` connect `Admin Filters And Review` to `Admin Monitoring Page`, `Admin Master Data Page`, `Corrections UI`, `Check-in Page`, `Admin Leave Review`, `Employee History Page`, `Employee Leave Page`?**
  _High betweenness centrality (0.009) - this node is a cross-community bridge._
- **Why does `apiClient` connect `Admin Face Scan` to `Employee API Client`, `Admin Monitoring Page`, `Attendance API Client`, `Web Auth Client`, `Corrections UI`, `Master Data API Client`, `Enrollment API Client`, `Login Page`, `Employee Leave Page`?**
  _High betweenness centrality (0.004) - this node is a cross-community bridge._
- **Why does `useAuthStore` connect `Web Auth Client` to `Web App Bootstrap`, `Profile Edit Modal`, `Admin Layout Shell`, `Login Page`, `Employee Nav And Team`?**
  _High betweenness centrality (0.004) - this node is a cross-community bridge._
- **What connects `loginRequest`, `createRequest`, `reviewRequest` to the rest of the system?**
  _729 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Face Recognition Service` be split into smaller, more focused modules?**
  _Cohesion score 0.05152394775036284 - nodes in this community are weakly interconnected._
- **Should `Web Package Dependencies` be split into smaller, more focused modules?**
  _Cohesion score 0.04878048780487805 - nodes in this community are weakly interconnected._
- **Should `MCP Server Tools` be split into smaller, more focused modules?**
  _Cohesion score 0.09634146341463415 - nodes in this community are weakly interconnected._