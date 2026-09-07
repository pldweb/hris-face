# Graph Report - hris-face  (2026-09-08)

## Corpus Check
- 96 files · ~46,425 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 851 nodes · 1309 edges · 59 communities (48 shown, 11 thin omitted)
- Extraction: 98% EXTRACTED · 2% INFERRED · 0% AMBIGUOUS · INFERRED: 23 edges (avg confidence: 0.72)
- Token cost: 0 input · 0 output

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
- RegisterRoutes
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
- test_challenge.py
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
- PRODUCT.md — HRIS Absensi Face Recognition
- End-to-end tests and accuracy benchmarks
- AGENTS.md
- web-src-views-checkin-vue.md
- web/README.md
- report.ts
- tzDayjs
- client.ts
- main.ts
- EmployeeNav.vue
- Login.vue
- test_admin_ui.js
- test_admin_crud.py

## God Nodes (most connected - your core abstractions)
1. `Service` - 19 edges
2. `PRD — HRIS Absensi Berbasis Face Recognition` - 16 edges
3. `compilerOptions` - 15 edges
4. `RegisterRoutes()` - 14 edges
5. `RegisterRoutes()` - 13 edges
6. `Service` - 12 edges
7. `RegisterRoutes()` - 12 edges
8. `Service` - 12 edges
9. `employees` - 11 edges
10. `Service` - 11 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `NewPhotoStore()`  [INFERRED]
  api/cmd/api/main.go → api/internal/attendance/photostore.go
- `main()` --calls--> `Migrate()`  [INFERRED]
  api/cmd/api/main.go → api/internal/db/migrate.go
- `main()` --calls--> `RequireAuth()`  [INFERRED]
  api/cmd/api/main.go → api/internal/middleware/auth.go
- `RegisterScanRoute()` --calls--> `RequireRole()`  [INFERRED]
  api/internal/attendance/handler.go → api/internal/middleware/auth.go
- `RegisterRoutes()` --calls--> `RequireRole()`  [INFERRED]
  api/internal/employee/handler.go → api/internal/middleware/auth.go

## Import Cycles
- None detected.

## Communities (59 total, 11 thin omitted)

### Community 0 - "Enrollment.vue"
Cohesion: 0.06
Nodes (33): submitEnrollment(), FaceScanResult, scanFace(), cameraError, canScan, canvasEl, captureJpeg(), onScan() (+25 more)

### Community 1 - "dependencies"
Cohesion: 0.05
Nodes (40): @ant-design/icons-vue, ant-design-vue, axios, dayjs, @mediapipe/tasks-vision, pinia, @types/node, typescript (+32 more)

### Community 2 - "face_engine.py"
Cohesion: 0.08
Nodes (29): eye_openness(), face_signature(), ndarray, Eye openness, used for the blink challenge (docs/PRD.md F5).  Passive anti-spoof, Mean vertical eye opening as a fraction of face width.      Normalising by face, An 8x8 grayscale thumbnail of the face crop.      The blink challenge needs to p, DetectedFace, FaceEngine (+21 more)

### Community 3 - "CheckIn.vue"
Cohesion: 0.09
Nodes (27): submitAttendance(), submitChallenge(), fetchMe(), actionLabel, CameraError, canvasEl, captureBurst(), captureJpeg() (+19 more)

### Community 4 - "MyHistory.vue"
Cohesion: 0.15
Nodes (15): createCorrection(), listMyCorrections(), fetchMyHistory(), corrections, days, form, load(), loading (+7 more)

### Community 5 - "EmployeeList.vue"
Cohesion: 0.07
Nodes (40): createDepartment(), createEmployee(), CreateEmployeeInput, CreateEmployeeResult, deactivateEmployee(), Department, Employee, importEmployeesCsv() (+32 more)

### Community 6 - "AdminLayout.vue"
Cohesion: 0.18
Nodes (8): refreshAccessToken(), useAuthStore, auth, initials, pageTitle, route, router, selectedKeys

### Community 7 - "Service"
Cohesion: 0.10
Nodes (22): Service, Context, meanPairwiseVariation(), Duration, Time, NewPhotoStore(), TestPhotoStoreDisabled(), TestPhotoStoreRetention() (+14 more)

### Community 8 - "MasterData.vue"
Cohesion: 0.09
Nodes (32): approveDevice(), createLocation(), createSchedule(), deleteLocation(), Device, listDevices(), listSchedules(), Location (+24 more)

### Community 9 - "main"
Cohesion: 0.09
Nodes (29): main(), main(), mustEnv(), challengeHandler(), HandlerFunc, IRoutes, Service, markHandler() (+21 more)

### Community 10 - "Monitoring.vue"
Cohesion: 0.09
Nodes (24): exportReport(), applied, columns, defaultRange(), deletingID, departments, draft, editingID (+16 more)

### Community 11 - "Service"
Cohesion: 0.17
Nodes (13): Context, Service, Pool, Time, NewService(), randomPassword(), CreateEmployeeInput, CreateEmployeeResult (+5 more)

### Community 12 - "compilerOptions"
Cohesion: 0.10
Nodes (19): ES2023, node, vite.config.ts, compilerOptions, allowImportingTsExtensions, erasableSyntaxOnly, lib, module (+11 more)

### Community 13 - "Service"
Cohesion: 0.21
Nodes (9): Context, Pool, Time, NewService(), New(), Correction, Service, Config (+1 more)

