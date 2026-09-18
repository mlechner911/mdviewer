import './style.css';
import { hydrate } from 'svelte';
import App from './App.svelte';

const target = document.getElementById('app');
if (target) {
  hydrate(App, { target });
}
