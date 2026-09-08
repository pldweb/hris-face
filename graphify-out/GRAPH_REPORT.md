# Graph Report - hris-face  (2026-09-08)

## Corpus Check
- 129 files · ~105,927 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1537 nodes · 2150 edges · 128 communities (111 shown, 17 thin omitted)
- Extraction: 99% EXTRACTED · 1% INFERRED · 0% AMBIGUOUS · INFERRED: 26 edges (avg confidence: 0.71)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `bb844ef1`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Enrollment.vue
- dependencies
- face_engine.py
- CheckIn.vue
- MyHistory.vue
- EmployeeList.vue
- AdminLayout.vue
- Service
- MasterData.vue
- main
- Monitoring.vue
- Service
- compilerOptions
- Service
- employee.ts
- Service
- report/handler.go
- Service
- Service
- compilerOptions
- employee/handler.go
- auth/handler.go
- RegisterRoutes
- Service
- Corrections.vue
- Tone & Voice
- e2e_api.py
- PRD — HRIS Absensi Berbasis Face Recognition
- throttle
- employees
- .ImportCSV
- run-local.sh
- RegisterRoutes
- e2e_admin_browser.js
- export_antispoof_onnx.py
- e2e_browser.js
- tsconfig.json
- 0002_refresh_tokens.up.sql
- deploy.sh
- github.com/hris-face/api
- Face service
- End-to-end tests and accuracy benchmarks
- AGENTS.md
- web-src-views-checkin-vue.md
- web/README.md
- Tone & Voice
- tzDayjs
- antislop-layoutmobile
- main.ts
- EmployeeNav.vue
- client.ts
- test_admin_ui.js
- test_admin_crud.py
- antislop-layoutmobile
- antislop-human
- antislop-human
- Group 1: Hard Gate (absolute, no exceptions)
- Group 1: Hard Gate (absolute, no exceptions)
- LeaveRequest.vue
- Comments That Add Nothing
- Comments That Add Nothing
- FaceScan.vue
- LeaveRequests.vue
- leave/handler.go
- Group 2: Purpose-Gate (technique allowed, purpose required)
- Group 2: Purpose-Gate (technique allowed, purpose required)
- agent/skills/antislop-human/contrast-check.py
- .agents/skills/antislop-human/contrast-check.py
- Layout & Components
- Visual & Color
- Layout & Components
- Visual & Color
- agent/skills/antislop-human/contrast-mcp.py
- Group 3: Quality Locks (consistency)
- Decorative Elements
- .agents/skills/antislop-human/contrast-mcp.py
- Group 3: Quality Locks (consistency)
- Decorative Elements
- antislop
- Part 1: AI Slop Patterns (Warning Signs)
- App & Dashboard
- antislop
- Part 1: AI Slop Patterns (Warning Signs)
- App & Dashboard
- antislop-ui
- antislop-ui
- The Craftsmanship Standard
- The Craftsmanship Standard
- test_leave.py
- Delivery Gate (Mandatory)
- Delivery Gate (Mandatory)
- Part 3: Liveliness Toolkit
- Structural & Flow
- Part 3: Liveliness Toolkit
- Structural & Flow
- 0004_leave_requests.up.sql
- 0004_leave_requests.down.sql
- liveness.py
- FaceEngine
- main.py
- test_challenge.py
- test_all_pages.js
- benchmark_liveness.py
- seed_demo.py
- ProfileModal.vue
- masterdata.ts
- InlineMasterSelect.vue
- RegisterRoutes
- 10. Arah Desain Frontend
- 7. Kebutuhan Fungsional
- load
- 11. Deployment — VPS Tanpa Docker
- PRODUCT.md — HRIS Absensi Face Recognition
- Login.vue
- 9. Arsitektur
- HRIS Absensi — Face Recognition
- test_hr_editing.py
- test_hr_editing_ui.js
- attendance.ts
- 6. Alur Utama
- 0005_profile_and_display_name.down.sql
- 0005_profile_and_display_name.up.sql

