import Antd from 'ant-design-vue'
import 'ant-design-vue/dist/reset.css'
import { createPinia } from 'pinia'
import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { i18n } from './i18n'
import './theme/global.css'

createApp(App).use(createPinia()).use(router).use(i18n).use(Antd).mount('#app')
