import './style.css';
import { mount } from 'svelte';
import App from './App.svelte';

const target = document.getElementById('app');

// Only mount on client-side (not during SSR)
if (typeof window !== 'undefined' && target) {
  mount(App, { target });
}
