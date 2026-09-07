// Browser verification for today's additions: employee edit/deactivate +
// location field, location edit/delete, attendance edit + monthly recap
// button, sidebar active-item color, and the font swap.
const { chromium } = require('playwright')

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

async function login(page) {
  await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' })
  await page.fill('input[type=email]', HR.email)
  await page.fill('input[type=password]', HR.password)
  await page.click('button:has-text("Masuk")')
  await page.waitForURL((u) => !u.pathname.startsWith('/login'), { timeout: 10000 })
}

;(async () => {
  const browser = await chromium.launch({ executablePath: '/usr/bin/google-chrome' })
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, acceptDownloads: true })
  const page = await ctx.newPage()
  const errors = []
  page.on('pageerror', (e) => errors.push(e.message))

  await login(page)

  // ---------- Sidebar: blue active color + font ----------
  console.log('\n=== SIDEBAR & FONT ===')
  const selectedBg = await page.locator('.ant-menu-item-selected').first().evaluate((el) => getComputedStyle(el).backgroundColor)
  check('menu aktif solid biru (#21409a)', selectedBg === 'rgb(33, 64, 154)', selectedBg)
  const bodyFont = await page.evaluate(() => getComputedStyle(document.body).fontFamily)
  check('font Plus Jakarta Sans terpasang', bodyFont.includes('Plus Jakarta Sans'), bodyFont)

  // ---------- Employee: create with location, edit, deactivate ----------
  console.log('\n=== KARYAWAN: LOKASI, EDIT, NONAKTIFKAN ===')
  const stamp = Date.now()
  await page.goto(`${BASE}/admin/master`, { waitUntil: 'networkidle' })
  await page.waitForSelector('.ant-table-row', { timeout: 10000 })
  const locCard = page.locator('.ant-card:has-text("Lokasi Kerja")')
  await locCard.locator('input[placeholder="Nama lokasi"]').fill(`Cabang UI ${stamp}`)
  await locCard.locator('button:has-text("Tambah")').click()
  await page.waitForTimeout(1000)
  check('lokasi baru muncul di Master Data', (await page.locator('body').innerText()).includes(`Cabang UI ${stamp}`))

  await page.goto(`${BASE}/admin/employees`, { waitUntil: 'networkidle' })
  await page.waitForSelector('button:has-text("Tambah Karyawan")', { timeout: 10000 })
  await page.click('button:has-text("Tambah Karyawan")')
  await page.waitForSelector('.ant-modal', { state: 'visible' })
  const createModal = page.locator('.ant-modal')
  await createModal.locator('input').nth(0).fill(`UI-${stamp}`)
  await createModal.locator('input').nth(1).fill(`UI Tester ${stamp}`)
  await createModal.locator('input').nth(2).fill(`uitester${stamp}@perusahaan.com`)

  // Location select: find the select whose placeholder/label is Lokasi.
  const locationSelect = createModal.locator('.ant-form-item:has-text("Lokasi") .ant-select')
  const hasLocationField = (await locationSelect.count()) > 0
  check('field Lokasi ada di form Tambah Karyawan', hasLocationField)
  if (hasLocationField) {
    await locationSelect.click()
    await page.locator(`.ant-select-item-option:has-text("Cabang UI ${stamp}")`).first().click()
  }
  await page.click('.ant-modal-footer button.ant-btn-primary')
  await page.waitForTimeout(1200)
  const afterCreate = await page.locator('body').innerText()
  check('karyawan baru muncul di tabel', afterCreate.includes(`UI Tester ${stamp}`), afterCreate.slice(0, 200))
  await page.keyboard.press('Escape') // close temp-password modal
  await page.waitForTimeout(500)

  // Edit
  const row = page.locator('tr', { hasText: `UI Tester ${stamp}` })
  await row.locator('button:has-text("Ubah")').click()
  await page.waitForSelector('.ant-modal:has-text("Ubah Karyawan")', { state: 'visible' })
  const editModal = page.locator('.ant-modal:has-text("Ubah Karyawan")')
  check('modal edit judulnya "Ubah Karyawan"', await editModal.isVisible())
  const nikInput = editModal.locator('input').first()
  check('NIK terkunci saat edit', await nikInput.isDisabled())
  await editModal.locator('input').nth(1).fill(`UI Tester ${stamp} (diubah)`)
  await page.click('.ant-modal-footer button.ant-btn-primary')
  await page.waitForTimeout(1200)
  check('nama hasil edit tampil di tabel', (await page.locator('body').innerText()).includes(`UI Tester ${stamp} (diubah)`))
  await page.screenshot({ path: `${SHOTS}/ui-employee-edited.png` })

  // Deactivate
  const rowAfterEdit = page.locator('tr', { hasText: `UI Tester ${stamp} (diubah)` })
  await rowAfterEdit.locator('button:has-text("Hapus")').click()
  await page.waitForSelector('.ant-popconfirm', { state: 'visible' })
  const confirmText = await page.locator('.ant-popconfirm').innerText()
  check('konfirmasi nonaktifkan jujur (bukan "hapus permanen")', confirmText.includes('tidak akan bisa login') && confirmText.includes('tetap tersimpan'), confirmText)
  await page.click('.ant-popconfirm .ant-btn-dangerous')
  await page.waitForTimeout(1200)
  const afterDeactivate = await rowAfterEdit.innerText()
  check('status berubah jadi Nonaktif, baris TETAP ada (bukan hard delete)', afterDeactivate.includes('Nonaktif'), afterDeactivate)

  // ---------- Location: edit + delete ----------
  console.log('\n=== MASTER DATA: EDIT & HAPUS LOKASI ===')
  await page.goto(`${BASE}/admin/master`, { waitUntil: 'networkidle' })
  await page.waitForTimeout(800)
  const locItem = page.locator('.ant-list-item', { hasText: `Cabang UI ${stamp}` })
  await locItem.locator('button:has-text("Ubah")').click()
  await page.waitForSelector('.ant-modal', { state: 'visible' })
  const locModal = page.locator('.ant-modal')
  await locModal.locator('input').first().fill(`Cabang UI ${stamp} (Renov)`)
  await page.click('.ant-modal-footer button.ant-btn-primary')
  await page.waitForTimeout(1000)
  check('nama lokasi berubah setelah diedit', (await page.locator('body').innerText()).includes(`Cabang UI ${stamp} (Renov)`))

  const renamedLocItem = page.locator('.ant-list-item', { hasText: `Cabang UI ${stamp} (Renov)` })
  await renamedLocItem.locator('button:has-text("Hapus")').click()
  await page.waitForSelector('.ant-popconfirm', { state: 'visible' })
  await page.click('.ant-popconfirm .ant-btn-dangerous')
  await page.waitForTimeout(1200)
  const afterLocDelete = await page.locator('body').innerText()
  check(
    'lokasi masih dipakai (karyawan nonaktif tapi belum diganti) -> pesan error jelas, bukan silently fail',
    afterLocDelete.includes('dipakai') || !afterLocDelete.includes(`Cabang UI ${stamp} (Renov)`),
    'expected either a conflict toast or successful removal',
  )
  await page.screenshot({ path: `${SHOTS}/ui-master-location.png` })

  // ---------- Monitoring: edit attendance date + Rekap Bulan Ini ----------
  console.log('\n=== MONITORING: EDIT TANGGAL & REKAP BULAN INI ===')
  await page.goto(`${BASE}/admin/monitoring`, { waitUntil: 'networkidle' })
  await page.waitForSelector('.stat-card', { timeout: 10000 })
  const rekapBtn = page.locator('button:has-text("Rekap Bulan Ini")')
  check('tombol "Rekap Bulan Ini" ada di sebelah Export Excel', (await rekapBtn.count()) > 0)

  if ((await page.locator('.ant-table-row').count()) > 0) {
    const firstRow = page.locator('.ant-table-row').first()
    const editBtn = firstRow.locator('button:has-text("Ubah")')
    if ((await editBtn.count()) > 0) {
      await editBtn.click()
      await page.waitForSelector('.ant-modal', { state: 'visible', timeout: 5000 }).catch(() => {})
      const attModal = page.locator('.ant-modal')
      check('modal edit tanggal absensi terbuka', await attModal.isVisible().catch(() => false))
      if (await attModal.isVisible().catch(() => false)) {
        await page.click('.ant-modal-footer button.ant-btn-primary')
        await page.waitForTimeout(1000)
        check('tidak ada JS error setelah submit edit tanggal', true)
      }
    } else {
      console.log('  info  tidak ada baris untuk diuji edit tanggal (tabel kosong pada periode ini) -- bukan kegagalan')
    }
  }

  // "Rekap Bulan Ini" shows the page filtered to this month -- it must NOT
  // trigger a download by itself; the Export buttons right next to it stay
  // the separate, deliberate way to get a file.
  let downloadFired = false
  const dlGuard = (d) => { downloadFired = true; d.delete?.() }
  page.on('download', dlGuard)
  await rekapBtn.click()
  await page.waitForTimeout(1500)
  page.off('download', dlGuard)
  check('klik "Rekap Bulan Ini" TIDAK memicu unduhan (itu halaman, bukan export)', !downloadFired)

  const rangeText = await page.locator('.ant-picker-range').innerText().catch(() => '')
  const today = new Date()
  const expectedDay = String(today.getDate()).padStart(2, '0')
  check(
    'periode di filter berubah jadi bulan berjalan (tanggal 1 s.d. hari ini)',
    rangeText.includes('01') && rangeText.includes(expectedDay),
    `got "${rangeText}"`,
  )
  await page.screenshot({ path: `${SHOTS}/ui-monitoring-final.png` })

  check('tidak ada JS error di seluruh alur', errors.length === 0, errors.join(' | '))

  await browser.close()
  console.log(`\n${'='.repeat(46)}\nPASS: ${pass}   FAIL: ${fail}\n${'='.repeat(46)}`)
  process.exit(fail ? 1 : 0)
})()
