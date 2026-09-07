<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { apiClient } from '../api/client'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const submitting = ref(false)
const errorMessage = ref('')

const form = reactive({ email: '', password: '' })

async function onSubmit() {
  submitting.value = true
  errorMessage.value = ''
  try {
    const { data } = await apiClient.post('/auth/login', form)
    auth.setToken(data.access_token)
    const defaultRoute = ['hr', 'superadmin'].includes(auth.role ?? '') ? '/admin/monitoring' : '/checkin'
    const redirect = (router.currentRoute.value.query.redirect as string) || defaultRoute
    router.push(redirect)
  } catch (err: any) {
    errorMessage.value = err?.response?.data?.error ?? 'Terjadi kesalahan'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-shell">
      <section class="login-intro">
        <div class="brand-mark">H</div>
        <p class="eyebrow">HRIS · FACE ATTENDANCE</p>
        <h1>Absensi yang jelas, tercatat, dan siap dipertanggungjawabkan.</h1>
        <p class="intro-copy">Masuk untuk mengelola kehadiran atau melakukan absensi melalui verifikasi wajah.</p>
        <div class="intro-rule" />
        <span class="intro-note">Sistem internal perusahaan</span>
      </section>
      <a-card class="login-card">
        <div class="login-heading">
          <p class="eyebrow">PORTAL KARYAWAN</p>
          <h2>Selamat datang</h2>
          <p>Gunakan akun perusahaan Anda untuk melanjutkan.</p>
        </div>
        <!-- :model and per-item name are required: without them a-form never
             emits `finish` and the submit button does nothing at all. -->
        <a-form layout="vertical" :model="form" @finish="onSubmit">
          <a-form-item label="Email" name="email" :rules="[{ required: true, message: 'Email wajib diisi' }]">
            <a-input v-model:value="form.email" size="large" type="email" autocomplete="username" />
          </a-form-item>
          <a-form-item label="Password" name="password" :rules="[{ required: true, message: 'Password wajib diisi' }]">
            <a-input-password v-model:value="form.password" size="large" autocomplete="current-password" />
          </a-form-item>
          <a-alert v-if="errorMessage" type="error" :message="errorMessage" show-icon class="login-error" />
          <a-button type="primary" html-type="submit" size="large" :loading="submitting" block>Masuk ke sistem</a-button>
        </a-form>
      </a-card>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px;
  background: #eef2f8;
}

.login-shell {
  width: min(940px, 100%);
  min-height: 560px;
  display: grid;
  grid-template-columns: 1.05fr .95fr;
  overflow: hidden;
  border: 1px solid #dbe3f0;
  border-radius: 18px;
  background: #fff;
  box-shadow: 0 24px 70px rgba(23, 37, 84, .14);
}

.login-intro {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 64px;
  color: #fff;
  background: #172f78;
}

.brand-mark {
  width: 42px;
  height: 42px;
  display: grid;
  place-items: center;
  margin-bottom: 64px;
  border: 1px solid rgba(255, 255, 255, .32);
  border-radius: 10px;
  font-size: 21px;
  font-weight: 800;
}

.eyebrow {
  margin: 0 0 12px;
  color: #8ea7e9;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: .12em;
}

.login-intro h1 {
  max-width: 430px;
  margin: 0;
  font-size: clamp(30px, 4vw, 44px);
  line-height: 1.08;
  letter-spacing: -.04em;
}

.intro-copy {
  max-width: 380px;
  margin: 24px 0 0;
  color: #cbd8fa;
  font-size: 16px;
  line-height: 1.6;
}

.intro-rule { width: 56px; height: 2px; margin: 48px 0 14px; background: #8ea7e9; }
.intro-note { color: #aebfe8; font-size: 13px; }

.login-card {
  align-self: center;
  width: 100%;
  max-width: 390px;
  margin: 0 auto;
  padding: 18px;
  border: 0;
  box-shadow: none;
}

.login-heading { margin-bottom: 28px; }
.login-heading .eyebrow { color: #5470bc; }
.login-heading h2 { margin: 0; font-size: 30px; letter-spacing: -.03em; }
.login-heading p:last-child { margin: 8px 0 0; color: #64748b; }

.login-error {
  margin-bottom: 16px;
}

@media (max-width: 720px) {
  .login-page { padding: 16px; }
  .login-shell { display: block; min-height: 0; }
  .login-intro { padding: 32px; }
  .brand-mark { margin-bottom: 42px; }
  .login-card { padding: 32px; }
}
</style>
