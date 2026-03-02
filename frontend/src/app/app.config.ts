import { ApplicationConfig, provideZoneChangeDetection, importProvidersFrom, } from "@angular/core";
import { provideAnimations } from "@angular/platform-browser/animations";
import { provideToastr } from "ngx-toastr";
import { provideRouter } from "@angular/router";
import { provideHttpClient, HTTP_INTERCEPTORS, withInterceptorsFromDi } from "@angular/common/http";
import { routes } from "./app.routes";
import { TranslateModule, TranslateLoader } from "@ngx-translate/core";
import { HttpClient } from "@angular/common/http";
import { HttpLoaderFactory } from "./utils/translate-loader";
import { AuthInterceptor } from "./core/interceptors/auth.interceptor";
import { LoaderInterceptor } from "./core/interceptors/loader.interceptor";
import { TenantInterceptor } from "./core/interceptors/tenant.interceptor";

export const appConfig: ApplicationConfig = {
  providers: [
    provideZoneChangeDetection({ eventCoalescing: true }),
    provideAnimations(),
    provideToastr({
      timeOut: 3000,
      positionClass: "toast-bottom-right",
      preventDuplicates: true,
      closeButton: true,
    }),
    provideHttpClient(withInterceptorsFromDi()),
    provideRouter(routes),
    {
      provide: HTTP_INTERCEPTORS,
      useClass: LoaderInterceptor,
      multi: true,
    },
    {
      provide: HTTP_INTERCEPTORS,
      useClass: AuthInterceptor,
      multi: true,
    },
    {
      provide: HTTP_INTERCEPTORS,
      useClass: TenantInterceptor,
      multi: true,
    },
    importProvidersFrom(
      TranslateModule.forRoot({
        defaultLanguage: "en",
        loader: {
          provide: TranslateLoader,
          useFactory: HttpLoaderFactory,
          deps: [HttpClient],
        },
      })
    ),
  ],
};