## God Nodes (most connected - your core abstractions)
1. `Service` - 19 edges
2. `RegisterRoutes()` - 18 edges
3. `Group 1: Hard Gate (absolute, no exceptions)` - 18 edges
4. `Group 1: Hard Gate (absolute, no exceptions)` - 18 edges
5. `Service` - 16 edges
6. `PRD — HRIS Absensi Berbasis Face Recognition` - 16 edges
7. `tzDayjs()` - 15 edges
8. `compilerOptions` - 15 edges
9. `Service` - 13 edges
10. `RegisterRoutes()` - 13 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `FaceEngine`  [EXTRACTED]
  test/e2e/benchmark_liveness.py → face/app/face_engine.py
- `main()` --calls--> `LivenessDetector`  [EXTRACTED]
  test/e2e/benchmark_liveness.py → face/app/liveness.py
- `main()` --calls--> `NewPhotoStore()`  [INFERRED]
  api/cmd/api/main.go → api/internal/attendance/photostore.go
- `main()` --calls--> `Migrate()`  [INFERRED]
  api/cmd/api/main.go → api/internal/db/migrate.go
- `RegisterRoutes()` --calls--> `RequireRole()`  [INFERRED]
  api/internal/correction/handler.go → api/internal/middleware/auth.go

## Import Cycles
- None detected.

## Communities (128 total, 17 thin omitted)

### Community 0 - "Enrollment.vue"
Cohesion: 0.09
Nodes (21): EnrollOutcome, PhotoRejection, submitEnrollment(), cameraError, cameraReady, canCapture, canSubmit, canvasEl (+13 more)

### Community 1 - "dependencies"
Cohesion: 0.05
Nodes (40): @ant-design/icons-vue, ant-design-vue, axios, dayjs, @mediapipe/tasks-vision, pinia, @types/node, typescript (+32 more)

### Community 2 - "face_engine.py"
Cohesion: 0.23
Nodes (9): eye_openness(), face_signature(), ndarray, Eye openness, used for the blink challenge (docs/PRD.md F5).  Passive anti-spoof, Mean vertical eye opening as a fraction of face width.      Normalising by face, An 8x8 grayscale thumbnail of the face crop.      The blink challenge needs to p, DetectedFace, ndarray (+1 more)

### Community 3 - "CheckIn.vue"
Cohesion: 0.09
Nodes (27): submitAttendance(), submitChallenge(), fetchMe(), actionLabel, CameraError, canvasEl, captureBurst(), captureJpeg() (+19 more)

### Community 4 - "MyHistory.vue"
Cohesion: 0.15
Nodes (15): createCorrection(), listMyCorrections(), fetchMyHistory(), corrections, days, form, load(), loading (+7 more)

### Community 5 - "EmployeeList.vue"
Cohesion: 0.07
Nodes (27): deactivateEmployee(), importEmployeesCsv(), columns, deactivatingId, departmentOptions, departments, editingId, employees (+19 more)

### Community 6 - "AdminLayout.vue"
Cohesion: 0.22
Nodes (7): auth, initials, pageTitle, profileOpen, route, router, selectedKeys

### Community 7 - "Service"
Cohesion: 0.10
Nodes (22): Service, Context, meanPairwiseVariation(), Duration, Time, NewPhotoStore(), TestPhotoStoreDisabled(), TestPhotoStoreRetention() (+14 more)

### Community 8 - "MasterData.vue"
Cohesion: 0.07
Nodes (28): deleteDepartment(), approveDevice(), deleteLocation(), DAY_LABELS, deletingDepartmentId, deletingLocationId, departmentForm, departmentModal (+20 more)

### Community 9 - "main"
Cohesion: 0.07
Nodes (34): main(), main(), mustEnv(), challengeHandler(), HandlerFunc, IRoutes, Service, markHandler() (+26 more)

### Community 10 - "Monitoring.vue"
Cohesion: 0.08
Nodes (35): AttendanceFilter, AttendanceRow, DailySummary, DayRecord, deleteAttendance(), exportReport(), fetchToday(), listAttendances() (+27 more)

### Community 11 - "Service"
Cohesion: 0.14
Nodes (15): Context, Service, Pool, Time, NewService(), randomPassword(), CreateEmployeeInput, CreateEmployeeResult (+7 more)

### Community 12 - "compilerOptions"
Cohesion: 0.10
Nodes (19): ES2023, node, vite.config.ts, compilerOptions, allowImportingTsExtensions, erasableSyntaxOnly, lib, module (+11 more)

