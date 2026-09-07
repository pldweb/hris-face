// Full browser end-to-end: real Chrome, real Vue app, real Go API, real Postgres.
// Only the ML model is stubbed. Drives the actual UI the way an employee would.
const { chromium } = require('playwright')

const SCRATCH = '/tmp/claude-1000/-home-erahajj-projek-hris-face/1d9b5f7c-df64-4781-a725-f5845c468e14/scratchpad'
const SHOTS = '/tmp/shots'
const BASE = 'http://localhost:5173'

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

async function launch(videoFile) {
  return chromium.launch({
    executablePath: '/usr/bin/google-chrome',
    args: [
      '--use-fake-ui-for-media-stream',
      '--use-fake-device-for-media-stream',
      `--use-file-for-fake-video-capture=${SCRATCH}/fakecam/${videoFile}`,
    ],
  })
}

;(async () => {
  // ---------- HR: create an employee through the admin UI ----------
  console.log('\n=== ADMIN UI: buat karyawan ===')
  let browser = await launch('realA1.y4m')
  let page = await browser.newPage({ viewport: { width: 1440, height: 900 } })
  const errors = []
  page.on('pageerror', (e) => errors.push(e.message))

  await page.goto(`${BASE}/admin/employees`, { waitUntil: 'networkidle' })
  check('akses admin tanpa login diarahkan ke /login', page.url().includes('/login'), page.url())

  await page.fill('input[type=email]', 'hr@perusahaan.com')
  await page.fill('input[type=password]', 'rahasia-sekali-123')
  await page.click('button:has-text("Masuk")')
  // The guard stored ?redirect=/admin/employees, so login should return there.
  await page.waitForURL((u) => !u.pathname.startsWith('/login'), { timeout: 10000 })
  check(
    'login HR mengembalikan ke halaman tujuan semula',
    page.url().includes('/admin/employees'),
    page.url(),
  )

  await page.goto(`${BASE}/admin/employees`, { waitUntil: 'networkidle' })
  await page.waitForSelector('button:has-text("Tambah Karyawan")')
  await page.click('button:has-text("Tambah Karyawan")')
  await page.waitForSelector('.ant-modal', { state: 'visible' })
  check('modal tambah karyawan terbuka', true)

  const modal = page.locator('.ant-modal')
  await modal.locator('input').nth(0).fill('E-100')
  await modal.locator('input').nth(1).fill('Siti Rahayu')
  await modal.locator('input').nth(2).fill('siti@perusahaan.com')
  await page.screenshot({ path: `${SHOTS}/e2e-1-form.png` })
  await page.click('.ant-modal-footer button.ant-btn-primary')

  await page.waitForSelector('.ant-typography', { timeout: 10000 })
  const tempPassword = (await page.locator('.temp-password').innerText()).trim()
  check('password sementara ditampilkan setelah dibuat', tempPassword.length > 5, `got "${tempPassword}"`)
  await page.screenshot({ path: `${SHOTS}/e2e-2-temp-password.png` })

  await page.keyboard.press('Escape')
  await page.waitForTimeout(400)
  const rowText = await page.locator('.ant-table-tbody').innerText()
  check('karyawan baru muncul di tabel', rowText.includes('Siti Rahayu'), rowText)
  check('status awal "Belum enrollment"', rowText.includes('Belum enrollment'), rowText)
  await page.screenshot({ path: `${SHOTS}/e2e-3-employee-list.png` })
  await browser.close()

  // ---------- Employee: enrollment through the UI ----------
  console.log('\n=== KARYAWAN UI: enrollment ===')
  browser = await launch('realA1.y4m')
  page = await browser.newPage({ viewport: { width: 1440, height: 900 } })
  page.on('pageerror', (e) => errors.push(e.message))

  await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' })
  await page.fill('input[type=email]', 'siti@perusahaan.com')
  await page.fill('input[type=password]', tempPassword)
  await page.click('button:has-text("Masuk")')
  await page.waitForURL('**/checkin', { timeout: 10000 })
  check('karyawan login dengan password sementara', true)

  await page.goto(`${BASE}/enroll`, { waitUntil: 'networkidle' })
  const captureBtn = page.locator('button:has-text("Ambil Foto")')
  await captureBtn.waitFor({ state: 'visible' })
  for (let i = 0; i < 30 && (await captureBtn.isDisabled()); i++) await page.waitForTimeout(300)
  check('tombol Ambil Foto aktif setelah wajah terdeteksi', !(await captureBtn.isDisabled()))

  const submitBtn = page.locator('button:has-text("Kirim")')
  check('tombol Kirim terkunci sebelum 3 foto', await submitBtn.isDisabled())

  for (let i = 0; i < 3; i++) {
    await captureBtn.click()
    await page.waitForTimeout(500)
  }
  const thumbs = await page.locator('.thumb').count()
  check('3 foto ter-capture dan tampil sebagai thumbnail', thumbs === 3, `got ${thumbs}`)
  await page.screenshot({ path: `${SHOTS}/e2e-4-enrollment.png` })

  check('tombol Kirim aktif setelah 3 foto', !(await submitBtn.isDisabled()))
  await submitBtn.click()
  await page.waitForURL('**/checkin', { timeout: 15000 })
  check('enrollment sukses dan diarahkan ke check-in', page.url().includes('/checkin'), page.url())

  // ---------- Employee: check-in ----------
  console.log('\n=== KARYAWAN UI: check-in ===')
  await page.waitForTimeout(3000)
  const statusText = await page.locator('.status-text').innerText()
  await page.screenshot({ path: `${SHOTS}/e2e-5-checkin-result.png` })
  check(
    'check-in otomatis berhasil dan menampilkan nama + jam',
    statusText.includes('Siti Rahayu') && /\d{2}:\d{2}/.test(statusText),
    `status="${statusText}"`,
  )
  check('status kehadiran ditampilkan', /Tepat waktu|Terlambat/.test(statusText), `status="${statusText}"`)

  const greeting = await page.locator('.greeting').innerText()
  check('sapaan memakai nama karyawan, bukan "Karyawan"', greeting.includes('Siti Rahayu'), `greeting="${greeting}"`)

  const summary = await page.locator('.summary-row').innerText()
  check('ringkasan menampilkan jam masuk sungguhan', /Masuk \d{2}:\d{2}/.test(summary), `summary="${summary}"`)
  await browser.close()

  // ---------- Check-out in the UI ----------
  console.log('\n=== KARYAWAN UI: absen pulang ===')
  browser = await launch('realA1.y4m')
  page = await browser.newPage({ viewport: { width: 1440, height: 900 } })
  page.on('pageerror', (e) => errors.push(e.message))
  await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' })
  await page.fill('input[type=email]', 'siti@perusahaan.com')
  await page.fill('input[type=password]', tempPassword)
  await page.click('button:has-text("Masuk")')
  await page.waitForURL('**/checkin', { timeout: 10000 })

  const outBtn = page.locator('button:has-text("Absen Pulang")')
  await outBtn.waitFor({ state: 'visible', timeout: 10000 })
  check('tombol berubah jadi "Absen Pulang" setelah masuk', true)

  for (let i = 0; i < 30 && (await outBtn.isDisabled()); i++) await page.waitForTimeout(300)
  await outBtn.click()
  await page.waitForTimeout(5000)
  const outSummary = await page.locator('.summary-row').innerText()
  await page.screenshot({ path: `${SHOTS}/e2e-7-checkout.png` })
  check('ringkasan menampilkan jam pulang', /Pulang \d{2}:\d{2}/.test(outSummary), `summary="${outSummary}"`)
  check('durasi kerja terhitung', /Durasi \d+j \d+m/.test(outSummary), `summary="${outSummary}"`)
  await browser.close()

  // ---------- Nothing left to do ----------
  console.log('\n=== ATURAN BISNIS DI UI ===')
  browser = await launch('realA1.y4m')
  page = await browser.newPage({ viewport: { width: 1440, height: 900 } })
  page.on('pageerror', (e) => errors.push(e.message))
  await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' })
  await page.fill('input[type=email]', 'siti@perusahaan.com')
  await page.fill('input[type=password]', tempPassword)
  await page.click('button:has-text("Masuk")')
  await page.waitForURL('**/checkin', { timeout: 10000 })
  await page.waitForTimeout(4000)
  await page.screenshot({ path: `${SHOTS}/e2e-6-done.png` })
  const doneText = await page.locator('.all-done').innerText()
  check('layar menyatakan absen hari ini sudah lengkap', doneText.includes('lengkap'), `text="${doneText}"`)

  const dupSummary = await page.locator('.summary-row').innerText()
  check(
    'ringkasan menampilkan masuk, pulang, dan durasi',
    /Masuk \d{2}:\d{2}/.test(dupSummary) && /Pulang \d{2}:\d{2}/.test(dupSummary),
    `summary="${dupSummary}"`,
  )
  await browser.close()

  check('tidak ada JS error di seluruh alur', errors.length === 0, errors.join(' | '))

  console.log(`\n${'='.repeat(46)}\nPASS: ${pass}   FAIL: ${fail}\n${'='.repeat(46)}`)
  process.exit(fail ? 1 : 0)
})()
