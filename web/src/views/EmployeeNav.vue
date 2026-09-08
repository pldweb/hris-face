<script setup lang="ts">
// Shared top bar for every employee-facing page (Absen, Riwayat, Tim). Before
// this, each page rolled its own ad-hoc header with different buttons in a
// different order and no visual link to the admin shell -- confusing to
// navigate and visibly a different product. This mirrors AdminLayout's sider:
// same brand mark, same solid-blue active pill, same box-shadow-not-dropdown
// language, so the employee side reads as the same app, not a bolted-on page.
import { CalendarOutlined, DashboardOutlined, DownOutlined, FileTextOutlined, HistoryOutlined, LogoutOutlined, TeamOutlined, UserOutlined } from '@ant-design/icons-vue'
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { fetchProfile } from '../api/employee'
import { useAuthStore } from '../stores/auth'
import ProfileModal from './ProfileModal.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const isAdmin = computed(() => ['hr', 'superadmin'].includes(auth.role ?? ''))
// The team screen is served by /team/*, which the API restricts to managers
// and above. Showing "Tim" to everyone meant an ordinary employee clicked a
// menu item that could only ever answer 403.
const canSeeTeam = computed(() => ['manager', 'hr', 'superadmin'].includes(auth.role ?? ''))
const initials = computed(() =>
  (auth.fullName ?? 'Karyawan').split(' ').map((part) => part[0]).join('').slice(0, 2).toUpperCase(),
)

const links = computed(() => [
  { to: '/checkin', label: 'Absen', icon: CalendarOutlined },
  { to: '/riwayat', label: 'Riwayat Saya', icon: HistoryOutlined },
  { to: '/cuti', label: 'Cuti', icon: FileTextOutlined },
  ...(canSeeTeam.value ? [{ to: '/tim', label: 'Tim', icon: TeamOutlined }] : []),
])

const profileOpen = ref(false)

onMounted(async () => {
  if (auth.fullName) return
  try {
    auth.fullName = (await fetchProfile()).full_name
  } catch {
    // Non-fatal: the nav falls back to a generic label.
  }
})

async function onLogout() {
  await auth.logoutFully()
  router.push('/login')
}
</script>

<template>
  <header class="employee-nav">
    <router-link to="/checkin" class="brand"><span class="brand-mark">H</span> HRIS</router-link>

    <nav class="nav-links">
      <router-link
        v-for="link in links"
        :key="link.to"
        :to="link.to"
        class="nav-link"
        :class="{ active: route.path === link.to }"
      >
        <component :is="link.icon" />
        {{ link.label }}
      </router-link>
    </nav>

    <div class="nav-right">
      <router-link v-if="isAdmin" to="/admin/monitoring">
        <a-button class="admin-shortcut" type="primary" ghost>
          <template #icon><DashboardOutlined /></template>
          Dashboard admin
        </a-button>
      </router-link>
      <a-dropdown>
        <a-button type="text" class="user-menu">
          <span class="user-avatar">{{ initials }}</span>
          <span class="user-name">{{ auth.fullName ?? 'Karyawan' }}</span>
          <DownOutlined class="user-chevron" />
        </a-button>
        <template #overlay>
          <a-menu>
            <a-menu-item @click="profileOpen = true"><UserOutlined /> Profil Saya</a-menu-item>
            <a-menu-item @click="onLogout"><LogoutOutlined /> Keluar</a-menu-item>
          </a-menu>
        </template>
      </a-dropdown>
    </div>

    <ProfileModal v-model:open="profileOpen" />
  </header>
</template>

<style scoped>
.employee-nav {
  display: flex;
  align-items: center;
  gap: 28px;
  padding: 12px 32px;
  background: #fff;
  box-shadow: 0 6px 24px -12px rgba(23, 37, 84, 0.18);
  position: relative;
  z-index: 1;
}

.brand {
  display: flex;
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

.nav-links {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-right: auto;
}
.nav-link {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  border-radius: 8px;
  color: #475569;
  font-size: 14px;
  font-weight: 500;
}
.nav-link:hover { background: #f1f5fb; color: #172554; }
.nav-link.active {
  background: #21409a;
  color: #fff;
  box-shadow: 0 6px 16px -4px rgba(33, 64, 154, 0.35);
}

.nav-right { display: flex; align-items: center; gap: 12px; }
.admin-shortcut { border-radius: 8px; }

.user-menu { height: 44px; display: inline-flex; align-items: center; gap: 10px; padding: 4px 8px 4px 4px; border: 1px solid #e2e8f0; border-radius: 10px; background: #fff; }
.user-menu:hover { border-color: #b9c8ee; background: #f8faff; }
.user-avatar { width: 32px; height: 32px; display: grid; place-items: center; border-radius: 8px; color: #21409a; background: #e9efff; font-size: 11px; font-weight: 800; }
.user-name { color: #172554; font-size: 13px; font-weight: 600; max-width: 140px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.user-chevron { color: #94a3b8; font-size: 11px; }

@media (max-width: 720px) {
  .employee-nav { padding: 10px 16px; gap: 12px; }
  .brand span:not(.brand-mark) { display: none; }
  .nav-link span { display: none; }
  .user-name { display: none; }
}
</style>