### Community 14 - "RegisterRoutes"
Cohesion: 0.21
Nodes (15): createHandler(), HandlerFunc, IRoutes, Service, listHandler(), myListHandler(), RegisterRoutes(), reviewHandler() (+7 more)

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
Cohesion: 0.30
Nodes (17): approveDeviceHandler(), createDepartmentHandler(), createHandler(), deactivateHandler(), HandlerFunc, IRoutes, Service, importHandler() (+9 more)

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

### Community 25 - "test_challenge.py"
Cohesion: 0.18
Nodes (9): analyze(), Image, JSONResponse, Request, jpeg_bytes(), post(), Verifies the movement challenge (docs/PRD.md F5) against the real stack.  The pr, Simulates a person moving slightly between frames. (+1 more)

### Community 26 - "e2e_api.py"
Cohesion: 0.17
Nodes (5): prepare_faces(), End-to-end test against the real Go API + real Postgres/pgvector.  Runs against, Like request() but returns the raw body -- the CSV export is not JSON., Two LFW identities (two photos each) plus a third as the unknown face.      Enro, request_raw()

### Community 27 - "PRD — HRIS Absensi Berbasis Face Recognition"
Cohesion: 0.04
Nodes (47): 10.1 Tesis layar absen, 10.2 Sistem visual, 10.3 Layar absen — komposisi, 10.4 Layar admin, 10.5 State yang wajib ada (bukan opsional), 10.6 Motion, 10.7 Aksesibilitas & i18n, 10.8 Dependency frontend (+39 more)

### Community 28 - "throttle"
Cohesion: 0.29
Nodes (6): Duration, Time, newThrottle(), attemptRecord, throttle, Mutex

### Community 29 - "employees"
Cohesion: 0.40
Nodes (10): attendance_corrections, attendances, audit_logs, departments, devices, employees, face_embeddings, users (+2 more)

### Community 30 - ".ImportCSV"
Cohesion: 0.25
Nodes (8): Context, Service, mapHeader(), ImportResult, ImportRow, healthz(), get, Reader

### Community 31 - "run-local.sh"
Cohesion: 0.25
Nodes (6): DATABASE_URL, FACE_SERVICE_URL, JWT_SECRET, LISTEN_ADDR, OFFICE_IP_ALLOWLIST, run-local.sh script

### Community 32 - "RegisterRoutes"
Cohesion: 0.47
Nodes (5): HandlerFunc, IRoutes, Service, RegisterRoutes(), uploadHandler()

### Community 34 - "export_antispoof_onnx.py"
Cohesion: 0.50
Nodes (4): fetch(), main(), Path, Exports the official Silent-Face-Anti-Spoofing weights to ONNX.  Provenance matt

### Community 46 - "PRODUCT.md — HRIS Absensi Face Recognition"
Cohesion: 0.09
Nodes (19): Accuracy, Anti-spoof models, Face service, Run, Setup, Two things that will silently break this model, Apa ini, Batasan yang mengikat (+11 more)

### Community 47 - "End-to-end tests and accuracy benchmarks"
Cohesion: 0.40
Nodes (4): End-to-end tests and accuracy benchmarks, Running, Test identities, What these tests still do not cover

### Community 51 - "report.ts"
Cohesion: 0.21
Nodes (11): AttendanceFilter, AttendanceRow, DailySummary, DayRecord, deleteAttendance(), fetchToday(), listAttendances(), updateAttendance() (+3 more)

### Community 52 - "tzDayjs"
Cohesion: 0.33
Nodes (5): Dayjs, DayjsArgs, tzDayjs(), onEditOpen(), openCorrection()

### Community 53 - "client.ts"
Cohesion: 0.24
Nodes (6): CheckInError, CheckInErrorCode, CheckInResult, apiClient, EnrollOutcome, PhotoRejection

### Community 54 - "main.ts"
Cohesion: 0.27
Nodes (4): i18n, router, attendanceStatusColor, themeTokens

### Community 55 - "EmployeeNav.vue"
Cohesion: 0.22
Nodes (6): auth, initials, isAdmin, links, route, router

### Community 56 - "Login.vue"
Cohesion: 0.29
Nodes (5): auth, errorMessage, form, router, submitting

## Knowledge Gaps
- **273 isolated node(s):** `github.com/hris-face/api`, `Service`, `loginRequest`, `createRequest`, `reviewRequest` (+268 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **11 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `RequireRole()` connect `RegisterRoutes` to `report/handler.go`, `main`, `employee/handler.go`, `RegisterRoutes`?**
  _High betweenness centrality (0.031) - this node is a cross-community bridge._
- **Why does `main()` connect `main` to `RegisterRoutes`, `Service`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._
- **Why does `RegisterScanRoute()` connect `main` to `RegisterRoutes`?**
  _High betweenness centrality (0.025) - this node is a cross-community bridge._
- **What connects `github.com/hris-face/api`, `Service`, `loginRequest` to the rest of the system?**
  _273 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Enrollment.vue` be split into smaller, more focused modules?**
  _Cohesion score 0.06025641025641026 - nodes in this community are weakly interconnected._
- **Should `dependencies` be split into smaller, more focused modules?**
  _Cohesion score 0.04878048780487805 - nodes in this community are weakly interconnected._
- **Should `face_engine.py` be split into smaller, more focused modules?**
  _Cohesion score 0.07557354925775979 - nodes in this community are weakly interconnected._