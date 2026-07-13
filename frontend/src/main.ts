import { bootstrapApplication } from '@angular/platform-browser';
import { appConfig } from './app/app.config';
import { AppComponent } from './app/app.component';
import { environment } from './environments/environment';

const root = document.documentElement;
root.classList.remove('pos-theme', 'bofc-theme');
root.classList.add(environment.pos ? 'pos-theme' : 'bofc-theme');
root.setAttribute('data-app-mode', environment.pos ? 'pos' : 'bofc');

bootstrapApplication(AppComponent, appConfig)
  .catch((err) => console.error(err));
