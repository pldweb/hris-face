import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/checkin' },
    {
      path: '/checkin',
      name: 'checkin',
      component: () => import('../views/CheckIn.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/riwayat',
      name: 'my-history',
      component: () => import('../views/MyHistory.vue'),
      meta: { requiresAuth: true },
    },
    { path: '/login', name: 'login', component: () => import('../views/Login.vue') },
    {
      path: '/enroll',
      name: 'enroll',
      component: () => import('../views/Enrollment.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/tim',
      name: 'team',
      component: () => import('../views/TeamDashboard.vue'),
      // /team/* is manager-and-above on the API; without this a plain employee
      // could deep-link here and get a screen that only renders 403 toasts.
      meta: { requiresAuth: true, roles: ['manager', 'hr', 'superadmin'] },
    },
    {
      path: '/cuti',
      name: 'leave',
      component: () => import('../views/LeaveRequest.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/admin',
      component: () => import('../views/admin/AdminLayout.vue'),
      meta: { requiresAuth: true, roles: ['hr', 'superadmin'] },
      children: [
        { path: '', redirect: '/admin/monitoring' },
        { path: 'monitoring', name: 'admin-monitoring', component: () => import('../views/admin/Monitoring.vue') },
        { path: 'face-scan', name: 'admin-face-scan', component: () => import('../views/admin/FaceScan.vue') },
        { path: 'employees', name: 'admin-employees', component: () => import('../views/admin/EmployeeList.vue') },
        { path: 'corrections', name: 'admin-corrections', component: () => import('../views/admin/Corrections.vue') },
        { path: 'leave-requests', name: 'admin-leave-requests', component: () => import('../views/admin/LeaveRequests.vue') },
        { path: 'master', name: 'admin-master', component: () => import('../views/admin/MasterData.vue') },
      ],
    },
    // Catch-all, must stay last. Without it Vue Router renders an empty
    // <router-view> for any unknown path -- a blank white page that looks
    // like a crash instead of a wrong address.
    { path: '/:pathMatch(.*)*', name: 'not-found', component: () => import('../views/NotFound.vue') },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.accessToken) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  // Honour whatever roles a route declares, so a new restricted route only has
  // to set meta.roles instead of growing another special case here.
  const allowed = to.meta.roles as string[] | undefined
  if (allowed && !allowed.includes(auth.role ?? '')) {
    return { path: '/checkin' }
  }
  if (to.path === '/login' && auth.accessToken) {
    return { path: ['hr', 'superadmin'].includes(auth.role ?? '') ? '/admin/monitoring' : '/checkin' }
  }
  return true
})

export default router
