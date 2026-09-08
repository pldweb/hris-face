<script setup lang="ts">
// Editing your own account, reachable from the header on both the admin and
// the employee side. Password change is deliberately opt-in: the fields stay
// collapsed until asked for, so the common "fix my name" edit never makes
// someone think about their password.
import { message } from 'ant-design-vue'
import { ref, watch } from 'vue'
import { fetchProfile, updateProfile, type Profile } from '../api/employee'
import { useAuthStore } from '../stores/auth'

const open = defineModel<boolean>('open', { required: true })
const emit = defineEmits<{ saved: [name: string] }>()

const auth = useAuthStore()
const loading = ref(false)
const saving = ref(false)
const profile = ref<Profile | null>(null)
const form = ref({ full_name: '', email: '' })
const changingPassword = ref(false)
const passwords = ref({ current_password: '', new_password: '', confirm: '' })

async function load() {
  loading.value = true
  changingPassword.value = false
  passwords.value = { current_password: '', new_password: '', confirm: '' }
  try {
    const p = await fetchProfile()
    profile.value = p
    form.value = { full_name: p.full_name, email: p.email }
  } catch {
    message.error('Gagal memuat profil')
  } finally {
    loading.value = false
  }
}

watch(open, (isOpen) => {
  if (isOpen) load()
})

const ROLE_LABELS: Record<string, string> = {
  superadmin: 'Super Admin',
  hr: 'HR',
  manager: 'Manager',
  employee: 'Karyawan',
}

async function onSubmit() {
  if (!form.value.full_name.trim()) {
    message.error('Nama wajib diisi')
    return
  }
  if (!form.value.email.trim()) {
    message.error('Email wajib diisi')
    return
  }
  if (changingPassword.value) {
    if (!passwords.value.current_password) {
      message.error('Password saat ini wajib diisi untuk mengganti password')
      return
    }
    if (passwords.value.new_password.length < 8) {
      message.error('Password baru minimal 8 karakter')
      return
    }
    if (passwords.value.new_password !== passwords.value.confirm) {
      message.error('Konfirmasi password tidak cocok')
      return
    }
  }

  saving.value = true
  try {
    await updateProfile({
      full_name: form.value.full_name.trim(),
      email: form.value.email.trim(),
      ...(changingPassword.value
        ? {
            current_password: passwords.value.current_password,
            new_password: passwords.value.new_password,
          }
        : {}),
    })
    auth.fullName = form.value.full_name.trim()
    emit('saved', form.value.full_name.trim())
    open.value = false
    message.success(
      changingPassword.value
        ? 'Profil dan password diperbarui. Perangkat lain perlu login ulang.'
        : 'Profil diperbarui',
    )
  } catch (err: any) {
    message.error(err?.response?.data?.error ?? 'Gagal menyimpan profil')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <a-modal v-model:open="open" title="Profil Saya" :confirm-loading="saving" ok-text="Simpan" cancel-text="Batal" @ok="onSubmit">
    <a-spin :spinning="loading">
      <a-form layout="vertical" class="profile-form">
        <a-form-item v-if="profile" label="Peran">
          <a-tag color="blue">{{ ROLE_LABELS[profile.role] ?? profile.role }}</a-tag>
          <span v-if="profile.nik" class="nik">NIK {{ profile.nik }}</span>
        </a-form-item>
        <a-form-item label="Nama lengkap">
          <a-input v-model:value="form.full_name" placeholder="Nama lengkap" />
        </a-form-item>
        <a-form-item label="Email">
          <a-input v-model:value="form.email" type="email" placeholder="nama@perusahaan.com" />
        </a-form-item>

        <a-checkbox v-model:checked="changingPassword">Ganti password</a-checkbox>

        <template v-if="changingPassword">
          <a-form-item label="Password saat ini" class="pw-first">
            <a-input-password v-model:value="passwords.current_password" placeholder="Password saat ini" />
          </a-form-item>
          <a-form-item label="Password baru">
            <a-input-password v-model:value="passwords.new_password" placeholder="Minimal 8 karakter" />
          </a-form-item>
          <a-form-item label="Ulangi password baru">
            <a-input-password v-model:value="passwords.confirm" placeholder="Ulangi password baru" />
          </a-form-item>
          <p class="pw-hint">Setelah password diganti, sesi di perangkat lain akan diminta login ulang.</p>
        </template>
      </a-form>
    </a-spin>
  </a-modal>
</template>

<style scoped>
.profile-form {
  padding-top: 4px;
}
.nik {
  margin-left: 8px;
  color: #94a3b8;
  font-size: 12px;
}
.pw-first {
  margin-top: 12px;
}
.pw-hint {
  margin: -8px 0 0;
  color: #94a3b8;
  font-size: 12px;
}
</style>
