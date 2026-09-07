<script setup lang="ts">
import idID from 'ant-design-vue/es/locale/id_ID'
import { themeTokens } from './theme/tokens'
</script>

<template>
  <a-config-provider :theme="themeTokens" :locale="idID">
    <router-view />
  </a-config-provider>
</template>