### Community 13 - "Service"
Cohesion: 0.11
Nodes (19): Context, Pool, Time, NewService(), countWorkdays(), Context, Pool, Time (+11 more)

### Community 14 - "employee.ts"
Cohesion: 0.13
Nodes (15): createDepartment(), createEmployee(), CreateEmployeeInput, CreateEmployeeResult, Department, Employee, EmployeeStatus, ImportResult (+7 more)

### Community 15 - "Service"
Cohesion: 0.20
Nodes (9): Context, Pool, NewService(), New(), PhotoRejection, RejectedError, Service, AnalyzeResult (+1 more)

### Community 16 - "report/handler.go"
Cohesion: 0.34
Nodes (19): boolLabel(), deleteHandler(), exportHandler(), exportXlsxHandler(), filterFrom(), Context, HandlerFunc, IRoutes (+11 more)

### Community 17 - "Service"
Cohesion: 0.21
Nodes (11): Context, Location, Pool, Time, NewService(), DailySummary, DayRecord, Filter (+3 more)

### Community 18 - "Service"
Cohesion: 0.25
Nodes (10): Context, Duration, Pool, RegisteredClaims, hashToken(), NewService(), randomToken(), Claims (+2 more)

### Community 19 - "compilerOptions"
Cohesion: 0.12
Nodes (15): src/**/*.ts, src/**/*.tsx, src/**/*.vue, vite/client, @vue/tsconfig/tsconfig.dom.json, compilerOptions, allowArbitraryExtensions, erasableSyntaxOnly (+7 more)

### Community 20 - "employee/handler.go"
Cohesion: 0.25
Nodes (22): approveDeviceHandler(), createDepartmentHandler(), createHandler(), deactivateHandler(), deleteDepartmentHandler(), HandlerFunc, IRoutes, Service (+14 more)

### Community 21 - "auth/handler.go"
Cohesion: 0.36
Nodes (12): clearRefreshCookie(), Context, Duration, HandlerFunc, IRoutes, Service, loginHandler(), logoutHandler() (+4 more)

### Community 22 - "RegisterRoutes"
Cohesion: 0.35
Nodes (14): assignScheduleHandler(), createLocationHandler(), createScheduleHandler(), deleteLocationHandler(), HandlerFunc, IRoutes, Service, listLocationsHandler() (+6 more)

### Community 23 - "Service"
Cohesion: 0.27
Nodes (6): Context, Pool, NewService(), Location, Schedule, Service

### Community 24 - "Corrections.vue"
Cohesion: 0.22
Nodes (11): Correction, listCorrections(), reviewCorrection(), columns, confirmReview(), items, load(), loading (+3 more)

### Community 25 - "Tone & Voice"
Cohesion: 0.05
Nodes (38): Actorless Passive, All-Caps Emphasis, antislop-copywriting, Aphorism Formulas, Boldface Overuse, Chatbot Closers, Copywriting Skill Checklist, Draft, audit, final (+30 more)

### Community 26 - "e2e_api.py"
Cohesion: 0.17
Nodes (5): prepare_faces(), End-to-end test against the real Go API + real Postgres/pgvector.  Runs against, Like request() but returns the raw body -- the CSV export is not JSON., Two LFW identities (two photos each) plus a third as the unknown face.      Enro, request_raw()

### Community 27 - "PRD — HRIS Absensi Berbasis Face Recognition"
Cohesion: 0.17
Nodes (12): 12. Risiko Produk, 13. Rencana Rilis, 14. Pertanyaan Terbuka, 1. Ringkasan, 2. Masalah, 3.1 Hasil pengukuran (2026-09-07), 3. Tujuan & Metrik Sukses, 4. Skala Target (+4 more)

### Community 28 - "throttle"
Cohesion: 0.29
Nodes (6): Duration, Time, newThrottle(), attemptRecord, throttle, Mutex

### Community 29 - "employees"
Cohesion: 0.40
Nodes (10): attendance_corrections, attendances, audit_logs, departments, devices, employees, face_embeddings, users (+2 more)

### Community 30 - ".ImportCSV"
Cohesion: 0.33
Nodes (6): Context, Service, mapHeader(), ImportResult, ImportRow, Reader

