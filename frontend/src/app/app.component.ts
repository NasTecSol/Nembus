// app.component.ts
import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterOutlet, Router } from '@angular/router';
import { DeviceConfigService } from './core/services/device-config.service';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, RouterOutlet],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss'
})
export class AppComponent implements OnInit {
  constructor(
    private router: Router,
    private deviceConfig: DeviceConfigService
  ) { }

  async ngOnInit() {
    console.log('AppComponent initialized');

    try {
      // @ts-ignore
      if (window.go && window.go.main && window.go.main.App) {
        // ── Restore device config from disk before any navigation ──────────
        // This re-hydrates localStorage (store_id, terminal_id, cashier_id,
        // printer_settings, etc.) so the cashier does not have to redo the
        // setup wizard after every clean build or Wails restart.
        await this.deviceConfig.restore();

        // Running inside Wails desktop app — check if already configured
        // @ts-ignore
        const isSetup = await window.go.main.App.IsAppSetup();
        console.log('Is app setup?', isSetup);

        if (isSetup) {
          // Wait for the backend HTTP server to finish initialising before
          // navigating to login.  On first launch the setup wizard handles
          // this naturally; on subsequent launches the Gin server starts in a
          // goroutine and may not yet be ready when the frontend loads.
          await this.waitForBackend();
          console.log('Backend ready — navigating to login...');
          await this.router.navigate(['/auth/login']);
        } else {
          console.log('Navigating to setup...');
          await this.router.navigate(['/auth/setup']);
        }
      }
      // Running in browser — do nothing here.
      // Angular router + AuthGuard handle all navigation automatically.
    } catch (err) {
      console.error('Error checking setup status:', err);
      // Do NOT redirect to auth/setup on error in browser mode.
    }
  }

  /**
   * Polls IsBackendReady() every 500 ms until it returns true or the timeout
   * (10 seconds) is reached.  This prevents API calls being made before the
   * embedded PostgreSQL and Gin server have fully initialised.
   */
  private waitForBackend(timeoutMs = 10000): Promise<void> {
    return new Promise((resolve) => {
      const interval = 500;
      let elapsed = 0;

      const check = async () => {
        try {
          // @ts-ignore
          const ready = await window.go.main.App.IsBackendReady();
          if (ready) {
            resolve();
            return;
          }
        } catch {
          // Ignore errors during startup polling
        }

        elapsed += interval;
        if (elapsed >= timeoutMs) {
          console.warn('Backend readiness timeout reached — proceeding anyway');
          resolve();
          return;
        }

        setTimeout(check, interval);
      };

      check();
    });
  }
}
