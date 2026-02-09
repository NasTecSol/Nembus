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
    try {
      // @ts-ignore
      const isSetup = await window.go.main.App.IsAppSetup();
      if (isSetup) {
        this.router.navigate(['/login']);
      } else {
        this.router.navigate(['/setup']);
      }
    } catch (err) {
      console.error('Error checking setup status:', err);
      this.router.navigate(['/setup']);
    }
  }
}
