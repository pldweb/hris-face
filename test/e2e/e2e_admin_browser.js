// Browser coverage for the screens added after the first release slice:
// HR monitoring + CSV export, corrections approval, master data, and the
// employee's own history/correction flow.
const { chromium } = require('playwright')

const SCRATCH = '/tmp/claude-1000/-home-erahajj-projek-hris-face/1d9b5f7c-df64-4781-a725-f5845c468e14/scratchpad'
const SHOTS = '/tmp/shots'
const BASE = 'http://localhost:5173'
const HR = { email: 'hr@perusahaan.com', password: 'rahasia-sekali-123' }

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

// Creates its own employee (and enrols them through the UI) so this suite does
// not depend on whatever e2e_api.py happened to leave behind.
async function createEmployee(page, nik, name, email) {
  return page.evaluate(
    async ({ nik, name, email, token }) => {
      const res = await fetch('/api/v1/admin/employees', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ nik, full_name: name, email }),
      })
      return res.json()
    },
    { nik, name, email, token: await page.evaluate(() => sessionStorage.getItem('access_token')) },
  )
}

async function login(page, email, password) {
  await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' })
  await page.fill('input[type=email]', email)
  await page.fill('input[type=password]', password)
  await page.click('button:has-text("Masuk")')
  await page.waitForURL((u) => !u.pathname.startsWith('/login'), { timeout: 10000 })
}