### Community 31 - "run-local.sh"
Cohesion: 0.25
Nodes (6): DATABASE_URL, FACE_SERVICE_URL, JWT_SECRET, LISTEN_ADDR, OFFICE_IP_ALLOWLIST, run-local.sh script

### Community 32 - "RegisterRoutes"
Cohesion: 0.47
Nodes (5): HandlerFunc, IRoutes, Service, RegisterRoutes(), uploadHandler()

### Community 34 - "export_antispoof_onnx.py"
Cohesion: 0.50
Nodes (4): fetch(), main(), Path, Exports the official Silent-Face-Anti-Spoofing weights to ONNX.  Provenance matt

### Community 46 - "Face service"
Cohesion: 0.20
Nodes (6): Accuracy, Anti-spoof models, Face service, Run, Setup, Two things that will silently break this model

### Community 47 - "End-to-end tests and accuracy benchmarks"
Cohesion: 0.40
Nodes (4): End-to-end tests and accuracy benchmarks, Running, Test identities, What these tests still do not cover

### Community 51 - "Tone & Voice"
Cohesion: 0.05
Nodes (38): Actorless Passive, All-Caps Emphasis, antislop-copywriting, Aphorism Formulas, Boldface Overuse, Chatbot Closers, Copywriting Skill Checklist, Draft, audit, final (+30 more)

### Community 52 - "tzDayjs"
Cohesion: 0.25
Nodes (7): Dayjs, DayjsArgs, tzDayjs(), openEdit(), onEditOpen(), openModal(), openCorrection()

### Community 53 - "antislop-layoutmobile"
Cohesion: 0.07
Nodes (29): 100vh Sections, antislop-layoutmobile, Bottom Nav That Eats Content, Breakpoint Driven by Device List, Breakpoints, Columns That Don't Collapse, Desktop-Only Layout, Desktop-Sized Everything (+21 more)

### Community 54 - "main.ts"
Cohesion: 0.27
Nodes (4): i18n, router, attendanceStatusColor, themeTokens

### Community 55 - "EmployeeNav.vue"
Cohesion: 0.18
Nodes (8): auth, canSeeTeam, initials, isAdmin, links, profileOpen, route, router

### Community 56 - "client.ts"
Cohesion: 0.31
Nodes (5): refreshAccessToken(), useAuthStore, auth, homePath, route

### Community 59 - "antislop-layoutmobile"
Cohesion: 0.07
Nodes (29): 100vh Sections, antislop-layoutmobile, Bottom Nav That Eats Content, Breakpoint Driven by Device List, Breakpoints, Columns That Don't Collapse, Desktop-Only Layout, Desktop-Sized Everything (+21 more)

### Community 60 - "antislop-human"
Cohesion: 0.10
Nodes (20): antislop-human, Broken Tab Order, Color & Contrast, Color-Only Feedback, Focus & States, How to use this skill, Human Skill Checklist, Keyboard (+12 more)

### Community 61 - "antislop-human"
Cohesion: 0.10
Nodes (20): antislop-human, Broken Tab Order, Color & Contrast, Color-Only Feedback, Focus & States, How to use this skill, Human Skill Checklist, Keyboard (+12 more)

### Community 62 - "Group 1: Hard Gate (absolute, no exceptions)"
Cohesion: 0.11
Nodes (18): Group 1: Hard Gate (absolute, no exceptions), R-02 — Copywriting, R-03 — Mobile Responsiveness, R-17 — Data & Numbers, R-18 — Testimonials, R-23 — Clarification & Visual Assets, R-24 — Navigation, R-25 — Color Contrast (+10 more)

### Community 63 - "Group 1: Hard Gate (absolute, no exceptions)"
Cohesion: 0.11
Nodes (18): Group 1: Hard Gate (absolute, no exceptions), R-02 — Copywriting, R-03 — Mobile Responsiveness, R-17 — Data & Numbers, R-18 — Testimonials, R-23 — Clarification & Visual Assets, R-24 — Navigation, R-25 — Color Contrast (+10 more)

### Community 64 - "LeaveRequest.vue"
Cohesion: 0.16
Nodes (17): createLeaveRequest(), fetchMyLeaveBalance(), LeaveBalance, LeaveRequest, LeaveStatus, LeaveType, listMyLeaveRequests(), balance (+9 more)

