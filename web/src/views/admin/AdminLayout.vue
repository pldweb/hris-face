<script setup lang="ts">
// docs/PRD.md 10.4 -- standard AntD Sider + Header. No invented navigation;
// consistency beats surprise on an Operate surface.
import {
  ArrowRightOutlined,
  AuditOutlined,
  CalendarOutlined,
  CustomerServiceOutlined,
  DashboardOutlined,
  DownOutlined,
  LogoutOutlined,
  ScanOutlined,
  SettingOutlined,
  TeamOutlined,
} from '@ant-design/icons-vue'
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const selectedKeys = computed(() => [route.path])
const pageTitle = computed(() => route.path.includes('face-scan') ? 'Scan Wajah' : route.path.includes('employees') ? 'Karyawan' : route.path.includes('corrections') ? 'Koreksi Absen' : route.path.includes('master') ? 'Master Data' : 'Monitoring')
const initials = computed(() => (auth.fullName ?? 'Admin').split(' ').map((part) => part[0]).join('').slice(0, 2).toUpperCase())

async function onLogout() {
  await auth.logoutFully()
  router.push('/login')
}
</script>

<template>
  <a-layout class="admin-shell" style="min-height: 100vh">
    <a-layout-sider class="admin-sider" theme="light" width="244" breakpoint="lg" collapsed-width="0">
      <div class="brand"><span class="brand-mark">H</span><span>HRIS</span></div>
      <div class="brand-subtitle">Face Attendance</div>
      <a-menu mode="inline" :selected-keys="selectedKeys">
        <a-menu-item key="/admin/monitoring">
          <DashboardOutlined /><router-link to="/admin/monitoring">Monitoring</router-link>
        </a-menu-item>
        <a-menu-item key="/admin/face-scan">
          <ScanOutlined /><router-link to="/admin/face-scan">Scan Wajah</router-link>
        </a-menu-item>
        <a-menu-item key="/admin/employees">
          <TeamOutlined /><router-link to="/admin/employees">Karyawan</router-link>
        </a-menu-item>
        <a-menu-item key="/admin/corrections">
          <AuditOutlined /><router-link to="/admin/corrections">Koreksi Absen</router-link>
        </a-menu-item>
        <a-menu-item key="/admin/master">
          <SettingOutlined /><router-link to="/admin/master">Master Data</router-link>
        </a-menu-item>
      </a-menu>

      <div class="help-widget">
        <CustomerServiceOutlined class="help-icon" />
        <strong>Butuh bantuan?</strong>
        <p>Hubungi tim IT jika mengalami kendala.</p>
        <a href="mailto:it@perusahaan.com" class="help-link">
          Hubungi Sekarang <ArrowRightOutlined />
        </a>
      </div>
    </a-layout-sider>
    <a-layout>
      <a-layout-header class="admin-header">
        <div class="header-title">
          <strong style="line-height: normal;">{{ pageTitle }}</strong>
          <span class="header-subtitle" style="line-height: normal;">Panel administrasi HR</span>
        </div>
        <a-space size="middle">
          <router-link to="/checkin">
            <a-button class="attendance-shortcut" type="primary" ghost>
              <template #icon><CalendarOutlined /></template>
              Halaman absen
            </a-button>
          </router-link>
          <a-dropdown>
            <a-button type="text" class="user-menu">
              <span class="user-avatar">{{ initials }}</span>
              <span class="user-copy"><strong>{{ auth.fullName ?? 'Admin' }}</strong><small>Administrator</small></span>
              <DownOutlined class="user-chevron" />
            </a-button>
            <template #overlay>
              <a-menu>
                <a-menu-item @click="onLogout"><LogoutOutlined /> Keluar</a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </a-space>
      </a-layout-header>
      <a-layout-content class="admin-content">
        <router-view />
      </a-layout-content>
      <footer class="admin-footer">© {{ new Date().getFullYear() }} HRIS Face Attendance. All rights reserved.</footer>
    </a-layout>
  </a-layout>
</template>

