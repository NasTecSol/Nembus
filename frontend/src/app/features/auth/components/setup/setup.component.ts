import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';

@Component({
  selector: 'app-setup',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './setup.component.html',
  // styleUrl: './setup.component.scss'
})
export class SetupComponent  {
    step = 1;
    statusMessage = '';
    tenants: any[] = [];
    selectedTenant: any = null;

    dbConfig = {
        username: 'postgres',
        password: 'password',
        database: 'nembus',
        port: 5432
    };

    constructor(private router: Router) { }

    nextStep() {
        this.step++;
        if (this.step === 4) {
            this.fetchTenants();
        }
    }

    prevStep() {
        this.step--;
    }

    async startSetup() {
        this.step = 3;
        this.statusMessage = 'Starting embedded database...';

        try {
            // @ts-ignore
            const result = await window.go.main.App.StartDatabase(
                this.dbConfig.username,
                this.dbConfig.password,
                this.dbConfig.database,
                this.dbConfig.port
            );

            if (result === 'Success') {
                this.statusMessage = 'Database ready. Fetching Cloud Tenants...';
                setTimeout(() => {
                    this.nextStep();
                    this.statusMessage = '';
                }, 1500);
            } else {
                this.statusMessage = 'Error: ' + result;
                setTimeout(() => {
                    this.step = 2;
                }, 3000);
            }
        } catch (err) {
            this.statusMessage = 'Error: ' + err;
            setTimeout(() => {
                this.step = 2;
            }, 3000);
        }
    }

    async fetchTenants() {
        this.statusMessage = 'Fetching tenants from cloud...';
        try {
            // @ts-ignore
            const result = await window.go.main.App.FetchCloudTenants();
            if (Array.isArray(result)) {
                this.tenants = result;
            } else {
                this.statusMessage = 'Error: ' + result;
            }
        } catch (err) {
            this.statusMessage = 'Error: ' + err;
        }
        this.statusMessage = '';
    }

    selectTenant(tenant: any) {
        this.selectedTenant = tenant;
    }

    async confirmClone() {
        if (!this.selectedTenant) return;

        this.step = 5; // Transition to cloning progress
        this.statusMessage = `Cloning tenant: ${this.selectedTenant.tenant_name}...`;

        try {
            // @ts-ignore
            const result = await window.go.main.App.CloneTenant(this.selectedTenant.slug);
            if (result === 'Success') {
                this.statusMessage = 'Cloning complete!';
                setTimeout(() => {
                    this.step = 7; // Final Success Screen
                    this.statusMessage = '';
                }, 2000);
            } else {
                this.statusMessage = 'Error: ' + result;
                setTimeout(() => {
                    this.step = 4;
                }, 3000);
            }
        } catch (err) {
            this.statusMessage = 'Error: ' + err;
            setTimeout(() => {
                this.step = 4;
            }, 3000);
        }
    }

    launchApp() {
        console.log('App launched');
        this.router.navigate(['/home']);
    }
}