### Community 65 - "Comments That Add Nothing"
Cohesion: 0.12
Nodes (16): antislop-code, Code Comment Checklist, Comments That Add Nothing, Decorative Emoji, Decorative Separators, Empty Labels, End Markers, How It Should Read (+8 more)

### Community 66 - "Comments That Add Nothing"
Cohesion: 0.12
Nodes (16): antislop-code, Code Comment Checklist, Comments That Add Nothing, Decorative Emoji, Decorative Separators, Empty Labels, End Markers, How It Should Read (+8 more)

### Community 67 - "FaceScan.vue"
Cohesion: 0.13
Nodes (15): apiClient, FaceScanResult, scanFace(), cameraError, canScan, canvasEl, captureJpeg(), onScan() (+7 more)

### Community 68 - "LeaveRequests.vue"
Cohesion: 0.12
Nodes (20): decideLeaveRequest(), listAdminLeaveRequests(), updateLeaveRequest(), columns, confirmReview(), editForm, editing, editOpen (+12 more)

### Community 69 - "leave/handler.go"
Cohesion: 0.36
Nodes (15): createHandler(), Context, HandlerFunc, IRoutes, Service, listHandler(), myBalanceHandler(), myListHandler() (+7 more)

### Community 70 - "Group 2: Purpose-Gate (technique allowed, purpose required)"
Cohesion: 0.15
Nodes (13): Group 2: Purpose-Gate (technique allowed, purpose required), R-01 — Color & Gradients, R-04 — Icons, R-06 — Typography, R-07 — Background, R-08 — Button Arrows, R-09 — Badges, R-10 — Glassmorphism (+5 more)

### Community 71 - "Group 2: Purpose-Gate (technique allowed, purpose required)"
Cohesion: 0.15
Nodes (13): Group 2: Purpose-Gate (technique allowed, purpose required), R-01 — Color & Gradients, R-04 — Icons, R-06 — Typography, R-07 — Background, R-08 — Button Arrows, R-09 — Badges, R-10 — Glassmorphism (+5 more)

### Community 72 - "agent/skills/antislop-human/contrast-check.py"
Cohesion: 0.32
Nodes (11): contrast_ratio(), linearize(), luminance(), main(), parse_hex(), parse_pairing(), parse_reference_rows(), Turn 'White on #333333' or '#555555 on black' into two RGB tuples. (+3 more)

### Community 73 - ".agents/skills/antislop-human/contrast-check.py"
Cohesion: 0.32
Nodes (11): contrast_ratio(), linearize(), luminance(), main(), parse_hex(), parse_pairing(), parse_reference_rows(), Turn 'White on #333333' or '#555555 on black' into two RGB tuples. (+3 more)

### Community 74 - "Layout & Components"
Cohesion: 0.18
Nodes (11): 4-Column Template Footer, Bento Grid, Copy-Paste Feature Cards, Demo Without a Product, "How It Works" Always 3 Steps, Layout & Components, Monotonous Template Layout, "Most Popular" Pricing Card (+3 more)

### Community 75 - "Visual & Color"
Cohesion: 0.18
Nodes (11): Background Grid, Dark Mode Default for No Reason, Excessive Accent Color, Excessive Border Radius, Excessive Glassmorphism, Generic Blue-Purple Gradient, Glow Everywhere, Overly Soft Shadows (+3 more)

### Community 76 - "Layout & Components"
Cohesion: 0.18
Nodes (11): 4-Column Template Footer, Bento Grid, Copy-Paste Feature Cards, Demo Without a Product, "How It Works" Always 3 Steps, Layout & Components, Monotonous Template Layout, "Most Popular" Pricing Card (+3 more)

### Community 77 - "Visual & Color"
Cohesion: 0.18
Nodes (11): Background Grid, Dark Mode Default for No Reason, Excessive Accent Color, Excessive Border Radius, Excessive Glassmorphism, Generic Blue-Purple Gradient, Glow Everywhere, Overly Soft Shadows (+3 more)

### Community 78 - "agent/skills/antislop-human/contrast-mcp.py"
Cohesion: 0.38
Nodes (9): _channel(), check_contrast(), contrast_ratio(), _error(), main(), relative_luminance(), _reply(), _send() (+1 more)

