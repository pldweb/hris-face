// Opens every route in the app as both an admin and a regular employee and
// asserts each one actually renders something and raises no errors.
//
// This exists because a missing catch-all route made unknown paths render a
// literally blank page, and nothing in the suite would have caught it: the
// per-feature tests only ever visit paths they know exist.
//
//   node test/e2e/test_all_pages.js
//
// Needs the API, the Vite dev server, and `python3 test/e2e/seed_demo.py`.
const { chromium } = require('playwright')

const BASE = process.env.WEB_BASE ?? 'http://localhost:5173'
const HR = { email: 'hr@perusahaan.com', password: 'rahasia-sekali-123' }
const EMPLOYEE_EMAIL = process.env.EMPLOYEE_EMAIL ?? 'budi@perusahaan.com'
const EMPLOYEE_PASSWORD = process.env.EMPLOYEE_PASSWORD
const MANAGER_EMAIL = process.env.MANAGER_EMAIL ?? 'siti@perusahaan.com'
const MANAGER_PASSWORD = process.env.MANAGER_PASSWORD

let pass = 0
let fail = 0
function check(name, ok, detail = '') {
  if (ok) {
    pass++
    console.log(`  PASS  ${name}`)
  } else {
    fail++
    console.log(`  FAIL  ${name}  ${detail}`)
  }
}

// Noise that is not a page defect: a headless browser has no real camera, and
// MediaPipe's WASM runtime prints its own INFO lines through console.error.
const IGNORED_ERROR =
  /favicon|ResizeObserver|\[Vue warn\]: Slot|getUserMedia|NotAllowedError|NotFoundError|Permission denied|camera|TensorFlow|XNNPACK|^INFO:|delegate for CPU/i

function watch(page) {
  const errors = []
  page.on('pageerror', (e) => errors.push('pageerror: ' + e.message))
  page.on('console', (m) => {
    if (m.type() !== 'error') return
    const text = m.text()
    if (!IGNORED_ERROR.test(text)) errors.push('console: ' + text.slice(0, 160))
  })
  return errors
}

// Reading the shared nav must never hang the whole run if a page lacks it.
async function navOf(page) {
  return await page.locator('.employee-nav').innerText({ timeout: 5000 }).catch(() => '(nav tidak ditemukan)')
}

async function login(page, email, password) {
  await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' })
  await page.fill('input[type=email]', email)
  await page.fill('input[type=password]', password)
  await page.click('button:has-text("Masuk")')
  await page.waitForURL((u) => !u.pathname.startsWith('/login'), { timeout: 15000 })
}

// A rendered page must show real content, not an empty shell. 60 chars is
// comfortably below the smallest real screen and far above a blank body.
const MIN_BODY = 60

async function visit(page, errors, path, label, opts = {}) {
  errors.length = 0
  await page.goto(BASE + path, { waitUntil: 'networkidle' }).catch(() => {})
  await page.waitForTimeout(opts.settle ?? 900)

  const body = (await page.locator('body').innerText().catch(() => '')).trim()
  check(`${label} (${path}) terbuka, tidak blank`, body.length >= MIN_BODY,
    `hanya ${body.length} karakter di <body>`)
  check(`${label} tanpa error JS`, errors.length === 0, errors.slice(0, 2).join(' | '))

  if (opts.expectText) {
    check(`${label} menampilkan "${opts.expectText}"`, body.includes(opts.expectText),
      `body: ${JSON.stringify(body.slice(0, 120))}`)
  }
  if (opts.expectPath) {
    const got = new URL(page.url()).pathname
    check(`${label} berakhir di ${opts.expectPath}`, got === opts.expectPath, `got ${got}`)
  }
  return body
}