<style scoped>
.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 700;
  font-size: 19px;
  letter-spacing: -.03em;
  padding: 22px 24px 4px;
}
.brand-mark { width: 30px; height: 30px; display: grid; place-items: center; border-radius: 8px; color: #fff; background: #21409a; font-size: 15px; }
.brand-subtitle { padding: 0 24px 26px; color: #94a3b8; font-size: 12px; }
.admin-sider {
  display: flex;
  flex-direction: column;
  /* A soft, wide elevation reads as a lifted panel; a tight dark shadow reads as
     a floating dropdown menu, which is what "jangan dropdown" rules out. */
  box-shadow: 6px 0 32px -12px rgba(23, 37, 84, 0.14);
  position: relative;
  z-index: 1;
}
.admin-sider :deep(.ant-layout-sider-children) { display: flex; flex-direction: column; height: 100%; }
.admin-sider :deep(.ant-menu) { border-inline-end: 0; padding: 0 12px; }
.admin-sider :deep(.ant-menu-item) { margin: 4px 0; border-radius: 8px; }
/* Solid primary blue, matching the other buttons on the page (Terapkan Filter,
   Export, etc.) -- not AntD's default light-tint selected style. */
.admin-sider :deep(.ant-menu-item-selected) {
  background: #21409a !important;
  box-shadow: 0 6px 16px -4px rgba(33, 64, 154, 0.35);
}
.admin-sider :deep(.ant-menu-item-selected .ant-menu-title-content),
.admin-sider :deep(.ant-menu-item-selected .ant-menu-title-content a),
.admin-sider :deep(.ant-menu-item-selected .anticon) {
  color: #fff;
}
/* The icon + label live inside AntD's own .ant-menu-title-content wrapper, not
   as direct children of .ant-menu-item -- gap/flex has to target that wrapper
   or it silently does nothing (verified in devtools before this fix). */
.admin-sider :deep(.ant-menu-title-content) { display: flex; align-items: center; gap: 14px; }
.admin-sider :deep(.ant-menu-title-content .anticon) { margin: 0; font-size: 16px; }
.admin-sider :deep(.ant-menu-item a) { flex: 1; }

.help-widget {
  margin: auto 16px 20px;
  padding: 18px;
  border-radius: 12px;
  background: linear-gradient(155deg, #eef2ff, #e6ecff);
  border: 1px solid #dbe3fb;
}
.help-icon { color: #21409a; font-size: 18px; margin-bottom: 8px; }
.help-widget strong { display: block; color: #172554; font-size: 13px; margin-bottom: 4px; }
.help-widget p { margin: 0 0 10px; color: #64748b; font-size: 12px; line-height: 1.5; }
.help-link { display: inline-flex; align-items: center; gap: 6px; color: #21409a; font-size: 12px; font-weight: 600; }
.help-link:hover { text-decoration: underline; }

.admin-header {
  background: #fff;
  display: flex;
  justify-content: flex-end;
  align-items: center;
  min-height: 72px;
  padding: 0 32px;
  border-bottom: 1px solid #e2e8f0;
}
.header-title { display: flex; flex-direction: column; gap: 3px; margin-right: auto; }
.header-title span { color: #64748b; font-size: 13px; }
.header-title strong { color: #172554; font-size: 19px; letter-spacing: -.02em; }
.attendance-shortcut { border-radius: 8px; }

.user-menu { height: 48px; display: inline-flex; align-items: center; gap: 10px; padding: 4px 8px 4px 4px; border: 1px solid #e2e8f0; border-radius: 10px; background: #fff; }
.user-menu:hover { border-color: #b9c8ee; background: #f8faff; }
.user-avatar { width: 36px; height: 36px; display: grid; place-items: center; border-radius: 8px; color: #21409a; background: #e9efff; font-size: 12px; font-weight: 800; }
.user-copy { display: flex; flex-direction: column; align-items: flex-start; line-height: 1.2; }
.user-copy strong { color: #172554; font-size: 13px; }
.user-copy small { color: #64748b; font-size: 11px; }
.user-chevron { color: #94a3b8; font-size: 11px; }

.admin-content {
  padding: 32px;
  flex: 1;
}

.admin-footer {
  padding: 16px 32px 24px;
  text-align: center;
  color: #94a3b8;
  font-size: 12px;
}

@media (max-width: 720px) {
  .admin-header { padding: 0 16px; }
  .header-title { display: none; }
  .admin-content { padding: 20px 16px; }
}
</style>
