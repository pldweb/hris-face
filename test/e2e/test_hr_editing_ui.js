// Drives the HR editing features through the real UI: the full employee form
// (including inline departemen/lokasi creation and a password reset), editing
// a leave request, and editing your own profile from the header.
//
//   EMPLOYEE_PASSWORD=<pw> node test/e2e/test_hr_editing_ui.js
//
// Needs the API, the Vite dev server, and test/e2e/seed_demo.py.
const { chromium } = require('playwright')

const BASE = process.env.WEB_BASE ?? 'http://localhost:5173'
const HR = { email: 'hr@perusahaan.com', password: 'rahasia-sekali-123' }
const EMPLOYEE_EMAIL = process.env.EMPLOYEE_EMAIL ?? 'budi@perusahaan.com'
const EMPLOYEE_PASSWORD = process.env.EMPLOYEE_PASSWORD

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

async function login(page, email, password) {
  await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' })
  await page.fill('input[type=email]', email)
  await page.fill('input[type=password]', password)
  await page.click('button:has-text("Masuk")')
  await page.waitForURL((u) => !u.pathname.startsWith('/login'), { timeout: 15000 })
}

;(async () => {
  if (!EMPLOYEE_PASSWORD) {
    console.error('EMPLOYEE_PASSWORD belum diisi (lihat output seed_demo.py)')
    process.exit(2)
  }

  const browser = await chromium.launch({ executablePath: '/usr/bin/google-chrome' })
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 1000 } })
  const page = await ctx.newPage()
  const errors = []
  page.on('pageerror', (e) => errors.push('pageerror: ' + e.message))
  page.on('console', (m) => {
    if (m.type() === 'error' && !/favicon|ResizeObserver/i.test(m.text())) errors.push('console: ' + m.text().slice(0, 140))
  })

  const stamp = Date.now()
  await login(page, HR.email, HR.password)

  // ---------- Employee form: all fields + inline master data ----------
  console.log('\n=== FORM KARYAWAN: SEMUA FIELD ===')
  await page.goto(`${BASE}/admin/employees`, { waitUntil: 'networkidle' })
  await page.waitForSelector('.ant-table-row', { timeout: 10000 })

  const row = page.locator('tr', { hasText: 'Budi Santoso' })
  await row.locator('button:has-text("Ubah")').click()
  await page.waitForSelector('.ant-modal:has-text("Ubah Karyawan")', { state: 'visible' })
  const modal = page.locator('.ant-modal:has-text("Ubah Karyawan")')

  const modalText = await modal.innerText()
  for (const label of ['NIK', 'Jadwal kerja', 'Atasan', 'Status', 'Kuota cuti tahunan', 'Reset password karyawan']) {
    check(`form edit punya field "${label}"`, modalText.includes(label))
  }
  check('NIK bisa diedit (tidak lagi terkunci)',
    await modal.locator('input').first().isEnabled())

  // Inline "tambah departemen" without leaving the form.
  console.log('\n=== TAMBAH DEPARTEMEN LANGSUNG DARI FORM ===')
  const deptItem = modal.locator('.ant-form-item:has-text("Departemen")').first()
  await deptItem.locator('button').first().click() // the + button
  await page.waitForTimeout(400)
  const deptInput = deptItem.locator('input')
  check('form berubah jadi input departemen baru', await deptInput.isVisible().catch(() => false))
  await deptInput.fill(`Divisi UI ${stamp}`)
  await deptItem.locator('button.ant-btn-primary').click()
  await page.waitForTimeout(1200)
  const deptAfter = await deptItem.innerText()
  check('departemen baru langsung terpilih di form', deptAfter.includes(`Divisi UI ${stamp}`), deptAfter)

  // Inline "tambah lokasi".
  const locItem = modal.locator('.ant-form-item:has-text("Lokasi kerja")').first()
  await locItem.locator('button').first().click()
  await page.waitForTimeout(400)
  await locItem.locator('input').fill(`Lokasi UI ${stamp}`)
  await locItem.locator('button.ant-btn-primary').click()
  await page.waitForTimeout(1200)
  check('lokasi baru langsung terpilih di form',
    (await locItem.innerText()).includes(`Lokasi UI ${stamp}`))

  // Quota + password reset in the same save.
  console.log('\n=== UBAH KUOTA & RESET PASSWORD ===')
  const quota = modal.locator('.ant-form-item:has-text("Kuota cuti") input')
  await quota.fill('')
  await quota.type('18')
  await modal.locator('text=Reset password karyawan').click()
  await page.waitForTimeout(300)
  await modal.locator('input[type=password]').fill('PasswordUjiUI123')
  await page.locator('.ant-modal-footer button.ant-btn-primary').click()
  await page.waitForTimeout(1600)
  check('simpan perubahan karyawan berhasil',
    !(await page.locator('body').innerText()).includes('Gagal memperbarui'))

  // The reset must actually take effect.
  const ctx2 = await browser.newContext()
  const page2 = await ctx2.newPage()
  await login(page2, EMPLOYEE_EMAIL, 'PasswordUjiUI123')
  check('karyawan bisa login dengan password hasil reset HR',
    !new URL(page2.url()).pathname.startsWith('/login'), page2.url())
  await ctx2.close()

  // Quota change should show on the employee's leave screen; verified via the
  // admin list instead, which is where HR would look.
  await page.goto(`${BASE}/admin/employees`, { waitUntil: 'networkidle' })
  await page.waitForTimeout(800)
  await page.locator('tr', { hasText: 'Budi Santoso' }).locator('button:has-text("Ubah")').click()
  await page.waitForSelector('.ant-modal:has-text("Ubah Karyawan")', { state: 'visible' })
  const quotaValue = await page.locator('.ant-modal .ant-form-item:has-text("Kuota cuti") input').inputValue()
  check('kuota cuti tersimpan dan tampil kembali', quotaValue === '18', `got ${quotaValue}`)
  await page.locator('.ant-modal-footer button:has-text("Batal")').click()
  await page.waitForTimeout(400)

  // ---------- Leave request editing ----------
  console.log('\n=== HR UBAH PENGAJUAN CUTI ===')
  await page.goto(`${BASE}/admin/leave-requests`, { waitUntil: 'networkidle' })
  await page.waitForTimeout(900)
  const leaveRow = page.locator('.ant-table-row').first()
  check('ada pengajuan cuti untuk diuji', (await leaveRow.count()) > 0)
  if (await leaveRow.count()) {
    await leaveRow.locator('button:has-text("Ubah")').click()
    await page.waitForSelector('.ant-modal:has-text("Ubah Pengajuan Cuti")', { state: 'visible' })
    const lm = page.locator('.ant-modal:has-text("Ubah Pengajuan Cuti")')
    check('modal ubah pengajuan cuti terbuka', await lm.isVisible())
    const reason = lm.locator('textarea')
    await reason.fill('diralat oleh HR lewat UI')
    await page.locator('.ant-modal-footer button.ant-btn-primary').click()
    await page.waitForTimeout(1500)
    const listText = await page.locator('body').innerText()
    check('alasan hasil edit tampil di tabel', listText.includes('diralat oleh HR lewat UI'), listText.slice(0, 200))
  }

  // ---------- Own profile from the header ----------
  console.log('\n=== EDIT PROFIL SENDIRI (HR) ===')
  await page.goto(`${BASE}/admin/monitoring`, { waitUntil: 'networkidle' })
  await page.waitForTimeout(700)
  await page.locator('.user-menu').click()
  await page.waitForTimeout(500)
  await page.locator('.ant-dropdown-menu-item:has-text("Profil Saya")').click()
  await page.waitForSelector('.ant-modal:has-text("Profil Saya")', { state: 'visible' })
  const pm = page.locator('.ant-modal:has-text("Profil Saya")')
  check('modal profil terbuka dari header', await pm.isVisible())
  const nameInput = pm.locator('input').first()
  await nameInput.fill(`Ratna HR ${stamp}`)
  await page.locator('.ant-modal-footer button.ant-btn-primary').click()
  await page.waitForTimeout(1500)
  const header = await page.locator('.user-menu').innerText()
  check('nama di header ikut berubah setelah simpan profil',
    header.includes(`Ratna HR ${stamp}`), header)

  await page.reload({ waitUntil: 'networkidle' })
  await page.waitForTimeout(1200)
  check('nama profil bertahan setelah reload (bukan lagi "Admin")',
    (await page.locator('.user-menu').innerText()).includes('Ratna HR'),
    await page.locator('.user-menu').innerText())

  console.log('\n=== EDIT PROFIL SENDIRI (KARYAWAN) ===')
  const ctx3 = await browser.newContext()
  const page3 = await ctx3.newPage()
  const errors3 = []
  page3.on('pageerror', (e) => errors3.push(e.message))
  await login(page3, EMPLOYEE_EMAIL, 'PasswordUjiUI123')
  await page3.goto(`${BASE}/riwayat`, { waitUntil: 'networkidle' })
  await page3.waitForTimeout(700)
  await page3.locator('.user-menu').click()
  await page3.waitForTimeout(500)
  await page3.locator('.ant-dropdown-menu-item:has-text("Profil Saya")').click()
  await page3.waitForSelector('.ant-modal:has-text("Profil Saya")', { state: 'visible' })
  check('karyawan bisa membuka profilnya sendiri',
    await page3.locator('.ant-modal:has-text("Profil Saya")').isVisible())
  check('profil karyawan menampilkan NIK',
    (await page3.locator('.ant-modal:has-text("Profil Saya")').innerText()).includes('NIK'))
  check('tidak ada error JS di alur karyawan', errors3.length === 0, errors3.join(' | '))
  await ctx3.close()

  check('tidak ada error JS di seluruh alur HR', errors.length === 0, errors.slice(0, 3).join(' | '))

  await page.screenshot({ path: '/tmp/shots/hr-editing.png', fullPage: true })
  await browser.close()
  console.log(`\n${'='.repeat(52)}\nPASS: ${pass}   FAIL: ${fail}\n${'='.repeat(52)}`)
  process.exit(fail ? 1 : 0)
})()