;(async () => {
  const browser = await chromium.launch({
    executablePath: '/usr/bin/google-chrome',
    args: [
      '--use-fake-ui-for-media-stream',
      '--use-fake-device-for-media-stream',
      `--use-file-for-fake-video-capture=${SCRATCH}/fakecam/realA1.y4m`,
    ],
  })
  const errors = []

  // ---------- HR: monitoring ----------
  console.log('\n=== HR: MONITORING ===')
  const hrCtx = await browser.newContext({ viewport: { width: 1440, height: 900 }, acceptDownloads: true })
  const page = await hrCtx.newPage()
  page.on('pageerror', (e) => errors.push(e.message))

  await login(page, HR.email, HR.password)

  const stamp = Date.now()
  const employee = { nik: `T-${stamp}`, name: `Tester ${stamp}`, email: `tester${stamp}@perusahaan.com`, password: '' }
  const created = await createEmployee(page, employee.nik, employee.name, employee.email)
  employee.password = created.temp_password
  check('setup: karyawan uji dibuat', !!employee.password, JSON.stringify(created).slice(0, 150))

  await page.goto(`${BASE}/admin/monitoring`, { waitUntil: 'networkidle' })
  await page.waitForSelector('.stat-card', { timeout: 10000 })

  const statTitles = await page.locator('.stat-title').allInnerTexts()
  check('empat kartu statistik tampil', statTitles.length === 4, `got ${JSON.stringify(statTitles)}`)
  check('kartu "Belum absen" ada', statTitles.some((t) => t.includes('Belum absen')), `got ${statTitles}`)

  const headerText = await page.locator('.ant-table-thead').innerText()
  check('tabel absensi punya kolom yang benar',
    headerText.includes('Nama') && headerText.includes('Status') && headerText.includes('Tipe'),
    `got "${headerText}"`)
  await page.screenshot({ path: `${SHOTS}/admin-monitoring.png` })

  // Filter by status is now a two-step apply, not auto-reload on change: pick
  // the option, then press "Terapkan Filter" for the table to actually narrow.
  await page.locator('.filters .ant-select').last().click()
  await page.locator('.ant-select-item-option:has-text("Terlambat")').first().click()
  await page.click('button:has-text("Terapkan Filter")')
  await page.waitForTimeout(1200)
  const filtered = await page.locator('.ant-table-tbody').innerText()
  check(
    'filter status menyaring tabel',
    !filtered.includes('Tepat waktu') || filtered.includes('Tidak ada data'),
    `got "${filtered.slice(0, 120)}"`,
  )

  await page.click('button:has-text("Reset")')
  await page.waitForTimeout(1000)

  // ---------- HR: CSV export ----------
  console.log('\n=== HR: EXPORT ===')
  const downloadPromise = page.waitForEvent('download', { timeout: 15000 })
  await page.click('button:has-text("Export CSV")')
  const download = await downloadPromise
  check('klik Export CSV menghasilkan unduhan', !!download)
  check('nama file CSV benar', download.suggestedFilename().endsWith('.csv'), download.suggestedFilename())

  // ---------- HR: master data ----------
  console.log('\n=== HR: MASTER DATA ===')
  await page.goto(`${BASE}/admin/master`, { waitUntil: 'networkidle' })
  // Wait for the schedule rows themselves: the card renders before its data
  // arrives, so asserting on the card alone reads an empty table.
  await page.waitForSelector('.ant-table-row', { timeout: 10000 })
  const masterText = await page.locator('body').innerText()
  check('jadwal default tampil', masterText.includes('Reguler 08:00-17:00'), `got "${masterText.slice(0, 200)}"`)
  check('bagian perangkat tampil', masterText.includes('Perangkat'), 'section missing')

  await page.click('button:has-text("Tambah Jadwal")')
  await page.waitForSelector('.ant-modal', { state: 'visible' })
  const modal = page.locator('.ant-modal')
  await modal.locator('input').nth(0).fill('Shift Malam')
  await modal.locator('input').nth(1).fill('22:00')
  await modal.locator('input').nth(2).fill('06:00')
  await page.click('.ant-modal-footer button.ant-btn-primary')
  await page.waitForTimeout(1500)
  check('jadwal baru muncul di daftar', (await page.locator('body').innerText()).includes('Shift Malam'))
  await page.screenshot({ path: `${SHOTS}/admin-master.png` })

  // ---------- HR: corrections queue (empty first) ----------
  await page.goto(`${BASE}/admin/corrections`, { waitUntil: 'networkidle' })
  await page.waitForTimeout(1000)
  check(
    'antrean koreksi kosong di awal',
    (await page.locator('body').innerText()).includes('Tidak ada pengajuan'),
    'expected empty state',
  )
  await hrCtx.close()

  // ---------- Employee: history + file a correction ----------
  console.log('\n=== KARYAWAN: RIWAYAT & KOREKSI ===')
  const empCtx = await browser.newContext({ viewport: { width: 1440, height: 900 } })
  const emp = await empCtx.newPage()
  emp.on('pageerror', (e) => errors.push(e.message))

  await login(emp, employee.email, employee.password)

  // Enrol and check in so the history screen has something real to show.
  await emp.goto(`${BASE}/enroll`, { waitUntil: 'networkidle' })
  const captureBtn = emp.locator('button:has-text("Ambil Foto")')
  await captureBtn.waitFor({ state: 'visible', timeout: 15000 })
  for (let i = 0; i < 40 && (await captureBtn.isDisabled()); i++) await emp.waitForTimeout(300)
  for (let i = 0; i < 3; i++) {
    await captureBtn.click()
    await emp.waitForTimeout(500)
  }
  await emp.locator('button:has-text("Kirim")').click()
  await emp.waitForURL('**/checkin', { timeout: 20000 })
  await emp.waitForTimeout(5000) // auto check-in fires on a stable face

  await emp.goto(`${BASE}/riwayat`, { waitUntil: 'networkidle' })
  await emp.waitForSelector('.ant-table-tbody', { timeout: 10000 })

  const historyText = await emp.locator('.ant-table-tbody').innerText()
  check('riwayat menampilkan hari-hari bulan ini', historyText.split('\n').length > 3, `got ${historyText.slice(0, 80)}`)
  check('riwayat memuat jam absen hari ini', /\d{2}:\d{2}/.test(historyText), `got ${historyText.slice(0, 120)}`)
  await emp.screenshot({ path: `${SHOTS}/employee-history.png` })

  await emp.click('button:has-text("Ajukan Koreksi")')
  await emp.waitForSelector('.ant-modal', { state: 'visible' })
  await emp.locator('.ant-modal textarea').fill('Lupa absen pulang karena rapat sampai malam')
  await emp.click('.ant-modal-footer button.ant-btn-primary')
  await emp.waitForTimeout(1500)
  const afterSubmit = await emp.locator('body').innerText()
  check('pengajuan koreksi tercatat di layar karyawan', afterSubmit.includes('Menunggu'), `got ${afterSubmit.slice(0, 200)}`)
  await empCtx.close()

  // ---------- HR: approve it ----------
  console.log('\n=== HR: SETUJUI KOREKSI ===')
  const hr2 = await browser.newContext({ viewport: { width: 1440, height: 900 } })
  const hrPage = await hr2.newPage()
  hrPage.on('pageerror', (e) => errors.push(e.message))
  await login(hrPage, HR.email, HR.password)
  await hrPage.goto(`${BASE}/admin/corrections`, { waitUntil: 'networkidle' })
  await hrPage.waitForSelector('.ant-table-tbody', { timeout: 10000 })

  const queueText = await hrPage.locator('.ant-table-tbody').innerText()
  check('pengajuan karyawan muncul di antrean HR', queueText.includes(employee.name), `got ${queueText.slice(0, 150)}`)
  await hrPage.screenshot({ path: `${SHOTS}/admin-corrections.png` })

  await hrPage.click('button:has-text("Setujui")')
  await hrPage.waitForSelector('.ant-modal-confirm', { state: 'visible' })
  const confirmText = await hrPage.locator('.ant-modal-confirm').innerText()
  check('konfirmasi menjelaskan akibatnya', confirmText.includes('koreksi manual'), `got ${confirmText.slice(0, 150)}`)
  await hrPage.click('.ant-modal-confirm .ant-btn-primary')
  await hrPage.waitForTimeout(1800)

  await hrPage.goto(`${BASE}/admin/monitoring`, { waitUntil: 'networkidle' })
  await hrPage.waitForTimeout(1200)
  const monitorAfter = await hrPage.locator('.ant-table-tbody').innerText()
  check('record hasil koreksi ditandai manual di monitoring', monitorAfter.includes('manual'), `got ${monitorAfter.slice(0, 200)}`)
  await hr2.close()

  check('tidak ada JS error di seluruh alur', errors.length === 0, errors.join(' | '))

  await browser.close()
  console.log(`\n${'='.repeat(46)}\nPASS: ${pass}   FAIL: ${fail}\n${'='.repeat(46)}`)
  process.exit(fail ? 1 : 0)
})()
