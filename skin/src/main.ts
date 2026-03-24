import { createApp } from 'vue'
import './style.css'
import App from './App.vue'

const app = createApp(App)
import router from './router'
app.use(router)
import pinia from './store'
app.use(pinia)


app.mount('#app')
