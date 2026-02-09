// starter.component.ts
import { CommonModule } from "@angular/common";
import { Component, ViewChild, OnInit } from "@angular/core";
import { TranslateModule } from "@ngx-translate/core";
import {
  NgApexchartsModule,
  ChartComponent,
  ApexAxisChartSeries,
  ApexXAxis,
  ApexDataLabels,
  ApexTitleSubtitle,
  ApexStroke,
  ApexGrid,
} from "ng-apexcharts";
import {
  ApexNonAxisChartSeries,
  ApexResponsive,
  ApexChart,
  ApexLegend,
} from "ng-apexcharts";

export interface DonutChartOptions {
  series: ApexNonAxisChartSeries;
  chart: ApexChart;
  labels: string[];
  colors: string[];
  responsive: ApexResponsive[];
  legend?: ApexLegend;
}
export type LineChartOptions = {
  series: ApexAxisChartSeries;
  chart: ApexChart;
  xaxis: ApexXAxis;
  dataLabels: ApexDataLabels;
  grid: ApexGrid;
  stroke: ApexStroke;
  title: ApexTitleSubtitle;
};

@Component({
  selector: "app-home",
  imports: [CommonModule,NgApexchartsModule,TranslateModule],
  templateUrl: "./home.component.html",
})
export class HomeComponent {
  @ViewChild("chart", { static: true }) chart!: ChartComponent;

  // ✅ exact type — no “maybe-undefined” properties
  public donutChartOptioins: DonutChartOptions = {
    series: [13, 43, 22],
    chart: { type: "donut" },
    labels: ["Branch 1", "Branch 2", "Branch 3"],
    colors: ["#F79009", "#335C67", "#06AED4"],

    responsive: [
      {
        breakpoint: 480,
        options: {
          chart: { width: 200 },
          legend: { position: "top" },
        },
      },
    ],
  };

  public lineChartOptions: LineChartOptions = {
    series: [
      {
        name: "Desktops",
        data: [10, 41, 35, 51, 49, 62, 69, 91, 148],
      },
    ],
    chart: {
      height: 350,
      type: "line",
      zoom: {
        enabled: false,
      },
    },
    dataLabels: {
      enabled: false,
    },
    stroke: {
      curve: "straight",
    },
    title: {
      text: "Product Trends by Month",
      align: "left",
    },
    grid: {
      row: {
        colors: ["#f3f3f3", "transparent"], // takes an array which will be repeated on columns
        opacity: 0.5,
      },
    },
    xaxis: {
      categories: [
        "Jan",
        "Feb",
        "Mar",
        "Apr",
        "May",
        "Jun",
        "Jul",
        "Aug",
        "Sep",
      ],
    },
  };
   public topItems = [
    { code: 'A001', desc: 'Olive Oil' },
    { code: 'A002', desc: 'Rice 10kg' },
    { code: 'A003', desc: 'Sugar 5kg' },
    { code: 'A004', desc: 'Wheat Flour' },
    { code: 'A005', desc: 'Milk Pack' }
  ];

  public stockItems = [
    { name: 'Tomato', qty: '2 kg' },
    { name: 'Potato', qty: '5 kg' },
    { name: 'Cooking Oil', qty: '1 L' },
    { name: 'Salt', qty: 'Out of stock' },
    { name: 'Onion', qty: '3 kg' }
  ];

  ngOnInit(): void {
    // run-time updates (optional)
  }
}