### Community 79 - "Group 3: Quality Locks (consistency)"
Cohesion: 0.20
Nodes (10): Group 3: Quality Locks (consistency), R-05 — Layout & Page Structure, R-11 — Border Radius, R-15 — CTA (Call to Action), R-16 — Copywriting & Buzzwords, R-20 — Visual Identity, R-21 — Dark Mode, R-29 — Color Palette (+2 more)

### Community 80 - "Decorative Elements"
Cohesion: 0.20
Nodes (10): AI Capsule Badges, Colored Left Stripe, Decorative Elements, Emoji as Decoration, Fake Terminal Window, Generic AI Icons, Generic AI Typography, Illustrations With No Connection (+2 more)

### Community 81 - ".agents/skills/antislop-human/contrast-mcp.py"
Cohesion: 0.38
Nodes (9): _channel(), check_contrast(), contrast_ratio(), _error(), main(), relative_luminance(), _reply(), _send() (+1 more)

### Community 82 - "Group 3: Quality Locks (consistency)"
Cohesion: 0.20
Nodes (10): Group 3: Quality Locks (consistency), R-05 — Layout & Page Structure, R-11 — Border Radius, R-15 — CTA (Call to Action), R-16 — Copywriting & Buzzwords, R-20 — Visual Identity, R-21 — Dark Mode, R-29 — Color Palette (+2 more)

### Community 83 - "Decorative Elements"
Cohesion: 0.20
Nodes (10): AI Capsule Badges, Colored Left Stripe, Decorative Elements, Emoji as Decoration, Fake Terminal Window, Generic AI Icons, Generic AI Typography, Illustrations With No Connection (+2 more)

