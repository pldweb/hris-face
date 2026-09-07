<script setup lang="ts">
// A manager sees exactly the monitoring screen, scoped to their direct reports
// by the server (employees.manager_id). Reusing the component means the two
// views cannot drift apart in what a status tag or a stat tile means.
import EmployeeNav from './EmployeeNav.vue'
import Monitoring from './admin/Monitoring.vue'
</script>

<template>
  <div class="team-shell">
    <EmployeeNav />
    <div class="team-page">
      <div class="team-column">
        <h2 class="page-title">Tim Saya</h2>
        <Monitoring team />
      </div>
    </div>
  </div>
</template>

<style scoped>
.team-shell {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.team-page {
  flex: 1;
  background: var(--ant-color-bg-layout, #f5f6f8);
  padding: 24px;
}

.team-column {
  max-width: 1100px;
  margin: 0 auto;
}

.page-title {
  margin: 0 0 16px;
}
</style>
