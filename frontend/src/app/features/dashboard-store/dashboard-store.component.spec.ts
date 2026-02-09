import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DashboardStoreComponent } from './dashboard-store.component';

describe('DashboardStoreComponent', () => {
  let component: DashboardStoreComponent;
  let fixture: ComponentFixture<DashboardStoreComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DashboardStoreComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(DashboardStoreComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