### Community 84 - "antislop"
Cohesion: 0.25
Nodes (7): antislop, Core Principle, First-Run Install Wizard, Functional Patterns, Part 2: Mandatory Rules (R-01 to R-38, grouped), Two Usage Modes, What This Is (and What It Isn't)

### Community 85 - "Part 1: AI Slop Patterns (Warning Signs)"
Cohesion: 0.25
Nodes (8): Accessibility, Copywriting & Content, Decorative Elements, Functionality & Content, Identity & Originality, Layout & Components, Part 1: AI Slop Patterns (Warning Signs), Visual & Color

### Community 86 - "App & Dashboard"
Cohesion: 0.25
Nodes (8): App & Dashboard, Charts Without a Question, Default Dashboard Shell, Filler Activity Feed, Filler Data in Fields and Columns, Generic Table Columns, Placeholder Empty and Loading States, Stat Cards With Invented Numbers

### Community 87 - "antislop"
Cohesion: 0.25
Nodes (7): antislop, Core Principle, First-Run Install Wizard, Functional Patterns, Part 2: Mandatory Rules (R-01 to R-38, grouped), Two Usage Modes, What This Is (and What It Isn't)

### Community 88 - "Part 1: AI Slop Patterns (Warning Signs)"
Cohesion: 0.25
Nodes (8): Accessibility, Copywriting & Content, Decorative Elements, Functionality & Content, Identity & Originality, Layout & Components, Part 1: AI Slop Patterns (Warning Signs), Visual & Color

### Community 89 - "App & Dashboard"
Cohesion: 0.25
Nodes (8): App & Dashboard, Charts Without a Question, Default Dashboard Shell, Filler Activity Feed, Filler Data in Fields and Columns, Generic Table Columns, Placeholder Empty and Loading States, Stat Cards With Invented Numbers

### Community 90 - "antislop-ui"
Cohesion: 0.29
Nodes (6): antislop-ui, Endless Pulses and Loops, How to use this skill, Motion, Template Animations Stacked, UI Skill Checklist

### Community 91 - "antislop-ui"
Cohesion: 0.29
Nodes (6): antislop-ui, Endless Pulses and Loops, How to use this skill, Motion, Template Animations Stacked, UI Skill Checklist

### Community 92 - "The Craftsmanship Standard"
Cohesion: 0.33
Nodes (6): C-1 — Intentionality, C-2 — Functional Completeness, C-3 — Content-Driven Composition, C-4 — Resilience, C-5 — Evidence Over Claims, The Craftsmanship Standard

### Community 93 - "The Craftsmanship Standard"
Cohesion: 0.33
Nodes (6): C-1 — Intentionality, C-2 — Functional Completeness, C-3 — Content-Driven Composition, C-4 — Resilience, C-5 — Evidence Over Claims, The Craftsmanship Standard

### Community 94 - "test_leave.py"
Cohesion: 0.33
Nodes (3): prev_weekday(), Verifies the leave/permission/sick (cuti/izin/sakit) module: create, list, balan, Most recent date (<= d) with the given Python weekday (Mon=0..Sun=6).

### Community 95 - "Delivery Gate (Mandatory)"
Cohesion: 0.40
Nodes (5): Block 1: Hard Gate (absolute), Block 2: Purpose-Gate (technique allowed, reason required), Block 3: Liveliness (required to be alive, not just clean), Block 4: Craftsmanship & Quality Locks, Delivery Gate (Mandatory)

### Community 96 - "Delivery Gate (Mandatory)"
Cohesion: 0.40
Nodes (5): Block 1: Hard Gate (absolute), Block 2: Purpose-Gate (technique allowed, reason required), Block 3: Liveliness (required to be alive, not just clean), Block 4: Craftsmanship & Quality Locks, Delivery Gate (Mandatory)

### Community 97 - "Part 3: Liveliness Toolkit"
Cohesion: 0.50
Nodes (4): Design Read (how the dials are set), Levers (how the dials become visual decisions), Part 3: Liveliness Toolkit, Three Dials (required)

### Community 98 - "Structural & Flow"
Cohesion: 0.50
Nodes (4): Dead Navigation, Non-Functional Controls, Sections That Fill a Template, Structural & Flow

### Community 99 - "Part 3: Liveliness Toolkit"
Cohesion: 0.50
Nodes (4): Design Read (how the dials are set), Levers (how the dials become visual decisions), Part 3: Liveliness Toolkit, Three Dials (required)

### Community 100 - "Structural & Flow"
Cohesion: 0.50
Nodes (4): Dead Navigation, Non-Functional Controls, Sections That Fill a Template, Structural & Flow

### Community 103 - "liveness.py"
Cohesion: 0.25
Nodes (8): _crop(), LivenessDetector, ndarray, Path, Passive liveness (anti-spoof), MiniFASNet ensemble.  Mirrors the upstream infere, Real-face probability in [0, 1] for the face at bbox (x, y, w, h)., Upstream CropImage: expand the box by `scale` about its centre, clamped     to t, _softmax()

### Community 104 - "FaceEngine"
Cohesion: 0.25
Nodes (7): FaceEngine, load_models(), on_event, main(), ndarray, Measures FAR/FRR of the real ArcFace pipeline against docs/PRD.md section 3.  Ru, to_bgr()

### Community 105 - "main.py"
Cohesion: 0.25
Nodes (7): analyze(), healthz(), Internal face service. Binds to 127.0.0.1 only -- never exposed publicly (see do, get, JSONResponse, Request, post()

### Community 106 - "test_challenge.py"
Cohesion: 0.29
Nodes (5): Image, jpeg_bytes(), Verifies the movement challenge (docs/PRD.md F5) against the real stack.  The pr, Simulates a person moving slightly between frames., shifted()

### Community 107 - "test_all_pages.js"
Cohesion: 0.29
Nodes (4): check(), { chromium }, HR, visit()

### Community 108 - "benchmark_liveness.py"
Cohesion: 0.40
Nodes (4): main(), ndarray, Measures anti-spoof accuracy against docs/PRD.md section 3 (>95% spoof caught)., to_bgr()

### Community 109 - "seed_demo.py"
Cohesion: 0.40
Nodes (3): Fills a freshly-seeded database with realistic demo data so every screen has som, A date `offset_days` back from today, skipped back off weekends., workday()

### Community 110 - "ProfileModal.vue"
Cohesion: 0.15
Nodes (15): fetchProfile(), Profile, updateProfile(), auth, changingPassword, emit, form, load() (+7 more)

### Community 111 - "masterdata.ts"
Cohesion: 0.17
Nodes (11): createLocation(), createSchedule(), Device, listDevices(), Location, Schedule, updateLocation(), updateSchedule() (+3 more)

### Community 112 - "InlineMasterSelect.vue"
Cohesion: 0.20
Nodes (9): busy, cancel(), draft, emit, Mode, model, props, save() (+1 more)

### Community 113 - "RegisterRoutes"
Cohesion: 0.38
Nodes (10): createHandler(), HandlerFunc, IRoutes, Service, listHandler(), myListHandler(), RegisterRoutes(), reviewHandler() (+2 more)

### Community 114 - "10. Arah Desain Frontend"
Cohesion: 0.22
Nodes (9): 10.1 Tesis layar absen, 10.2 Sistem visual, 10.3 Layar absen — komposisi, 10.4 Layar admin, 10.5 State yang wajib ada (bukan opsional), 10.6 Motion, 10.7 Aksesibilitas & i18n, 10.8 Dependency frontend (+1 more)

### Community 115 - "7. Kebutuhan Fungsional"
Cohesion: 0.22
Nodes (9): 7. Kebutuhan Fungsional, F1 — Autentikasi & Otorisasi, F2 — Master Data, F3 — Face Enrollment, F4 — Absensi, F5 — Anti-Spoofing (wajib rilis 1), F6 — Dashboard & Laporan, F7 — Koreksi & Audit (+1 more)

### Community 116 - "load"
Cohesion: 0.31
Nodes (9): listDepartments(), listEmployees(), updateDepartment(), listLocations(), listSchedules(), loadAll(), reloadMasterLists(), load() (+1 more)

### Community 117 - "11. Deployment — VPS Tanpa Docker"
Cohesion: 0.29
Nodes (7): 11.1 Komponen, 11.2 Caddy, 11.3 systemd, 11.4 Deploy, 11.5 Backup & operasional, 11.6 Risiko VPS satu mesin, 11. Deployment — VPS Tanpa Docker

### Community 118 - "PRODUCT.md — HRIS Absensi Face Recognition"
Cohesion: 0.29
Nodes (7): Apa ini, Batasan yang mengikat, Bukan tujuan (rilis 1), Mekanisme unik, Pengguna & scene nyata, PRODUCT.md — HRIS Absensi Face Recognition, Yang bikin hasil poles terasa salah

### Community 119 - "Login.vue"
Cohesion: 0.29
Nodes (5): auth, errorMessage, form, router, submitting

### Community 120 - "9. Arsitektur"
Cohesion: 0.33
Nodes (6): 9. Arsitektur, API (garis besar), Kenapa 3 bahasa, Model, Pencocokan — ArcFace + Cosine Similarity (final), Skema data inti

### Community 121 - "HRIS Absensi — Face Recognition"
Cohesion: 0.33
Nodes (6): Coba sendiri (satu perintah), Deploy ke VPS, Development, HRIS Absensi — Face Recognition, Status, Struktur

### Community 124 - "attendance.ts"
Cohesion: 0.40
Nodes (3): CheckInError, CheckInErrorCode, CheckInResult

### Community 125 - "6. Alur Utama"
Cohesion: 0.50
Nodes (4): 6.1 Enrollment (pendaftaran wajah), 6.2 Check-in / Check-out, 6.3 Koreksi, 6. Alur Utama

## Knowledge Gaps
- **716 isolated node(s):** `github.com/hris-face/api`, `Service`, `loginRequest`, `createRequest`, `reviewRequest` (+711 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **17 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `RequireRole()` connect `main` to `leave/handler.go`, `report/handler.go`, `RegisterRoutes`, `employee/handler.go`, `RegisterRoutes`?**
  _High betweenness centrality (0.017) - this node is a cross-community bridge._
- **Why does `main()` connect `main` to `Service`?**
  _High betweenness centrality (0.013) - this node is a cross-community bridge._
- **What connects `github.com/hris-face/api`, `Service`, `loginRequest` to the rest of the system?**
  _716 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Enrollment.vue` be split into smaller, more focused modules?**
  _Cohesion score 0.09 - nodes in this community are weakly interconnected._
- **Should `dependencies` be split into smaller, more focused modules?**
  _Cohesion score 0.04878048780487805 - nodes in this community are weakly interconnected._
- **Should `CheckIn.vue` be split into smaller, more focused modules?**
  _Cohesion score 0.09113300492610837 - nodes in this community are weakly interconnected._
- **Should `MyHistory.vue` be split into smaller, more focused modules?**
  _Cohesion score 0.14705882352941177 - nodes in this community are weakly interconnected._