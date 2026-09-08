<script setup lang="ts">
// Without a catch-all route Vue Router renders nothing at all for an unknown
// path -- a literally blank white page, which reads as "the app is broken"
// rather than "that link is wrong". That happens for a typo'd URL, a stale
// bookmark, and (the way this was found) a tab still running an older bundle
// that predates a newly added route.
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const auth = useAuthStore()

const homePath = computed(() =>
  ['hr', 'superadmin'].includes(auth.role ?? '') ? '/admin/monitoring' : '/checkin',
)
</script>

<template>
  <div class="notfound-page">
    <div class="notfound-card">
      <div class="brand"><span class="brand-mark">H</span> HRIS</div>
      <p class="code">404</p>
      <h1>Halaman tidak ditemukan</h1>
      <p class="lead">
        Alamat <code>{{ route.fullPath }}</code> tidak ada di aplikasi ini.
        Kemungkinan salah ketik, tautan lama, atau halaman sudah dipindahkan.
      </p>
      <p class="hint">
        Kalau halaman ini muncul setelah aplikasi baru diperbarui, muat ulang
        halaman (Ctrl+Shift+R) supaya browser mengambil versi terbaru.
      </p>
      <router-link :to="homePath">
        <a-button type="primary" size="large">Kembali ke halaman utama</a-button>
      </router-link>
    </div>
  </div>
</template>

<style scoped>
.notfound-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--ant-color-bg-layout, #f5f6f8);
  padding: 24px;
}

.notfound-card {
  width: 100%;
  max-width: 520px;
  padding: 40px;
  text-align: center;
  background: #fff;
  border-radius: 14px;
  box-shadow: 0 18px 48px -24px rgba(23, 37, 84, 0.28);
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font-weight: 700;
  font-size: 17px;
  letter-spacing: -0.03em;
  color: #172554;
}
.brand-mark {
  width: 28px;
  height: 28px;
  display: grid;
  place-items: center;
  border-radius: 8px;
  color: #fff;
  background: #21409a;
  font-size: 14px;
}

.code {
  margin: 22px 0 0;
  font-size: 56px;
  font-weight: 700;
  line-height: 1;
  letter-spacing: -0.04em;
  color: #21409a;
}

h1 {
  margin: 8px 0 12px;
  font-size: 21px;
  color: #172554;
}

.lead {
  margin: 0 0 10px;
  color: #64748b;
  line-height: 1.6;
}
.lead code {
  padding: 2px 6px;
  border-radius: 6px;
  background: #f1f5fb;
  color: #21409a;
  word-break: break-all;
}

.hint {
  margin: 0 0 24px;
  color: #94a3b8;
  font-size: 13px;
  line-height: 1.6;
}
</style>
