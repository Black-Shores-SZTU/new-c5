import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'

import App from './App.vue'
import router from './router'

// 全局令牌需在 Element Plus 样式之后加载，保证主题变量覆盖生效
import './styles/tokens.less'
import './styles/global.less'

const app = createApp(App)

app.use(router)
app.use(ElementPlus)
app.mount('#app')
