// app.component.ts
import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterOutlet, Router } from '@angular/router';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, RouterOutlet],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss'
})
export class AppComponent implements OnInit {
  constructor(private router: Router) { }

  async ngOnInit() {
    console.log('AppComponent initialized');

    try {
      // @ts-ignore
      if (window.go && window.go.main && window.go.main.App) {
        // @ts-ignore
        const isSetup = await window.go.main.App.IsAppSetup();
        console.log('Is app setup?', isSetup);

        if (isSetup) {
          console.log('Navigating to login...');
          await this.router.navigate(['/auth/login']);
        } else {
          console.log('Navigating to setup...');
          await this.router.navigate(['/auth/setup']);
        }
      } else {
        console.warn('Wails not ready, navigating to setup');
        await this.router.navigate(['/auth/setup']);
      }
    } catch (err) {
      console.error('Error checking setup status:', err);
      await this.router.navigate(['/auth/setup']);
    }
  }
}