;(async () => {
  if (!EMPLOYEE_PASSWORD) {
    console.error('EMPLOYEE_PASSWORD belum diisi. Jalankan test/e2e/seed_demo.py dulu, ' +
      'lalu: EMPLOYEE_PASSWORD=<password> node test/e2e/test_all_pages.js')
    process.exit(2)
  }

  const browser = await chromium.launch({
    executablePath: '/usr/bin/google-chrome',
    args: ['--use-fake-device-for-media-stream', '--use-fake-ui-for-media-stream'],
  })

  // ---------- Public ----------
  console.log('\n=== HALAMAN PUBLIK ===')
  {
    const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } })
    const page = await ctx.newPage()
    const errors = watch(page)
    await visit(page, errors, '/login', 'Login', { expectText: 'Masuk' })
    // Deep link while logged out must land on login, never on a blank page.
    await visit(page, errors, '/admin/monitoring', 'Deep link tanpa sesi', { expectPath: '/login' })
    await ctx.close()
  }

  // ---------- Admin / HR ----------
  console.log('\n=== HALAMAN ADMIN (HR) ===')
  {
    const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } })
    const page = await ctx.newPage()
    const errors = watch(page)
    await login(page, HR.email, HR.password)

    await visit(page, errors, '/admin', 'Admin root (redirect)', { expectPath: '/admin/monitoring' })
    await visit(page, errors, '/admin/monitoring', 'Monitoring', { expectText: 'Kehadiran' })
    await visit(page, errors, '/admin/face-scan', 'Scan Wajah', { settle: 1600 })
    await visit(page, errors, '/admin/employees', 'Karyawan', { expectText: 'Budi Santoso' })
    await visit(page, errors, '/admin/corrections', 'Koreksi Absen', { expectText: 'Koreksi Absen' })
    await visit(page, errors, '/admin/leave-requests', 'Cuti & Izin', { expectText: 'Cuti & Izin' })
    await visit(page, errors, '/admin/master', 'Master Data', { expectText: 'Master Data' })

    // Sidebar navigation must work, not just direct URLs.
    console.log('\n=== NAVIGASI SIDEBAR ===')
    errors.length = 0
    await page.goto(`${BASE}/admin/monitoring`, { waitUntil: 'networkidle' })
    for (const [label, expectPath] of [
      ['Cuti & Izin', '/admin/leave-requests'],
      ['Koreksi Absen', '/admin/corrections'],
      ['Karyawan', '/admin/employees'],
      ['Master Data', '/admin/master'],
      ['Monitoring', '/admin/monitoring'],
    ]) {
      await page.locator(`.admin-sider a:has-text("${label}")`).first().click()
      await page.waitForTimeout(700)
      const got = new URL(page.url()).pathname
      const body = (await page.locator('body').innerText().catch(() => '')).trim()
      check(`klik sidebar "${label}" -> ${expectPath}, halaman terisi`,
        got === expectPath && body.length >= MIN_BODY, `url=${got} len=${body.length}`)
    }
    check('navigasi sidebar tanpa error JS', errors.length === 0, errors.slice(0, 2).join(' | '))

    // The approval queue must be usable, not just visible.
    console.log('\n=== INTERAKSI: SETUJUI CUTI ===')
    errors.length = 0
    await page.goto(`${BASE}/admin/leave-requests`, { waitUntil: 'networkidle' })
    await page.waitForTimeout(900)
    const approveBtn = page.locator('button:has-text("Setujui")').first()
    if (await approveBtn.count()) {
      await approveBtn.click()
      await page.waitForTimeout(500)
      const dialog = page.locator('.ant-modal-confirm')
      check('dialog konfirmasi persetujuan cuti muncul', await dialog.isVisible().catch(() => false))
      await dialog.locator('button:has-text("Setujui")').click()
      await page.waitForTimeout(1200)
      const after = await page.locator('body').innerText()
      check('cuti disetujui, tabel dimuat ulang tanpa error',
        !after.includes('Gagal') && errors.length === 0, errors.slice(0, 2).join(' | '))
    } else {
      check('ada pengajuan cuti pending untuk diuji', false, 'jalankan seed_demo.py lebih dulu')
    }
    await ctx.close()
  }

  // ---------- Employee ----------
  console.log('\n=== HALAMAN KARYAWAN ===')
  {
    const ctx = await browser.newContext({
      viewport: { width: 1440, height: 900 },
      permissions: ['camera'],
    })
    const page = await ctx.newPage()
    const errors = watch(page)
    await login(page, EMPLOYEE_EMAIL, EMPLOYEE_PASSWORD)

    await visit(page, errors, '/checkin', 'Absen', { settle: 1800 })
    await visit(page, errors, '/riwayat', 'Riwayat Saya', { expectText: 'Riwayat Absensi' })
    await visit(page, errors, '/cuti', 'Cuti', { expectText: 'Kuota tahunan' })
    await visit(page, errors, '/enroll', 'Daftar Wajah', { settle: 1800 })
    // Enrolment used to be a dead end: no nav, no cancel, browser-back only.
    check('Daftar Wajah punya jalan keluar (nav bersama)',
      (await navOf(page)).includes('Absen'))

    // The team screen is manager-only on the API, so a plain employee must
    // neither see the menu item nor be able to deep-link into it.
    await page.goto(`${BASE}/riwayat`, { waitUntil: 'networkidle' })
    const navText = await navOf(page)
    check('menu "Tim" disembunyikan untuk karyawan biasa', !navText.includes('Tim'), navText)
    await visit(page, errors, '/tim', 'Karyawan buka /tim', { expectPath: '/checkin' })

    console.log('\n=== INTERAKSI: FORM PENGAJUAN CUTI ===')
    errors.length = 0
    await page.goto(`${BASE}/cuti`, { waitUntil: 'networkidle' })
    await page.waitForTimeout(800)
    await page.locator('button:has-text("Ajukan")').first().click()
    await page.waitForTimeout(600)
    check('modal pengajuan cuti terbuka',
      await page.locator('.ant-modal').isVisible().catch(() => false))
    check('modal pengajuan cuti tanpa error JS', errors.length === 0, errors.slice(0, 2).join(' | '))

    // An employee must not be able to reach the admin area.
    console.log('\n=== BATAS AKSES KARYAWAN ===')
    await visit(page, errors, '/admin/employees', 'Karyawan buka /admin', { expectPath: '/checkin' })
    await ctx.close()
  }

  // ---------- Manager ----------
  if (MANAGER_EMAIL && MANAGER_PASSWORD) {
    console.log('\n=== HALAMAN MANAGER ===')
    const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } })
    const page = await ctx.newPage()
    const errors = watch(page)
    await login(page, MANAGER_EMAIL, MANAGER_PASSWORD)
    const navText = await navOf(page)
    check('menu "Tim" tampil untuk manager', navText.includes('Tim'), navText)
    await visit(page, errors, '/tim', 'Tim (manager)', { expectText: 'Kehadiran' })
    await ctx.close()
  } else {
    console.log('\n=== HALAMAN MANAGER (dilewati: MANAGER_EMAIL/PASSWORD belum diisi) ===')
  }

  // ---------- Unknown routes ----------
  console.log('\n=== ALAMAT TIDAK DIKENAL (404) ===')
  {
    const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } })
    const page = await ctx.newPage()
    const errors = watch(page)
    await login(page, HR.email, HR.password)
    for (const path of ['/halaman-typo', '/admin/tidak-ada', '/admin/leave-request']) {
      await visit(page, errors, path, 'Alamat salah', { expectText: 'Halaman tidak ditemukan' })
    }
    await ctx.close()
  }

  await browser.close()
  console.log(`\n${'='.repeat(52)}\nPASS: ${pass}   FAIL: ${fail}\n${'='.repeat(52)}`)
  process.exit(fail ? 1 : 0)
})()
